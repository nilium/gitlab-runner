package virtualbox_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-runner/common"
	"gitlab.com/gitlab-org/gitlab-runner/helpers"
	"gitlab.com/gitlab-org/gitlab-runner/helpers/ssh"
)

const vboxImage = "ubuntu-runner"
const vboxManage = "vboxmanage"

var vboxSSHConfig = &ssh.Config{
	User:     "vagrant",
	Password: "vagrant",
}

func TestVirtualBoxExecutorRegistered(t *testing.T) {
	executorNames := common.GetExecutorNames()
	assert.Contains(t, executorNames, "virtualbox")
}

func TestVirtualBoxCreateExecutor(t *testing.T) {
	executor := common.NewExecutor("virtualbox")
	assert.NotNil(t, executor)
}

func TestVirtualBoxSuccessRun(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	successfulBuild, err := common.GetRemoteSuccessfulBuild()
	assert.NoError(t, err)
	build := &common.Build{
		JobResponse: successfulBuild,
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         vboxImage,
					DisableSnapshots: true,
				},
				SSH: vboxSSHConfig,
			},
		},
	}

	err = build.Run(&common.Config{}, &common.Trace{Writer: os.Stdout})
	assert.NoError(t, err, "Make sure that you have done 'make -C tests/ubuntu virtualbox'")
}

func TestVirtualBoxBuildFail(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	failedBuild, err := common.GetRemoteFailedBuild()
	assert.NoError(t, err)
	build := &common.Build{
		JobResponse: failedBuild,
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         vboxImage,
					DisableSnapshots: true,
				},
				SSH: vboxSSHConfig,
			},
		},
	}

	err = build.Run(&common.Config{}, &common.Trace{Writer: os.Stdout})
	require.Error(t, err, "error")
	assert.IsType(t, err, &common.BuildError{})
	assert.Contains(t, err.Error(), "Process exited with: 1")
}

func TestVirtualBoxMissingImage(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	build := &common.Build{
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         "non-existing-image",
					DisableSnapshots: true,
				},
				SSH: vboxSSHConfig,
			},
		},
	}

	err := build.Run(&common.Config{}, &common.Trace{Writer: os.Stdout})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Could not find a registered machine named")
}

func TestVirtualBoxMissingSSHCredentials(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	build := &common.Build{
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         "non-existing-image",
					DisableSnapshots: true,
				},
			},
		},
	}

	err := build.Run(&common.Config{}, &common.Trace{Writer: os.Stdout})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing SSH config")
}

func TestVirtualBoxBuildAbort(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	longRunningBuild, err := common.GetRemoteLongRunningBuild()
	assert.NoError(t, err)
	build := &common.Build{
		JobResponse: longRunningBuild,
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         vboxImage,
					DisableSnapshots: true,
				},
				SSH: vboxSSHConfig,
			},
		},
		SystemInterrupt: make(chan os.Signal, 1),
	}

	abortTimer := time.AfterFunc(time.Second, func() {
		t.Log("Interrupt")
		build.SystemInterrupt <- os.Interrupt
	})
	defer abortTimer.Stop()

	timeoutTimer := time.AfterFunc(time.Minute, func() {
		t.Log("Timedout")
		t.FailNow()
	})
	defer timeoutTimer.Stop()

	err = build.Run(&common.Config{}, &common.Trace{Writer: os.Stdout})
	assert.EqualError(t, err, "aborted: interrupt")
}

func TestVirtualBoxBuildCancel(t *testing.T) {
	if helpers.SkipIntegrationTests(t, vboxManage, "--version") {
		return
	}

	longRunningBuild, err := common.GetRemoteLongRunningBuild()
	assert.NoError(t, err)
	build := &common.Build{
		JobResponse: longRunningBuild,
		Runner: &common.RunnerConfig{
			RunnerSettings: common.RunnerSettings{
				Executor: "virtualbox",
				VirtualBox: &common.VirtualBoxConfig{
					BaseName:         vboxImage,
					DisableSnapshots: true,
				},
				SSH: vboxSSHConfig,
			},
		},
	}

	trace := &common.Trace{Writer: os.Stdout}

	abortTimer := time.AfterFunc(time.Second, func() {
		t.Log("Interrupt")
		trace.CancelFunc()
	})
	defer abortTimer.Stop()

	timeoutTimer := time.AfterFunc(time.Minute, func() {
		t.Log("Timedout")
		t.FailNow()
	})
	defer timeoutTimer.Stop()

	err = build.Run(&common.Config{}, trace)
	assert.IsType(t, err, &common.BuildError{})
	assert.EqualError(t, err, "canceled")
}
