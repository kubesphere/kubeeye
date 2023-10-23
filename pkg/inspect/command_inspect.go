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
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os/exec"
)

type commandInspect struct {
}

func init() {
	RuleOperatorMap[constant.CustomCommand] = &commandInspect{}
}

func (o *commandInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	var commandRules []kubeeyev1alpha2.CustomCommandRule
	_ = json.Unmarshal(jobRule.RunRule, &commandRules)

	if commandRules == nil {
		return nil, fmt.Errorf("commandRules is empty")
	}
	var jobTemplate *v1.Job
	if commandRules[0].NodeName != nil {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, *commandRules[0].NodeName, nil, constant.CustomCommand)
	} else {
		jobTemplate = template.InspectJobsTemplate(config, jobRule.JobName, task, "", commandRules[0].NodeSelector, constant.CustomCommand)
	}

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, jobTemplate, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, err

}

func (o *commandInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var commandResult []kubeeyev1alpha2.CommandResultItem

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var commandRules []kubeeyev1alpha2.CustomCommandRule
		err := json.Unmarshal(phase.RunRule, &commandRules)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}
		for _, r := range commandRules {
			ctl := kubeeyev1alpha2.CommandResultItem{
				Name:    r.Name,
				Command: r.Command,
			}
			command := exec.Command("sh", "-c", r.Command)
			outputResult, err := command.Output()
			if err != nil {
				fmt.Println(err)
				ctl.Value = fmt.Sprintf("command execute failed, %s", err)
				ctl.Level = r.Level
				ctl.Assert = true
				continue
			}

			err, res := visitor.EventRuleEvaluate(map[string]interface{}{"result": string(outputResult)}, *r.Rule)
			if err != nil {
				ctl.Value = fmt.Sprintf("rule evaluate failed err:%s", err)
				ctl.Level = r.Level
				ctl.Assert = true
			} else {
				if res {
					ctl.Level = r.Level
				}
				ctl.Assert = res
			}

			commandResult = append(commandResult, ctl)
		}
	}

	marshal, err := json.Marshal(commandResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *commandInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

	var commandResult []kubeeyev1alpha2.CommandResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &commandResult)
	if err != nil {
		klog.Error("failed to get result", err)
		return nil, err
	}

	for i := range commandResult {
		commandResult[i].NodeName = runNodeName
	}
	resultCr.Spec.CommandResult = commandResult
	return resultCr, nil

}
