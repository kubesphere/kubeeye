package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/kubesphere/event-rule-engine/visitor"
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
)

type systemdInspect struct {
}

func init() {
	RuleOperatorMap[constant.Systemd] = &systemdInspect{}
}

func (o *systemdInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var systemdRules []kubeeyev1alpha2.SysRule
	_ = json.Unmarshal(jobRule.RunRule, &systemdRules)

	if systemdRules == nil && len(systemdRules) == 0 {
		return nil, fmt.Errorf("systemdRules is empty")
	}
	var jobTemplate *v1.Job
	if systemdRules[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *systemdRules[0].NodeName, nil, constant.Systemd)
	} else if systemdRules[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", systemdRules[0].NodeSelector, constant.Systemd)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.Systemd)
	}

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, err

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
				Name:  r.Name,
				Level: r.Level,
			}
			for _, status := range unitsContext {
				if status.Name == fmt.Sprintf("%s.service", r.Name) {
					ctl.Value = &status.ActiveState

					if r.Rule != nil {
						if _, err := visitor.CheckRule(*r.Rule); err != nil {
							sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
							klog.Error(sprintf)
							ctl.Value = &sprintf
						} else {
							err, res := visitor.EventRuleEvaluate(map[string]interface{}{r.Name: status.ActiveState}, *r.Rule)
							if err != nil {
								sprintf := fmt.Sprintf("err:%s", err.Error())
								ctl.Value = &sprintf
							} else {
								ctl.Assert = !res
							}
						}
					}
					break
				}
			}
			if ctl.Value == nil {
				errVal := fmt.Sprintf("name:%s to does not exist", r.Name)
				ctl.Assert = true
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

	for _, item := range systemdResult {

		item.NodeName = runNodeName
		resultCr.Spec.SystemdResult = append(resultCr.Spec.SystemdResult, item)
	}

	return resultCr, nil

}
