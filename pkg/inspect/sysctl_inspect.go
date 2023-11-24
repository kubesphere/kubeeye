package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/prometheus/procfs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

type sysctlInspect struct {
}

func init() {
	RuleOperatorMap[constant.Sysctl] = &sysctlInspect{}
}

func (o *sysctlInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, informers informers.SharedInformerFactory, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var SysctlResult []kubeeyev1alpha2.NodeMetricsResultItem

	fs, err := procfs.NewFS(constant.DefaultProcPath)
	if err != nil {
		return nil, err
	}

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var sysctl []kubeeyev1alpha2.SysRule
		err = json.Unmarshal(phase.RunRule, &sysctl)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		for _, sysRule := range sysctl {
			ctlRule, err := fs.SysctlStrings(sysRule.Name)
			klog.Infof("name:%s,value:%s", sysRule.Name, ctlRule)
			ctl := kubeeyev1alpha2.NodeMetricsResultItem{
				BaseResult: kubeeyev1alpha2.BaseResult{Name: sysRule.Name},
			}
			if err != nil {
				errVal := fmt.Sprintf("name:%s to does not exist", sysRule.Name)
				ctl.Value = &errVal
				ctl.Level = sysRule.Level
				ctl.Assert = true
			} else {
				val := parseSysctlVal(ctlRule)
				ctl.Value = &val
				if !utils.IsEmptyValue(sysRule.Rule) {
					err, res := visitor.EventRuleEvaluate(map[string]interface{}{sysRule.Name: val}, sysRule.Rule)
					if err != nil {
						evalErr := fmt.Sprintf("event rule evaluate to failed err:%s", err)
						ctl.Assert = true
						ctl.Value = &evalErr
						ctl.Level = sysRule.Level
					} else {
						if !res {
							ctl.Level = sysRule.Level
						}
						ctl.Assert = !res
					}

				}
			}
			SysctlResult = append(SysctlResult, ctl)
		}

	}

	marshal, err := json.Marshal(SysctlResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *sysctlInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var SysctlResult []kubeeyev1alpha2.NodeMetricsResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &SysctlResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for i := range SysctlResult {
		SysctlResult[i].NodeName = runNodeName

	}
	resultCr.Spec.SysctlResult = append(resultCr.Spec.SysctlResult, SysctlResult...)
	return resultCr, nil

}

func parseSysctlVal(val []string) string {
	if len(val) == 0 && val == nil {
		return ""
	}
	return val[0]
}
