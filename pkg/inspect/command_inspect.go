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

func (c *commandInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

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
				BaseResult: kubeeyev1alpha2.BaseResult{Name: r.Name},
				Command:    r.Command,
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

func (c *commandInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {

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
