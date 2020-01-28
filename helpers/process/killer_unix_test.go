// +build darwin dragonfly freebsd linux netbsd openbsd

package process

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Cases for UNIX systems that are used in `filler.go#TestKiller`.
func testKillerTestCases() map[string]testKillerTestCase {
	return map[string]testKillerTestCase{
		"command terminated": {
			alreadyStopped: false,
			skipTerminate:  true,
			expectedError:  "",
		},
		"command not terminated": {
			alreadyStopped: false,
			skipTerminate:  false,
			expectedError:  "exit status 1",
		},
		"command already stopped": {
			alreadyStopped: true,
			expectedError:  "signal: killed",
		},
	}
}

func Test_unixKiller_getPID(t *testing.T) {
	tests := []struct {
		processGroup bool
		expectedPID  int
	}{
		{
			processGroup: true,
			expectedPID:  -1,
		},
		{
			processGroup: false,
			expectedPID:  1,
		},
	}

	for _, tt := range tests {
		t.Run("processGroup_"+strconv.FormatBool(tt.processGroup), func(t *testing.T) {
			mCmd := new(MockCommander)
			defer mCmd.AssertExpectations(t)
			mLogger := new(MockLogger)
			defer mLogger.AssertExpectations(t)

			killer := unixKiller{logger: mLogger, cmd: mCmd}

			mCmd.On("IsProcessGroup").Return(tt.processGroup).Once()
			mCmd.On("Process").Return(&os.Process{Pid: 1}).Once()

			pid := killer.getPID()
			assert.Equal(t, tt.expectedPID, pid)
		})
	}
}
