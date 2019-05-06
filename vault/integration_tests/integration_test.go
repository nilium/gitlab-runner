package integration_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/gitlab-runner/vault"
	"gitlab.com/gitlab-org/gitlab-runner/vault/config"
)

func TestTokenLogin(t *testing.T) {
	s := newService(t)

	conf := config.Vault{
		Server: s.getVaultServerConfig(serviceProxyPort),
		Auth: config.VaultAuth{
			Token: s.getVaultTokenAuthConfig(),
		},
	}

	v := vault.New()

	err := v.Connect(conf.Server)
	assert.NoError(t, err)

	err = v.Authenticate(conf.Auth)
	assert.NoError(t, err)
}

func TestUserpassLogin(t *testing.T) {
	s := newService(t)

	conf := config.Vault{
		Server: s.getVaultServerConfig(serviceProxyPort),
		Auth: config.VaultAuth{
			Userpass: s.getVaultUserpassAuthConfig(),
		},
	}

	v := vault.New()

	err := v.Connect(conf.Server)
	assert.NoError(t, err)

	err = v.Authenticate(conf.Auth)
	assert.NoError(t, err)
}

func TestTLSLogin(t *testing.T) {
	s := newService(t)

	conf := config.Vault{
		Server: s.getVaultServerConfig(serviceDirectPort),
		Auth: config.VaultAuth{
			TLS: s.getVaultTLSAuthConfig(),
		},
	}

	v := vault.New()

	err := v.Connect(conf.Server)
	assert.NoError(t, err)

	err = v.Authenticate(conf.Auth)
	assert.NoError(t, err)
}
