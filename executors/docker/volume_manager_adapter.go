package docker

import (
	"context"

	"gitlab.com/gitlab-org/gitlab-runner/executors/docker/internal/volumes"
	docker_helpers "gitlab.com/gitlab-org/gitlab-runner/helpers/docker"
)

type volumesManagerAdapter struct {
	docker_helpers.Client

	e *executor
}

func (a *volumesManagerAdapter) ContainerLabels(containerType string, otherLabels ...string) map[string]string {
	typeLabel := a.e.containerTypeLabel(containerType)
	return a.e.labels(append(otherLabels, typeLabel)...)
}

func (a *volumesManagerAdapter) WaitForContainer(id string) error {
	return a.e.waitForContainer(a.e.Context, id)
}

func (a *volumesManagerAdapter) RemoveContainer(ctx context.Context, id string) error {
	return a.e.removeContainer(ctx, id)
}

var createVolumesManager = func(e *executor) (volumes.Manager, error) {
	adapter := &volumesManagerAdapter{
		Client: e.client,
		e:      e,
	}

	helperImage, err := e.getPrebuiltImage()
	if err != nil {
		return nil, err
	}

	ccManager := volumes.NewCacheContainerManager(
		e.Context,
		&e.BuildLogger,
		adapter,
		helperImage,
	)

	config := volumes.ManagerConfig{
		CacheDir:          e.Config.Docker.CacheDir,
		BaseContainerPath: e.Build.FullProjectDir(),
		UniqueName:        e.Build.ProjectUniqueName(),
		DisableCache:      e.Config.Docker.DisableCache,
	}

	volumesManager := volumes.NewManager(&e.BuildLogger, e.volumeParser, ccManager, config)

	return volumesManager, nil
}

func (e *executor) createVolumesManager() error {
	vm, err := createVolumesManager(e)
	if err != nil {
		return err
	}

	e.volumesManager = vm

	return nil
}
