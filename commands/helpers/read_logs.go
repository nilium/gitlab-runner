package helpers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli"

	"gitlab.com/gitlab-org/gitlab-runner/common"
)

const (
	defaultWaitFileToExistTimeout = time.Minute
	defaultReaderBufferSize       = 16 * 1024
)

var errWaitingFileTimeout = errors.New("timeout waiting for file to be created")

type logStreamProvider interface {
	Open() (logStream, error)
}

type logStream interface {
	io.ReadSeeker
	io.Closer
}

type fileLogStreamProvider struct {
	cmd                    *ReadLogsCommand
	waitFileToExistTimeout time.Duration
}

func (p *fileLogStreamProvider) Open() (logStream, error) {
	timeoutChan := time.After(p.waitFileToExistTimeout)

	for {
		select {
		case <-timeoutChan:
			return nil, errWaitingFileTimeout
		default:
		}

		f, err := os.Open(p.cmd.Path)
		if os.IsNotExist(err) {
			time.Sleep(time.Second)
			continue
		}

		return f, err
	}
}

type logOutputWriter interface {
	Write(string)
}

type streamLogOutputWriter struct {
	stream io.Writer
}

func (s *streamLogOutputWriter) Write(data string) {
	_, _ = fmt.Fprint(s.stream, data)
}

type ReadLogsCommand struct {
	Path   string `long:"path"`
	Offset int64  `long:"offset"`

	logStreamProvider logStreamProvider
	logOutputWriter   logOutputWriter
	readerBufferSize  int
}

func newReadLogsCommand() *ReadLogsCommand {
	cmd := new(ReadLogsCommand)
	cmd.logStreamProvider = &fileLogStreamProvider{
		cmd:                    cmd,
		waitFileToExistTimeout: defaultWaitFileToExistTimeout,
	}
	cmd.logOutputWriter = &streamLogOutputWriter{stream: os.Stdout}
	cmd.readerBufferSize = defaultReaderBufferSize

	return cmd
}

func (c *ReadLogsCommand) Execute(*cli.Context) {
	if err := c.readLogs(); err != nil {
		c.logOutputWriter.Write(fmt.Sprintf("error reading logs %v\n", err))
	}
}

func (c *ReadLogsCommand) readLogs() error {
	s, r, err := c.openFileReader(c.Offset)
	if err != nil {
		return err
	}

	defer func() {
		if s != nil {
			_ = s.Close()
		}
	}()

	offset := c.Offset
	for {
		buf, err := r.ReadSlice('\n')
		if len(buf) > 0 {
			offset += int64(len(buf))
			if !bytes.HasSuffix(buf, []byte{'\n'}) {
				buf = append(buf, '\n')
			}

			c.logOutputWriter.Write(fmt.Sprintf("%d %s", offset, buf))
		}

		if err == io.EOF {
			fmt.Println("sleep")
			_, _ = s.Seek(offset, 0)
			time.Sleep(time.Second)
		} else if err != nil && err != bufio.ErrBufferFull {
			return err
		}
	}
}

func (c *ReadLogsCommand) openFileReader(offset int64) (logStream, *bufio.Reader, error) {
	s, err := c.logStreamProvider.Open()
	if err != nil {
		return nil, nil, err
	}

	_, err = s.Seek(offset, 0)
	if err != nil {
		_ = s.Close()
		return nil, nil, err
	}

	return s, bufio.NewReaderSize(s, c.readerBufferSize), nil
}

func init() {
	common.RegisterCommand2("read-logs", "reads logs from a file", newReadLogsCommand())
}
