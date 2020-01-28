package process

import "os/exec"

func setProcessGroup(c *exec.Cmd) {
	// noop process groups not supported on Windows.
}

func (c *osCmd) IsProcessGroup() bool {
	// process groups not supported on Windows.
	return false
}
