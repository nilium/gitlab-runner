package common

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/gitlab-org/gitlab-runner/helpers"
	"gitlab.com/gitlab-org/gitlab-runner/helpers/url"
)

func (e *BuildLogger) getCurrentFormattedTime() string {
	dateTimeFormat := "2006-01-02 15:04 MST"
	datetime, err := time.Parse(dateTimeFormat, time.Now().UTC().String())
	if err != nil {
		return time.Now().UTC().String()
	}
	return datetime.String()
}

type BuildLogger struct {
	log   JobTrace
	entry *logrus.Entry
}

func (e *BuildLogger) WithFields(fields logrus.Fields) BuildLogger {
	return NewBuildLogger(e.log, e.entry.WithFields(fields))
}

func (e *BuildLogger) SendRawLog(args ...interface{}) {
	if e.log != nil {
		fmt.Fprint(e.log, args...)
	}
}

func (e *BuildLogger) sendLog(logger func(args ...interface{}), logPrefix string, args ...interface{}) {
	if e.log != nil {
		logLine := url_helpers.ScrubSecrets(logPrefix + fmt.Sprintln(args...))
		e.SendRawLog(helpers.ANSI_BOLD_CYAN + e.getCurrentFormattedTime() + helpers.ANSI_RESET + "\n")
		e.SendRawLog(logLine)
		e.SendRawLog(helpers.ANSI_RESET)

		if e.log.IsStdout() {
			return
		}
	}

	if len(args) == 0 {
		return
	}

	logger(args...)
}

func (e *BuildLogger) Debugln(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.entry.Debugln(args...)
}

func (e *BuildLogger) Println(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.sendLog(e.entry.Debugln, helpers.ANSI_CLEAR, args...)
}

func (e *BuildLogger) Infoln(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.sendLog(e.entry.Println, helpers.ANSI_BOLD_GREEN, args...)
}

func (e *BuildLogger) Warningln(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.sendLog(e.entry.Warningln, helpers.ANSI_BOLD_CYAN+"WARNING: ", args...)
}

func (e *BuildLogger) SoftErrorln(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.sendLog(e.entry.Warningln, helpers.ANSI_BOLD_RED+"ERROR: ", args...)
}

func (e *BuildLogger) Errorln(args ...interface{}) {
	if e.entry == nil {
		return
	}
	e.sendLog(e.entry.Errorln, helpers.ANSI_BOLD_RED+"ERROR: ", args...)
}

func NewBuildLogger(log JobTrace, entry *logrus.Entry) BuildLogger {
	return BuildLogger{
		log:   log,
		entry: entry,
	}
}
