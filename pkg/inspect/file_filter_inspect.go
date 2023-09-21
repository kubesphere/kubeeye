package inspect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"path"
	"regexp"
)

type fileFilterInspect struct {
}

func init() {
	RuleOperatorMap[constant.FileFilter] = &fileFilterInspect{}
}

func (o *fileFilterInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var filterRules []kubeeyev1alpha2.FileFilterRule
	_ = json.Unmarshal(jobRule.RunRule, &filterRules)

	if filterRules == nil && len(filterRules) == 0 {
		return nil, fmt.Errorf("file filter rule is empty")
	}
	var jobTemplate *v1.Job
	if filterRules[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *filterRules[0].NodeName, nil, constant.FileFilter)
	} else if filterRules[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", filterRules[0].NodeSelector, constant.FileFilter)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.FileFilter)
	}

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", jobTemplate.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *fileFilterInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var filterResult []kubeeyev1alpha2.FileChangeResultItem

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var filter []kubeeyev1alpha2.FileFilterRule
		err := json.Unmarshal(phase.RunRule, &filter)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}
		for _, rule := range filter {
			file, err := os.OpenFile(path.Join(constant.RootPathPrefix, rule.Path), os.O_RDONLY, 0222)
			filterR := kubeeyev1alpha2.FileChangeResultItem{
				FileName: rule.Name,
				Path:     rule.Path,
				Level:    rule.Level,
			}
			if err != nil {
				klog.Errorf(" Failed to open file . err:%s", err)
				filterR.Issues = append(filterR.Issues, fmt.Sprintf("Failed to open file for %s.", rule.Name))
				filterResult = append(filterResult, filterR)
				continue
			}
			reader := bufio.NewScanner(file)
			for reader.Scan() {
				matched, err := regexp.MatchString(fmt.Sprintf(".%s.", *rule.Rule), reader.Text())
				if err != nil {
					klog.Errorf(" Failed to match regex. err:%s", err)
					filterR.Issues = append(filterR.Issues, fmt.Sprintf("Failed to match regex for %s.", *rule.Rule))
					break
				}
				if matched && len(filterR.Issues) < 1000 {
					filterR.Issues = append(filterR.Issues, reader.Text())
				}
			}
			filterResult = append(filterResult, filterR)
		}
	}

	marshal, err := json.Marshal(filterResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *fileFilterInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var fileFilterResult []kubeeyev1alpha2.FileChangeResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &fileFilterResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for _, item := range fileFilterResult {
		item.NodeName = runNodeName
		resultCr.Spec.FileFilterResult = append(resultCr.Spec.FileFilterResult, item)
	}

	return resultCr, nil

}