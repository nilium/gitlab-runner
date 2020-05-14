package network

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"github.com/sirupsen/logrus"

	"gitlab.com/gitlab-org/gitlab-runner/common"
	"gitlab.com/gitlab-org/gitlab-runner/helpers/tls/ca_chain"
	"gitlab.com/gitlab-org/gitlab-runner/network/internal/response"
)

const jsonMimeType = "application/json"

type requestCredentials interface {
	GetURL() string
	GetToken() string
	GetTLSCAFile() string
	GetTLSCertFile() string
	GetTLSKeyFile() string
}

var (
	dialer = net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	backOffDelayMin    = 100 * time.Millisecond
	backOffDelayMax    = 60 * time.Second
	backOffDelayFactor = 2.0
	backOffDelayJitter = true
)

type client struct {
	http.Client
	url             *url.URL
	caFile          string
	certFile        string
	keyFile         string
	caData          []byte
	skipVerify      bool
	updateTime      time.Time
	lastUpdate      string
	requestBackOffs map[string]*backoff.Backoff
	lock            sync.Mutex

	requester requester
}

type ResponseTLSData struct {
	CAChain  string
	CertFile string
	KeyFile  string
}

func (n *client) getLastUpdate() string {
	return n.lastUpdate
}

func (n *client) setLastUpdate(headers http.Header) {
	if lu := headers.Get("X-GitLab-Last-Update"); len(lu) > 0 {
		n.lastUpdate = lu
	}
}

func (n *client) ensureTLSConfig() {
	// certificate got modified
	if stat, err := os.Stat(n.caFile); err == nil && n.updateTime.Before(stat.ModTime()) {
		n.Transport = nil
	}

	// client certificate got modified
	if stat, err := os.Stat(n.certFile); err == nil && n.updateTime.Before(stat.ModTime()) {
		n.Transport = nil
	}

	// client private key got modified
	if stat, err := os.Stat(n.keyFile); err == nil && n.updateTime.Before(stat.ModTime()) {
		n.Transport = nil
	}

	// create or update transport
	if n.Transport == nil {
		n.updateTime = time.Now()
		n.createTransport()
	}
}

func (n *client) addTLSCA(tlsConfig *tls.Config) {
	// load TLS CA certificate
	if file := n.caFile; file != "" && !n.skipVerify {
		logrus.Debugln("Trying to load", file, "...")

		data, err := ioutil.ReadFile(file)
		if err == nil {
			pool, err := x509.SystemCertPool()
			if err != nil {
				logrus.Warningln("Failed to load system CertPool:", err)
			}
			if pool == nil {
				pool = x509.NewCertPool()
			}
			if pool.AppendCertsFromPEM(data) {
				tlsConfig.RootCAs = pool
				n.caData = data
			} else {
				logrus.Errorln("Failed to parse PEM in", n.caFile)
			}
		} else {
			if !os.IsNotExist(err) {
				logrus.Errorln("Failed to load", n.caFile, err)
			}
		}
	}
}

func (n *client) addTLSAuth(tlsConfig *tls.Config) {
	// load TLS client keypair
	if cert, key := n.certFile, n.keyFile; cert != "" && key != "" {
		logrus.Debugln("Trying to load", cert, "and", key, "pair...")

		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{certificate}
			tlsConfig.BuildNameToCertificate()
		} else {
			if !os.IsNotExist(err) {
				logrus.Errorln("Failed to load", cert, key, err)
			}
		}
	}
}

func (n *client) createTransport() {
	// create reference TLS config
	tlsConfig := tls.Config{
		MinVersion:         tls.VersionTLS10,
		InsecureSkipVerify: n.skipVerify,
	}

	n.addTLSCA(&tlsConfig)
	n.addTLSAuth(&tlsConfig)

	// create transport
	n.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: func(network, addr string) (net.Conn, error) {
			logrus.Debugln("Dialing:", network, addr, "...")
			return dialer.Dial(network, addr)
		},
		TLSClientConfig:       &tlsConfig,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Minute,
	}
	n.Timeout = common.DefaultNetworkClientTimeout
}

func (n *client) ensureBackoff(method, uri string) *backoff.Backoff {
	n.lock.Lock()
	defer n.lock.Unlock()

	key := fmt.Sprintf("%s_%s", method, uri)
	if n.requestBackOffs[key] == nil {
		n.requestBackOffs[key] = &backoff.Backoff{
			Min:    backOffDelayMin,
			Max:    backOffDelayMax,
			Factor: backOffDelayFactor,
			Jitter: backOffDelayJitter,
		}
	}

	return n.requestBackOffs[key]
}

