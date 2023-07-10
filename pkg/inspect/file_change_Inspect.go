package inspect

import (
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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type fileChangeInspect struct {
}

func init() {
	RuleOperatorMap[constant.FileChange] = &fileChangeInspect{}
}

func (o *fileChangeInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) (*kubeeyev1alpha2.JobPhase, error) {

	var fileRule []kubeeyev1alpha2.FileChangeRule
	_ = json.Unmarshal(jobRule.RunRule, &fileRule)

	if fileRule == nil && len(fileRule) == 0 {
		return nil, fmt.Errorf("file change rule is empty")
	}
	var jobTemplate *v1.Job
	if fileRule[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, *fileRule[0].NodeName, nil, constant.FileChange)
	} else if fileRule[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", fileRule[0].NodeSelector, constant.FileChange)
	} else {
		jobTemplate = template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", nil, constant.FileChange)
	}

	_, err := clients.ClientSet.BatchV1().Jobs("kubeeye-system").Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", jobTemplate.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *fileChangeInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var fileResults []kubeeyev1alpha2.FileChangeResultItem
	_, exist, phase := utils.ArrayFinds(task.Spec.Rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})
	if !exist {
		return nil, fmt.Errorf("rule not exist")
	}
	var fileRule []kubeeyev1alpha2.FileChangeRule
	err := json.Unmarshal(phase.RunRule, &fileRule)
	if err != nil {
		klog.Error(err, " Failed to marshal kubeeye result")
		return nil, err
	}

	for _, file := range fileRule {
		var resultItem kubeeyev1alpha2.FileChangeResultItem

		resultItem.FileName = file.Name
		resultItem.Path = file.Path
		baseFile, fileErr := os.ReadFile(path.Join(constant.RootPathPrefix, file.Path))
		if fileErr != nil {
			klog.Errorf("Failed to open base file path:%s,error:%s", baseFile, fileErr)
			resultItem.Issues = []string{fmt.Sprintf("%s:The file does not exist", file.Name)}
			fileResults = append(fileResults, resultItem)

			continue
		}
		baseFileName := fmt.Sprintf("%s-%s", constant.BaseFilePrefix, file.Name)
		baseConfig, configErr := clients.ClientSet.CoreV1().ConfigMaps("kubeeye-system").Get(ctx, baseFileName, metav1.GetOptions{})
		if configErr != nil {
			klog.Errorf("Failed to open file. cause：file Do not exist,err:%s", err)

			if kubeErr.IsNotFound(configErr) {

				mapTemplate := template.BinaryFileConfigMapTemplate(baseFileName, task.Namespace, baseFile, true)
				_, createErr := clients.ClientSet.CoreV1().ConfigMaps("kubeeye-system").Create(ctx, mapTemplate, metav1.CreateOptions{})
				if createErr != nil {
					resultItem.Issues = []string{fmt.Sprintf("%s:create configMap failed", file.Name)}
				} else {
					resultItem.Issues = []string{fmt.Sprintf("success  initial base config file. name:%s", file.Name)}
				}
				fileResults = append(fileResults, resultItem)
				continue
			}
		}
		baseContent := baseConfig.BinaryData[constant.FileChange]

		diffString := utils.DiffString(string(baseContent), string(baseFile))

		for i := range diffString {
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[32m", "")
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[31m", "")
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[0m", "")
		}
		resultItem.Issues = diffString
		fileResults = append(fileResults, resultItem)
	}

	marshal, err := json.Marshal(fileResults)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *fileChangeInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {

	var fileChangeResult []kubeeyev1alpha2.FileChangeResultItem
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &fileChangeResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return jsonErr
	}
	if fileChangeResult == nil {
		return nil
	}
	runNodeName := findJobRunNode(ctx, jobs, c)
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Name: fmt.Sprintf("%s-nodeinfo", task.Name),
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
			inspectResult.Name = fmt.Sprintf("%s-nodeinfo", task.Name)
			inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
			inspectResult.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{runNodeName: {FileChangeResult: fileChangeResult}}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	infoResult, ok := inspectResult.Spec.NodeInfoResult[runNodeName]
	if ok {
		infoResult.FileChangeResult = append(infoResult.FileChangeResult, fileChangeResult...)
	} else {
		infoResult.FileChangeResult = fileChangeResult
	}

	inspectResult.Spec.NodeInfoResult[runNodeName] = infoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}

func findJobRunNode(ctx context.Context, job *v1.Job, c client.Client) string {
	var pods corev1.PodList

	err := c.List(ctx, &pods, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{"job-name": job.Name}),
		Namespace:     job.Namespace,
	})
	if err != nil {
		klog.Error(err)
		return ""
	}
	for _, item := range pods.Items {
		if item.Status.Phase == corev1.PodSucceeded {
			return item.Spec.NodeName
		}
	}

	return ""
}
