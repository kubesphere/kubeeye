package inspect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"path"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fileFilterInspect struct {
}

func init() {
	RuleOperatorMap[constant.FileFilter] = &fileFilterInspect{}
}

func (o *fileFilterInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) (*kubeeyev1alpha2.JobPhase, error) {

	var filterRules []kubeeyev1alpha2.FileFilterRule
	_ = json.Unmarshal(jobRule.RunRule, &filterRules)

	if filterRules == nil && len(filterRules) == 0 {
		return nil, fmt.Errorf("file filter rule is empty")
	}
	var jobTemplate *v1.Job
	if filterRules[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, *filterRules[0].NodeName, nil, constant.FileFilter)
	} else if filterRules[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", filterRules[0].NodeSelector, constant.FileFilter)
	} else {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", nil, constant.FileFilter)
	}

	_, err := clients.ClientSet.BatchV1().Jobs("kubeeye-system").Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", jobTemplate.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *fileFilterInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var filterResult []kubeeyev1alpha2.FileChangeResultItem

	_, exist, phase := utils.ArrayFinds(task.Spec.Rules, func(m kubeeyev1alpha2.JobRule) bool {
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
					filterR.Issues = append(filterR.Issues, fmt.Sprintf("Failed to match regex for %s.", rule.Rule))
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

func (o *fileFilterInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {

	var nodeInfoResult []kubeeyev1alpha2.FileChangeResultItem
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &nodeInfoResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return jsonErr
	}

	if nodeInfoResult == nil {
		return nil
	}
	runNodeName := findJobRunNode(ctx, jobs, c)
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-filefilter", task.Name),
	}, &inspectResult)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			var ownerRefBol = true
			resultRef := metav1.OwnerReference{
				APIVersion:         task.APIVersion,
				Kind:               task.Kind,
				Name:               task.Name,
				UID:                task.UID,
				Controller:         &ownerRefBol,
				BlockOwnerDeletion: &ownerRefBol,
			}
			inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
			inspectResult.Name = fmt.Sprintf("%s-filefilter", task.Name)
			inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
			inspectResult.Spec.FilterResult = map[string][]kubeeyev1alpha2.FileChangeResultItem{runNodeName: nodeInfoResult}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	infoResult, ok := inspectResult.Spec.FilterResult[runNodeName]
	if ok {
		infoResult = append(infoResult, nodeInfoResult...)
	} else {
		infoResult = nodeInfoResult
	}

	inspectResult.Spec.FilterResult[runNodeName] = infoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}
