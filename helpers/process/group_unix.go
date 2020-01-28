// +build darwin dragonfly freebsd linux netbsd openbsd

package process

import (
	"os/exec"
	"syscall"
)

func setProcessGroup(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func (c *osCmd) IsProcessGroup() bool {
	if c.internal.Process == nil {
		return false
	}

	attr := c.internal.SysProcAttr

	return attr != nil && attr.Setpgid
}
