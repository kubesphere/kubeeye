package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type OpaInspect struct {
}

func init() {
	RuleOperatorMap[constant.Opa] = &OpaInspect{}
}

func (o *OpaInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

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
			klog.Error("marshal opaRule failed,err:%s\n", err)
			return nil, err
		}

		return marshal, nil
	}
	return nil, nil
}

func (o *OpaInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {
	var opaResult kubeeyev1alpha2.KubeeyeOpaResult
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &opaResult)
	if err != nil {
		return nil, err
	}

	resultCr.Spec.OpaResult = opaResult

	return resultCr, nil
}
