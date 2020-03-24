package machine

import (
	"errors"

	"gitlab.com/gitlab-org/gitlab-runner/session/terminal"
	terminalsession "gitlab.com/gitlab-org/gitlab-runner/session/terminal"
)

func (e *machineExecutor) Connect() (terminalsession.Conn, error) {
	term, ok := e.executor.(terminal.InteractiveTerminal)
	if !ok {
		return nil, errors.New("executor does not have terminal")
	}

	return term.Connect()
}