func (n *client) backoffRequired(res *http.Response) bool {
	return res.StatusCode >= 400 && res.StatusCode < 600
}

func (n *client) checkBackoffRequest(req *http.Request, res *http.Response) {
	backoffDelay := n.ensureBackoff(req.Method, req.RequestURI)
	if n.backoffRequired(res) {
		time.Sleep(backoffDelay.Duration())
	} else {
		backoffDelay.Reset()
	}
}

func (n *client) do(uri, method string, request io.Reader, requestType string, headers http.Header) (*response.Response, error) {
	url, err := n.url.Parse(uri)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url.String(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to create NewRequest: %w", err)
	}

	if headers != nil {
		req.Header = headers
	}

	if request != nil {
		req.Header.Set("Content-Type", requestType)
		req.Header.Set("User-Agent", common.AppVersion.UserAgent())
	}

	n.ensureTLSConfig()

	httpResponse, err := n.requester.Do(req)
	if err != nil {
		return nil, err
	}

	n.checkBackoffRequest(req, httpResponse)

	return response.New(httpResponse), nil
}

func (n *client) doJSON(uri, method string, expectedStatusCode int, request interface{}, result interface{}) *response.Response {
	var body io.Reader

	if request != nil {
		requestBody, err := json.Marshal(request)
		if err != nil {
			return response.NewError(fmt.Errorf("failed to marshal project object: %w", err))
		}
		body = bytes.NewReader(requestBody)
	}

	headers := make(http.Header)

	if result != nil {
		headers.Set("Accept", jsonMimeType)
	}

	httpResponse, err := n.do(uri, method, body, jsonMimeType, headers)
	if err != nil {
		return response.NewError(err)
	}

	if httpResponse.StatusCode() == expectedStatusCode {
		if result != nil {
			err := httpResponse.IsApplicationJSON()
			if err != nil {
				return response.NewError(err)
			}

			err = httpResponse.DecodeJSONFromBody(result)
			if err != nil {
				return response.NewError(fmt.Errorf("decoding payload: %w", err))
			}
		}
	}

	n.setLastUpdate(httpResponse.Header())

	return httpResponse
}

func (n *client) getResponseTLSData(TLS *tls.ConnectionState) (ResponseTLSData, error) {
	TLSData := ResponseTLSData{
		CertFile: n.certFile,
		KeyFile:  n.keyFile,
	}

	caChain, err := n.buildCAChain(TLS)
	if err != nil {
		return TLSData, fmt.Errorf("couldn't build CA Chain: %w", err)
	}

	TLSData.CAChain = caChain

	return TLSData, nil
}

func (n *client) buildCAChain(tls *tls.ConnectionState) (string, error) {
	if len(n.caData) != 0 {
		return string(n.caData), nil
	}

	if tls == nil {
		return "", nil
	}

	builder := ca_chain.NewBuilder(logrus.StandardLogger())
	err := builder.BuildChainFromTLSConnectionState(tls)
	if err != nil {
		return "", fmt.Errorf("error while fetching certificates from TLS ConnectionState: %w", err)
	}

	return builder.String(), nil
}

func fixCIURL(url string) string {
	url = strings.TrimRight(url, "/")
	if strings.HasSuffix(url, "/ci") {
		url = strings.TrimSuffix(url, "/ci")
	}
	return url
}

func (n *client) findCertificate(certificate *string, base string, name string) {
	if *certificate != "" {
		return
	}
	path := filepath.Join(base, name)
	if _, err := os.Stat(path); err == nil {
		*certificate = path
	}
}

func newClient(requestCredentials requestCredentials) (c *client, err error) {
	url, err := url.Parse(fixCIURL(requestCredentials.GetURL()) + "/api/v4/")
	if err != nil {
		return
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		err = errors.New("only http or https scheme supported")
		return
	}

	c = &client{
		url:             url,
		caFile:          requestCredentials.GetTLSCAFile(),
		certFile:        requestCredentials.GetTLSCertFile(),
		keyFile:         requestCredentials.GetTLSKeyFile(),
		requestBackOffs: make(map[string]*backoff.Backoff),
	}
	c.requester = newRateLimitRequester(&c.Client)

	host := strings.Split(url.Host, ":")[0]
	if CertificateDirectory != "" {
		c.findCertificate(&c.caFile, CertificateDirectory, host+".crt")
		c.findCertificate(&c.certFile, CertificateDirectory, host+".auth.crt")
		c.findCertificate(&c.keyFile, CertificateDirectory, host+".auth.key")
	}

	return
}
