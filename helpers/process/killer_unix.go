// +build darwin dragonfly freebsd linux netbsd openbsd

package process

import (
	"syscall"
)

type unixKiller struct {
	logger Logger
	cmd    Commander
}

func newKiller(logger Logger, cmd Commander) killer {
	return &unixKiller{
		logger: logger,
		cmd:    cmd,
	}
}

func (pk *unixKiller) Terminate() {
	if pk.cmd.Process() == nil {
		return
	}

	err := syscall.Kill(pk.getPID(), syscall.SIGTERM)
	if err != nil {
		pk.logger.Warn("Failed to terminate process:", err)

		// try to kill right-after
		pk.ForceKill()
	}
}

func (pk *unixKiller) ForceKill() {
	if pk.cmd.Process() == nil {
		return
	}

	err := syscall.Kill(pk.getPID(), syscall.SIGKILL)
	if err != nil {
		pk.logger.Warn("Failed to force-kill:", err)
	}
}

// getPID will check if the process is a process group or not. If it's a process
// group return the negative PID (-PID) otherwise send the normal PID.
//
// The negative symbol comes from kill(2) https://linux.die.net/man/2/kill `If
// pid is less than -1, then sig is sent to every process in the process group
// whose ID is -pid.`
func (pk *unixKiller) getPID() int {
	pid := pk.cmd.Process().Pid

	if pk.cmd.IsProcessGroup() {
		pid *= -1
	}

	return pid
}
