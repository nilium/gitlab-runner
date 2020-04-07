package labels

import (
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/gitlab-runner/common"
)

const dockerLabelPrefix = "com.gitlab.gitlab-runner"

// Labeler is responsible for handling labelling logic for docker entities - networks, containers.
type Labeler interface {
	Labels(otherLabels ...string) map[string]string
}

// NewLabeler returns a new instance of a Labeler bound to this build.
func NewLabeler(b *common.Build) Labeler {
	return &defaultLabeler{
		build: b,
	}
}

type defaultLabeler struct {
	build *common.Build
}

func (l *defaultLabeler) Labels(otherLabels ...string) map[string]string {
	labels := map[string]string{
		dockerLabelPrefix + ".job.id":          strconv.Itoa(l.build.ID),
		dockerLabelPrefix + ".job.sha":         l.build.GitInfo.Sha,
		dockerLabelPrefix + ".job.before_sha":  l.build.GitInfo.BeforeSha,
		dockerLabelPrefix + ".job.ref":         l.build.GitInfo.Ref,
		dockerLabelPrefix + ".project.id":      strconv.Itoa(l.build.JobInfo.ProjectID),
		dockerLabelPrefix + ".pipeline.id":     l.build.GetAllVariables().Get("CI_PIPELINE_ID"),
		dockerLabelPrefix + ".runner.id":       l.build.Runner.ShortDescription(),
		dockerLabelPrefix + ".runner.local_id": strconv.Itoa(l.build.RunnerID),
	}
	for _, label := range otherLabels {
		keyValue := strings.SplitN(label, "=", 2)
		if len(keyValue) == 2 {
			labels[dockerLabelPrefix+"."+keyValue[0]] = keyValue[1]
		}
	}
	return labels
}
