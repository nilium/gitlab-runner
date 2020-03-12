package machine

import (
	"testing"
	"time"

	"gitlab.com/gitlab-org/gitlab-runner/common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestIfMachineProviderExposesCollectInterface(t *testing.T) {
	var provider common.ExecutorProvider
	provider = &machineProvider{}
	collector, ok := provider.(prometheus.Collector)
	assert.True(t, ok)
	assert.NotNil(t, collector)
}

func TestMachineProviderDeadInterval(t *testing.T) {
	provider := newMachineProvider("docker_machines", "docker")
	assert.Equal(t, 0, provider.collectDetails().Idle)

	details := provider.getMachineDetails("test")
	assert.Equal(t, 1, provider.collectDetails().Idle)

	details.LastSeenAt = time.Now().Add(-machineDeadInterval)
	assert.Equal(t, 0, provider.collectDetails().Idle)
}
