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

type mergeType string

const (
	nodeName     mergeType = "nodeName"
	nodeSelector mergeType = "nodeSelector"
)

type fileChangeInspect struct {
}

func init() {
	RuleOperatorMap[constant.FileChange] = &fileChangeInspect{}
}

func (o *fileChangeInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	jobName := fmt.Sprintf("%s-%s", task.Name, constant.FileChange)

	var fileChangeRules []kubeeyev1alpha2.FileChangeRule

	_ = json.Unmarshal(task.Spec.Rules[constant.FileChange], &fileChangeRules)

	nodeData, filterData := utils.ArrayFilter(fileChangeRules, func(v kubeeyev1alpha2.FileChangeRule) bool {
		return v.NodeName != nil
	})

	nodeNameRule, nodeNameStatus := mergeNodeRule(nodeData, nodeName)
	if nodeNameStatus {
		for key, v := range nodeNameRule {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, v[0].Name), task, key, nil, constant.FileChange)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(v)
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}

	}

	nodeSelectorData, residueData := utils.ArrayFilter(filterData, func(v kubeeyev1alpha2.FileChangeRule) bool {
		return v.NodeSelector != nil
	})
	nodeSelectorRule, nodeSelectorStatus := mergeNodeRule(nodeSelectorData, nodeSelector)
	if nodeSelectorStatus {
		for k, v := range nodeSelectorRule {
			labelsMap, _ := labels.ConvertSelectorToLabelsMap(k)
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, k), task, "", labelsMap, constant.FileChange)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(v)
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}
	}

	if len(residueData) > 0 {
		nodeAll, err := clients.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, nodeItem := range nodeAll.Items {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, nodeItem.Name), task, nodeItem.Name, nil, constant.FileChange)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(filterData)

			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}
	}

	return jobNames, nil
}

func (o *fileChangeInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var fileResults []kubeeyev1alpha2.FileChangeResultItem
	_, exist, phase := utils.ArrayFinds(task.Status.JobPhase, func(m kubeeyev1alpha2.JobPhase) bool {
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
		baseConfig, configErr := clients.ClientSet.CoreV1().ConfigMaps(task.Namespace).Get(ctx, baseFileName, metav1.GetOptions{})
		if configErr != nil {
			klog.Errorf("Failed to open file. cause：file Do not exist,err:%s", err)
			if kubeErr.IsNotFound(configErr) {
				var Immutable = true
				baseConfigMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:            baseFileName,
						Namespace:       task.Namespace,
						OwnerReferences: ownerRef,
						Labels:          map[string]string{constant.LabelConfigType: constant.BaseFile},
					},
					Immutable:  &Immutable,
					BinaryData: map[string][]byte{constant.FileChange: baseFile},
				}
				_, createErr := clients.ClientSet.CoreV1().ConfigMaps(task.Namespace).Create(ctx, baseConfigMap, metav1.CreateOptions{})
				if createErr != nil {
					resultItem.Issues = []string{fmt.Sprintf("%s:create configMap failed", file.Name)}
					fileResults = append(fileResults, resultItem)
				}
				continue
			}
		}
		baseContent := baseConfig.BinaryData[constant.FileChange]
		//baseContent configmap读取的基准内容  baseFile文件读取需要对比的内容
		diffString := utils.DiffString(string(baseContent), string(baseFile))

		for i := range diffString {
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[32m", "")
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[31m", "")
			diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[0m", "")
		}
		resultItem.Issues = diffString
		fileResults = append(fileResults, resultItem)
	}

	if fileResults == nil && len(fileResults) == 0 {
		return nil, nil
	}

	marshal, err := json.Marshal(fileResults)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *fileChangeInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Namespace: task.Namespace,
		Name:      fmt.Sprintf("%s-nodeinfo", task.Name),
	}, &inspectResult)
	var fileChangeResult []kubeeyev1alpha2.FileChangeResultItem
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &fileChangeResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return err
	}
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
			inspectResult.Namespace = task.Namespace
			inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
			inspectResult.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{jobs.Spec.Template.Spec.NodeName: {FileChangeResult: fileChangeResult}}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	infoResult, ok := inspectResult.Spec.NodeInfoResult[jobs.Spec.Template.Spec.NodeName]
	if ok {
		infoResult.FileChangeResult = append(infoResult.FileChangeResult, fileChangeResult...)
	} else {
		infoResult.FileChangeResult = fileChangeResult
	}

	inspectResult.Spec.NodeInfoResult[jobs.Spec.Template.Spec.NodeName] = infoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}

func mergeNodeRule(rule []kubeeyev1alpha2.FileChangeRule, types mergeType) (map[string][]kubeeyev1alpha2.FileChangeRule, bool) {
	var mergeNodeMap = make(map[string][]kubeeyev1alpha2.FileChangeRule)
	exists := false
	switch types {
	case nodeName:
		for _, changeRule := range rule {
			mergeNodeMap[*changeRule.NodeName] = append(mergeNodeMap[*changeRule.NodeName], changeRule)
			exists = true
		}
		break
	case nodeSelector:
		for _, changeRule := range rule {
			formatLabels := labels.FormatLabels(changeRule.NodeSelector)
			mergeNodeMap[formatLabels] = append(mergeNodeMap[formatLabels], changeRule)
		}
		break
	}

	return mergeNodeMap, exists
}
