package inspect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
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

func (f *fileChangeInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

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
		}

		baseFile, fileErr := os.ReadFile(path.Join(constant.RootPathPrefix, file.Path))
		if fileErr != nil {
			klog.Errorf("Failed to open base file path:%s,error:%s", baseFile, fileErr)
			resultItem.Issues = []string{fmt.Sprintf("%s:The file does not exist", file.Name)}
			resultItem.Level = file.Level
			resultItem.Assert = true
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
					resultItem.Level = file.Level
					resultItem.Assert = true
					fileResults = append(fileResults, resultItem)
				}
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
		if len(resultItem.Issues) > 0 {
			resultItem.Level = file.Level
			resultItem.Assert = true
		}
		fileResults = append(fileResults, resultItem)
	}

	marshal, err := json.Marshal(fileResults)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (f *fileChangeInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var fileChangeResult []kubeeyev1alpha2.FileChangeResultItem
	jsonErr := json.Unmarshal(resultCm.BinaryData[constant.Data], &fileChangeResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return nil, jsonErr
	}

	for i := range fileChangeResult {
		fileChangeResult[i].NodeName = runNodeName
	}
	resultCr.Spec.FileChangeResult = append(resultCr.Spec.FileChangeResult, fileChangeResult...)
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
