package process

import (
	"os"
)

type windowsKiller struct {
	logger  Logger
	process *os.Process
}

func newKiller(logger Logger, cmd Commander) killer {
	return &windowsKiller{
		logger:  logger,
		process: cmd.Process(),
	}
}

func (pk *windowsKiller) Terminate() {
	if pk.process == nil {
		return
	}

	err := pk.process.Kill()
	if err != nil {
		pk.logger.Warn("Failed to terminate process:", err)

		// try to kill right-after
		pk.ForceKill()
	}
}

func (pk *windowsKiller) ForceKill() {
	if pk.process == nil {
		return
	}

	err := pk.process.Kill()
	if err != nil {
		pk.logger.Warn("Failed to force-kill:", err)
	}
}
