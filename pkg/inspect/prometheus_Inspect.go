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
	"github.com/prometheus/client_golang/api"
	apiprometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type prometheusInspect struct {
}

func init() {
	RuleOperatorMap[constant.Prometheus] = &prometheusInspect{}
}

func (o *prometheusInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase

	job := template.InspectJobsTemplate(jobRule.JobName, task, "", nil, constant.Prometheus)

	_, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning})

	return jobNames, nil
}

func (o *prometheusInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(task.Spec.Rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if !exist {
		return nil, nil
	}

	var proRules []kubeeyev1alpha2.PrometheusRule
	err := json.Unmarshal(phase.RunRule, &proRules)
	if err != nil {
		klog.Error(err, " Failed to marshal kubeeye result")
		return nil, err
	}

	var proRuleResult [][]map[string]string
	for _, proRule := range proRules {
		proClient, err := api.NewClient(api.Config{
			Address: proRule.Endpoint,
		})
		if err != nil {
			klog.Error("create prometheus client failed", err)
			continue
		}
		queryApi := apiprometheusv1.NewAPI(proClient)
		query, _, err := queryApi.Query(ctx, *proRule.Rule, time.Now())
		if err != nil {
			klog.Errorf("failed to query rule:%s", *proRule.Rule)
			return nil, err
		}
		marshal, err := json.Marshal(query)

		var queryResults model.Samples
		err = json.Unmarshal(marshal, &queryResults)
		if err != nil {
			klog.Error("unmarshal modal Samples failed", err)
			continue
		}
		var queryResultsMap []map[string]string
		for i, result := range queryResults {
			temp := map[string]string{"value": result.Value.String(), "time": result.Timestamp.String()}
			klog.Info(i, result)
			for name, value := range result.Metric {
				klog.Info(name, value)
				temp[string(name)] = string(value)
			}
			queryResultsMap = append(queryResultsMap, temp)
		}

		proRuleResult = append(proRuleResult, queryResultsMap)
	}
	if proRules == nil && len(proRules) == 0 {
		return nil, nil
	}
	marshal, err := json.Marshal(proRuleResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func (o *prometheusInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var prometheus [][]map[string]string
	err := json.Unmarshal(result.BinaryData[constant.Result], &prometheus)
	klog.Info(prometheus)
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
	inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.Prometheus)
	inspectResult.Namespace = task.Namespace
	inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.PrometheusResult = prometheus
	err = c.Create(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
}
