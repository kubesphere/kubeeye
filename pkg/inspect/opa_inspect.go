package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye-v1alpha2/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye-v1alpha2/constant"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/conf"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/kube"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/template"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type opaInspect struct {
}

func init() {
	RuleOperatorMap[constant.Opa] = &opaInspect{}
}

func (o *opaInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	job := template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.Opa)

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", job.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *opaInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	klog.Info("getting  Rego rules")

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		k8sResources := kube.GetK8SResources(ctx, clients)
		var opaRules []kubeeyev1alpha2.OpaRule
		err := json.Unmarshal(phase.RunRule, &opaRules)
		if err != nil {
			fmt.Printf("unmarshal opaRule failed,err:%s\n", err)
			return nil, err
		}
		var RegoRules []string
		for i := range opaRules {
			RegoRules = append(RegoRules, *opaRules[i].Rule)
		}

		result := VailOpaRulesResult(ctx, k8sResources, RegoRules)
		marshal, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		return marshal, nil
	}
	return nil, nil
}

func (o *opaInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {
	var opaResult kubeeyev1alpha2.KubeeyeOpaResult
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &opaResult)
	if err != nil {
		return nil, err
	}

	resultCr.Spec.OpaResult = opaResult

	return resultCr, nil
}
