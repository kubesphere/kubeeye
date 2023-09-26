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
	"github.com/sergi/go-diff/diffmatchpatch"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strings"
)

type fileChangeInspect struct {
}

func init() {
	RuleOperatorMap[constant.FileChange] = &fileChangeInspect{}
}

func (o *fileChangeInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var fileRule []kubeeyev1alpha2.FileChangeRule
	_ = json.Unmarshal(jobRule.RunRule, &fileRule)

	if fileRule == nil {
		return nil, fmt.Errorf("file change rule is empty")
	}
	var jobTemplate *v1.Job
	if fileRule[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *fileRule[0].NodeName, nil, constant.FileChange)
	} else if fileRule[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", fileRule[0].NodeSelector, constant.FileChange)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.FileChange)
	}

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", jobTemplate.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *fileChangeInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var fileResults []kubeeyev1alpha2.FileChangeResultItem
	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
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
		resultItem := kubeeyev1alpha2.FileChangeResultItem{
			FileName: file.Name,
			Path:     file.Path,
			Level:    file.Level,
		}

		baseFile, fileErr := os.ReadFile(path.Join(constant.RootPathPrefix, file.Path))
		if fileErr != nil {
			klog.Errorf("Failed to open base file path:%s,error:%s", baseFile, fileErr)
			resultItem.Issues = []string{fmt.Sprintf("%s:The file does not exist", file.Name)}
			fileResults = append(fileResults, resultItem)
			continue
		}
		baseFileName := fmt.Sprintf("%s-%s", constant.BaseFilePrefix, file.Name)
		baseConfig, configErr := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Get(ctx, baseFileName, metav1.GetOptions{})
		if configErr != nil {
			klog.Errorf("Failed to open file. cause：file Do not exist,err:%s", err)
			if kubeErr.IsNotFound(configErr) {
				mapTemplate := template.BinaryFileConfigMapTemplate(baseFileName, constant.DefaultNamespace, baseFile, true)
				_, createErr := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Create(ctx, mapTemplate, metav1.CreateOptions{})
				if createErr != nil {
					resultItem.Issues = []string{fmt.Sprintf("%s:create configMap failed", file.Name)}
				}
				fileResults = append(fileResults, resultItem)
				continue
			}
		}
		baseContent := baseConfig.BinaryData[constant.FileChange]

		diffResult := diffString(string(baseContent), string(baseFile))

		for i := range diffResult {
			diffResult[i] = strings.ReplaceAll(diffResult[i], "\x1b[32m", "")
			diffResult[i] = strings.ReplaceAll(diffResult[i], "\x1b[31m", "")
			diffResult[i] = strings.ReplaceAll(diffResult[i], "\x1b[0m", "")
		}
		resultItem.Issues = diffResult
		fileResults = append(fileResults, resultItem)
	}

	marshal, err := json.Marshal(fileResults)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *fileChangeInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var fileChangeResult []kubeeyev1alpha2.FileChangeResultItem
	jsonErr := json.Unmarshal(resultCm.BinaryData[constant.Data], &fileChangeResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return nil, jsonErr
	}

	for i := range fileChangeResult {
		fileChangeResult[i].NodeName = runNodeName
	}
	resultCr.Spec.FileChangeResult = fileChangeResult
	return resultCr, nil

}

func diffString(base1 string, base2 string) []string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(base1, base2, false)
	scan := bufio.NewScanner(strings.NewReader(dmp.DiffPrettyText(diffs)))
	lineNum := 1
	var isseus []string
	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "\x1b[3") {
			isseus = append(isseus, fmt.Sprintf("%d行 %s\n", lineNum, line))
		}
		lineNum++
	}
	return isseus
}
