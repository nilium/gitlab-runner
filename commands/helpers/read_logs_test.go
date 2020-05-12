package helpers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewReadLogsCommandFileNotExist(t *testing.T) {
	cmd := newReadLogsCommand()
	cmd.logStreamProvider = &fileLogStreamProvider{
		cmd:                    cmd,
		waitFileToExistTimeout: 2 * time.Second,
	}
	cmd.Path = "not_exists"

	err := cmd.readLogs()

	assert.Equal(t, errWaitingFileTimeout, err)
}

func TestNewReadLogsCommandFileSeekToInvalidLocation(t *testing.T) {
	testFile, cleanup := setupTestFile(t)
	defer cleanup()

	cmd := newReadLogsCommand()
	cmd.Path = testFile.Name()
	cmd.Offset = -1

	err := cmd.readLogs()

	var expectedErr *os.PathError
	assert.True(t, errors.As(err, &expectedErr), "expected err %T, but got %T", expectedErr, err)
}

func setupTestFile(t *testing.T) (*os.File, func()) {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}

	return f, cleanup
}

func TestNewReadLogsCommandLines(t *testing.T) {
	lines := []string{"1", "2", "3"}
	f, cleanup := setupTestFile(t)
	defer cleanup()
	appendToFile(t, f, strings.Join(lines, "\n"))

	cmd := newReadLogsCommand()

	mockLogOutputWriter := new(mockLogOutputWriter)
	defer mockLogOutputWriter.AssertExpectations(t)
	wg := setupMockLogOutputWriterFromLines(mockLogOutputWriter, lines)
	cmd.logOutputWriter = mockLogOutputWriter

	mockLogStreamProvider := new(mockLogStreamProvider)
	defer mockLogStreamProvider.AssertExpectations(t)
	mockLogStreamProvider.On("Open").Return(f, nil)
	cmd.logStreamProvider = mockLogStreamProvider

	go func() {
		wg.Wait()
		_ = f.Close()
	}()

	err := cmd.readLogs()
	var expectedErr *os.PathError
	assert.True(t, errors.As(err, &expectedErr), "expected err %T, but got %T", expectedErr, err)
}

func appendToFile(t *testing.T, f *os.File, data string) {
	fw, err := os.OpenFile(f.Name(), os.O_WRONLY, 0600)
	require.NoError(t, err)
	_, err = fw.Write([]byte(data))
	require.NoError(t, err)
	_ = fw.Close()
}

func setupMockLogOutputWriterFromLines(lw *mockLogOutputWriter, lines []string) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(len(lines))

	var offset int
	for i, l := range lines {
		offset += len(l)
		if i < len(lines)-1 {
			offset++ // account for the len of the newline character \n
		}

		lw.On("Write", fmt.Sprintf("%d %s\n", offset, l)).Run(func(mock.Arguments) {
			wg.Done()
		})
	}

	return &wg
}

func TestNewReadLogsCommandWriteLinesWithDelay(t *testing.T) {
	lines1 := []string{"1", "2", "3"}
	lines2 := []string{"4", "5", "6"}

	f, cleanup := setupTestFile(t)
	defer cleanup()
	appendToFile(t, f, strings.Join(lines1, "\n"))

	cmd := newReadLogsCommand()

	mockLogOutputWriter := new(mockLogOutputWriter)
	defer mockLogOutputWriter.AssertExpectations(t)
	wg := setupMockLogOutputWriterFromLines(mockLogOutputWriter, lines1)
	cmd.logOutputWriter = mockLogOutputWriter

	mockLogStreamProvider := new(mockLogStreamProvider)
	defer mockLogStreamProvider.AssertExpectations(t)
	openCall := mockLogStreamProvider.On("Open").Return(f, nil)
	cmd.logStreamProvider = mockLogStreamProvider

	go func() {
		wg.Wait()

		<-time.After(5 * time.Second)
		wg = setupMockLogOutputWriterFromLines(mockLogOutputWriter, lines2)
		appendToFile(t, f, strings.Join(lines2, "\n"))
		openCall.Return(f, nil)

		wg.Wait()
	}()

	err := cmd.readLogs()
	var expectedErr *os.PathError
	assert.True(t, errors.As(err, &expectedErr), "expected err %T, but got %T", expectedErr, err)
}
