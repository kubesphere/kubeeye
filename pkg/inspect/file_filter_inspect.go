package inspect

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
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

func (f *fileFilterInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, informers informers.SharedInformerFactory, ownerRef ...metav1.OwnerReference) ([]byte, error) {

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
				Path:       rule.Path,
				BaseResult: kubeeyev1alpha2.BaseResult{Name: rule.Name},
			}
			if err != nil {
				klog.Errorf(" Failed to open file . err:%s", err)
				filterR.Issues = append(filterR.Issues, fmt.Sprintf("Failed to open file for %s.", rule.Name))
				filterR.Level = rule.Level
				filterR.Assert = true
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
			if len(filterR.Issues) > 0 {
				filterR.Assert = true
				filterR.Level = rule.Level
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

func (f *fileFilterInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var fileFilterResult []kubeeyev1alpha2.FileChangeResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &fileFilterResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for i := range fileFilterResult {
		fileFilterResult[i].NodeName = runNodeName
	}
	resultCr.Spec.FileFilterResult = append(resultCr.Spec.FileFilterResult, fileFilterResult...)
	return resultCr, nil

}
