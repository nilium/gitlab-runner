package process

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type killer interface {
	Terminate()
	ForceKill()
}

var newKillerFactory = newKiller

type Logger interface {
	WithFields(fields logrus.Fields) Logger
	Errorln(args ...interface{})
}

type KillWaiter interface {
	KillAndWait(command Commander, waitCh chan error) error
}

type DefaultKillWaiter struct {
	logger Logger

	gracefulKillTimeout time.Duration
	forceKillTimeout    time.Duration
}

func NewKillWaiter(logger Logger, gracefulKillTimeout time.Duration, forceKillTimeout time.Duration) KillWaiter {
	return &DefaultKillWaiter{
		logger:              logger,
		gracefulKillTimeout: gracefulKillTimeout,
		forceKillTimeout:    forceKillTimeout,
	}
}

func (kw *DefaultKillWaiter) KillAndWait(command Commander, waitCh chan error) error {
	process := command.Process()

	if process == nil {
		return errors.New("process not started yet")
	}

	log := kw.logger.WithFields(logrus.Fields{
		"PID": process.Pid,
	})

	processKiller := newKillerFactory(log, command)
	processKiller.Terminate()

	select {
	case err := <-waitCh:
		return err

	case <-time.After(kw.gracefulKillTimeout):
		processKiller.ForceKill()

		select {
		case err := <-waitCh:
			return err

		case <-time.After(kw.forceKillTimeout):
			return dormantProcessError(process)
		}
	}
}

func dormantProcessError(process *os.Process) error {
	return fmt.Errorf("failed to kill process PID=%d, likely process is dormant", process.Pid)
}
