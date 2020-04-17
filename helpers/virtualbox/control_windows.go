package virtualbox

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func init() {
	_ = addDirectoryToPATH(os.Getenv("ProgramFiles"))
	_ = addDirectoryToPATH(os.Getenv("ProgramW6432"))
}

func addDirectoryToPATH(programFilesPath string) error {
	if programFilesPath == "" {
		return errors.New("invalid ProgramFiles path")
	}

	virtualBoxPath := filepath.Join(programFilesPath, "Oracle", "VirtualBox")
	newPath := fmt.Sprintf("%s;%s", os.Getenv("PATH"), virtualBoxPath)
	err := os.Setenv("PATH", newPath)
	if err != nil {
		logrus.Warnf(
			"Failed to add path to VBoxManage.exe (%q) to end of local PATH: %s",
			virtualBoxPath,
			err)
		return err
	}

	logrus.Debugf("Added path to VBoxManage.exe to end of local PATH: %q", virtualBoxPath)
	return nil
}
