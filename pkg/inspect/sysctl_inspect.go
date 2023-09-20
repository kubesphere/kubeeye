package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/prometheus/procfs"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type sysctlInspect struct {
}

func init() {
	RuleOperatorMap[constant.Sysctl] = &sysctlInspect{}
}

func (o *sysctlInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var sysRules []kubeeyev1alpha2.SysRule
	_ = json.Unmarshal(jobRule.RunRule, &sysRules)

	if sysRules == nil && len(sysRules) == 0 {
		return nil, fmt.Errorf("sysctl rule is empty")
	}

	var jobTemplate *v1.Job
	if sysRules[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *sysRules[0].NodeName, nil, constant.Sysctl)
	} else if sysRules[0].NodeSelector != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", sysRules[0].NodeSelector, constant.Sysctl)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.Sysctl)
	}

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, err

}

func (o *sysctlInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var nodeInfoResult kubeeyev1alpha2.NodeInfoResult

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
			ctl := kubeeyev1alpha2.NodeResultItem{
				Name:  sysRule.Name,
				Level: sysRule.Level,
			}
			if err != nil {
				errVal := fmt.Sprintf("name:%s to does not exist", sysRule.Name)
				ctl.Value = &errVal
				ctl.Assert = true
			} else {
				val := parseSysctlVal(ctlRule)
				ctl.Value = &val
				if sysRule.Rule != nil {
					if _, err := visitor.CheckRule(*sysRule.Rule); err != nil {
						checkErr := fmt.Sprintf("rule condition is not correct, %s", err.Error())
						ctl.Value = &checkErr
						ctl.Assert = true
					} else {
						err, res := visitor.EventRuleEvaluate(map[string]interface{}{sysRule.Name: val}, *sysRule.Rule)
						if err != nil {
							evalErr := fmt.Sprintf("event rule evaluate to failed err:%s", err)
							ctl.Assert = true
							ctl.Value = &evalErr
						} else {
							ctl.Assert = !res
						}

					}

				}
			}
			nodeInfoResult.SysctlResult = append(nodeInfoResult.SysctlResult, ctl)
		}

	}

	marshal, err := json.Marshal(nodeInfoResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *sysctlInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var nodeInfoResult kubeeyev1alpha2.NodeInfoResult
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &nodeInfoResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	if resultCr.Spec.NodeInfoResult == nil {
		resultCr.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{runNodeName: nodeInfoResult}
		return resultCr, nil
	}

	infoResult, ok := resultCr.Spec.NodeInfoResult[runNodeName]
	if ok {
		infoResult.SysctlResult = append(infoResult.SysctlResult, nodeInfoResult.SysctlResult...)
	} else {
		infoResult = nodeInfoResult
	}

	resultCr.Spec.NodeInfoResult[runNodeName] = infoResult

	return resultCr, nil

}

func parseSysctlVal(val []string) string {
	if len(val) == 0 && val == nil {
		return ""
	}
	return val[0]
}
