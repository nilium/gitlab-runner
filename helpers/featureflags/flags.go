package featureflags

import (
	"strconv"
)

const (
	CmdDisableDelayedErrorLevelExpansion string = "FF_CMD_DISABLE_DELAYED_ERROR_LEVEL_EXPANSION"
	UseLegacyBuildsDirForDocker          string = "FF_USE_LEGACY_BUILDS_DIR_FOR_DOCKER"
	UseLegacyVolumesMountingOrder        string = "FF_USE_LEGACY_VOLUMES_MOUNTING_ORDER"
	ShellExecutorUseLegacyProcessKill    string = "FF_SHELL_EXECUTOR_USE_LEGACY_PROCESS_KILL"
)

type FeatureFlag struct {
	Name            string
	DefaultValue    string
	Deprecated      bool
	ToBeRemovedWith string
	Description     string
}

// REMEMBER to update the documentation after adding or removing a feature flag
//
// Please use `make update_feature_flags_docs` to make the update automatic and
// properly formatted. It will replace the existing table with the new one, computed
// basing on the values below
var flags = []FeatureFlag{
	{
		Name:            CmdDisableDelayedErrorLevelExpansion,
		DefaultValue:    "false",
		Deprecated:      false,
		ToBeRemovedWith: "",
		Description:     "Disables [EnableDelayedExpansion](https://ss64.com/nt/delayedexpansion.html) for error checking for when using [Window Batch](https://docs.gitlab.com/runner/shells/#windows-batch) shell",
	},
	{
		Name:            UseLegacyBuildsDirForDocker,
		DefaultValue:    "false",
		Deprecated:      true,
		ToBeRemovedWith: "12.7",
		Description:     "Disables the new strategy for Docker executor to cache the content of `/builds` directory instead of `/builds/group-org`",
	},
	{
		Name:            UseLegacyVolumesMountingOrder,
		DefaultValue:    "false",
		Deprecated:      true,
		ToBeRemovedWith: "12.6",
		Description:     "Disables the new ordering of volumes mounting when `docker*` executors are being used.",
	},
	{
		Name:            ShellExecutorUseLegacyProcessKill,
		DefaultValue:    "false",
		Deprecated:      true,
		ToBeRemovedWith: "13.0",
		Description:     "Use the old process termination that was used prior to GitLab 12.8 where only `SIGKILL` was sent",
	},
}

func GetAll() []FeatureFlag {
	return flags
}

func IsOn(value string) (bool, error) {
	if value == "" {
		return false, nil
	}

	on, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return on, nil
}
