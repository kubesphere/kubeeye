package template

import (
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
)

type JobTemplateOptions struct {
	JobConfig    *conf.JobConfig
	JobName      string
	Task         *kubeeyev1alpha2.InspectTask
	NodeName     string
	NodeSelector map[string]string
	RuleType     string
}
