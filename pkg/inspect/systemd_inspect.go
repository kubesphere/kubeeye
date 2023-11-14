package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type systemdInspect struct {
}

func init() {
	RuleOperatorMap[constant.Systemd] = &systemdInspect{}
}

func (o *systemdInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var nodeResult []kubeeyev1alpha2.NodeMetricsResultItem

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var systemd []kubeeyev1alpha2.SysRule
		err := json.Unmarshal(phase.RunRule, &systemd)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		conn, err := dbus.NewWithContext(ctx)
		if err != nil {
			return nil, err
		}
		unitsContext, err := conn.ListUnitsContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range systemd {
			ctl := kubeeyev1alpha2.NodeMetricsResultItem{
				BaseResult: kubeeyev1alpha2.BaseResult{Name: r.Name},
			}
			for _, status := range unitsContext {
				if status.Name == fmt.Sprintf("%s.service", r.Name) {
					ctl.Value = &status.ActiveState
					if r.Rule != nil {
						err, res := visitor.EventRuleEvaluate(map[string]interface{}{r.Name: status.ActiveState}, *r.Rule)
						if err != nil {
							sprintf := fmt.Sprintf("err:%s", err.Error())
							ctl.Value = &sprintf
							ctl.Assert = true
							ctl.Level = r.Level
						} else {
							if !res {
								ctl.Level = r.Level
							}
							ctl.Assert = !res
						}
					}
					break
				}
			}
			if ctl.Value == nil {
				errVal := fmt.Sprintf("name:%s to does not exist", r.Name)
				ctl.Assert = true
				ctl.Level = r.Level
				ctl.Value = &errVal
			}
			nodeResult = append(nodeResult, ctl)
		}
	}

	marshal, err := json.Marshal(nodeResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *systemdInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var systemdResult []kubeeyev1alpha2.NodeMetricsResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &systemdResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for i := range systemdResult {
		systemdResult[i].NodeName = runNodeName
	}
	resultCr.Spec.SystemdResult = append(resultCr.Spec.SystemdResult, systemdResult...)
	return resultCr, nil

}
