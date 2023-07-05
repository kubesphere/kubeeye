package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type opaInspect struct {
}

func init() {
	RuleOperatorMap[constant.Opa] = &opaInspect{}
}

func (o *opaInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {

	var jobNames []kubeeyev1alpha2.JobPhase

	job := template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", nil, constant.Opa)

	_, err := clients.ClientSet.BatchV1().Jobs("kubeeye-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning})

	return jobNames, nil
}

func (o *opaInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	klog.Info("getting  Rego rules")

	_, exist, phase := utils.ArrayFinds(task.Spec.Rules, func(m kubeeyev1alpha2.JobRule) bool {
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

func (o *opaInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var opaResult kubeeyev1alpha2.KubeeyeOpaResult
	err := json.Unmarshal(result.BinaryData[constant.Result], &opaResult)
	if err != nil {
		return err
	}
	var ownerRefBol = true
	resultRef := metav1.OwnerReference{
		APIVersion:         task.APIVersion,
		Kind:               task.Kind,
		Name:               task.Name,
		UID:                task.UID,
		Controller:         &ownerRefBol,
		BlockOwnerDeletion: &ownerRefBol,
	}

	var inspectResult kubeeyev1alpha2.InspectResult
	inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.Opa)
	inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.OpaResult = opaResult
	err = c.Create(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
}
