# Shell

The Shell executor is a simple executor that allows you to execute builds
locally to the machine that the Runner is installed. It supports all systems on
which the Runner can be installed. That means that it's possible to use scripts
generated for Bash, Windows PowerShell and Windows Batch.

NOTE: **Note:**
Always use the latest version of Git available. Additionally, GitLab Runner will use
the `git lfs` command if [Git LFS](https://git-lfs.github.com) is installed on the machine,
so ensure Git LFS is up-to-date when GitLab Runner will run using the shell executor.

## Overview

The scripts can be run as unprivileged user if the `--user` is added to the
[`gitlab-runner run` command][run]. This feature is only supported by Bash.

The source project is checked out to:
`<working-directory>/builds/<short-token>/<concurrent-id>/<namespace>/<project-name>`.

The caches for project are stored in
`<working-directory>/cache/<namespace>/<project-name>`.

Where:

- `<working-directory>` is the value of `--working-directory` as passed to the
  `gitlab-runner run` command or the current directory where the Runner is
  running
- `<short-token>` is a shortened version of the Runner's token (first 8 letters)
- `<concurrent-id>` is a unique number, identifying the local job ID on the
  particular Runner in context of the project
- `<namespace>` is the namespace where the project is stored on GitLab
- `<project-name>` is the name of the project as it is stored on GitLab

To overwrite the `<working-directory>/builds` and `<working-directory/cache`
specify the `builds_dir` and `cache_dir` options under the `[[runners]]` section
in [`config.toml`](../configuration/advanced-configuration.md).

## Running as unprivileged user

If GitLab Runner is installed on Linux from the [official `.deb` or `.rpm`
packages][packages], the installer will try to use the `gitlab_ci_multi_runner`
user if found. If it is not found, it will create a `gitlab-runner` user and use
this instead.

All shell builds will be then executed as either the `gitlab-runner` or
`gitlab_ci_multi_runner` user.

In some testing scenarios, your builds may need to access some privileged
resources, like Docker Engine or VirtualBox. In that case you need to add the
`gitlab-runner` user to the respective group:

```bash
usermod -aG docker gitlab-runner
usermod -aG vboxusers gitlab-runner
```

## Selecting your shell

GitLab Runner [supports certain shells](../shells/index.md). To select a shell, specify it in your `config.toml` file. For example:

```toml
...
[[runners]]
  name = "shell executor runner"
  executor = "shell"
  shell = "powershell"
...
```

## Security

Generally it's unsafe to run tests with shell executors. The jobs are run with
the user's permissions (`gitlab-runner`) and can "steal" code from other
projects that are run on this server. Use it only for running builds on a
server you trust and own.

## Terminating and killing processes

GitLab Runner terminates any process under any of the following
conditions:

- The job [times out](https://docs.gitlab.com/ee/user/project/pipelines/settings.html#timeout).
- The job is canceled.

The shell executor starts the script for the job in a new process and on
UNIX systems it sets the main process as a [process
group](http://www.informit.com/articles/article.aspx?p=397655&seqNum=6).

### GitLab 12.7 or lower

On UNIX it will send a `SIGKILL` to the process to terminate it since
the child process belongs to the same process group this signal is also
sent to them. Whilst on Windows it will send a `taskkill /F
/T`.

### GitLab 12.8 or higher

On UNIX it will first send `SIGTERM` to the process and it's child
processes, and after 10 minutes it will send `SIGKILL`. This allows
graceful termination for the process. Windows don't have a `SIGTERM`
equivalent so the kill process it sent twice, the second is sent after
10 minutes.

If for some reason this new termination process has problems with your
scripts but works with the  [old method](#gitlab-127-or-lower) you can
set the feature flag
[`FF_SHELL_EXECUTOR_USE_LEGACY_PROCESS_KILL`](../configuration/feature-flags.md)
to `true`, and it will use the old method. Keep in mind that this
feature flag will be removed in GitLab Runner 13.0 so you still need to
fix your script to handle the new termination.

[run]: ../commands/README.md#gitlab-runner-run
[packages]: https://packages.gitlab.com/runner/gitlab-runner
