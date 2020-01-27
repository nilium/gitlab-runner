// +build darwin dragonfly freebsd linux netbsd openbsd

package process

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetProcessGroup(t *testing.T) {
	cmd := exec.Command("sleep", "1")
	require.Nil(t, cmd.SysProcAttr)
	setProcessGroup(cmd)
	assert.True(t, cmd.SysProcAttr.Setpgid)
}

func Test_cmd_IsProcessGroup(t *testing.T) {
	tests := map[string]struct {
		cmd      exec.Cmd
		expected bool
	}{
		"no process": {
			cmd: exec.Cmd{
				Process: nil,
			},
			expected: false,
		},
		"no sys proc attr": {
			cmd: exec.Cmd{
				Process: &os.Process{
					Pid: 1,
				},
				SysProcAttr: nil,
			},
			expected: false,
		},
		"pgid is false": {
			cmd: exec.Cmd{
				Process: &os.Process{
					Pid: 1,
				},
				SysProcAttr: &syscall.SysProcAttr{
					Setpgid: false,
				},
			},
			expected: false,
		},
		"pgid is true": {
			cmd: exec.Cmd{
				Process: &os.Process{
					Pid: 1,
				},
				SysProcAttr: &syscall.SysProcAttr{
					Setpgid: true,
				},
			},
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := osCmd{internal: &tt.cmd}
			assert.Equal(t, tt.expected, c.IsProcessGroup())
		})
	}
}
