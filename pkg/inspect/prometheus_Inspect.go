package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/conf"
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
	"strings"
	"time"
)

type prometheusInspect struct {
}

func init() {
	RuleOperatorMap[constant.Prometheus] = &prometheusInspect{}
}

func (o *prometheusInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	job := template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.Prometheus)

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", job.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, err

}

func (o *prometheusInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
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
		if queryResults.Len() == 0 {
			continue
		}
		var queryResultsMap []map[string]string
		for i, result := range queryResults {
			temp := map[string]string{"value": result.Value.String(), "time": result.Timestamp.String()}
			klog.Info(i, result)
			for name, value := range result.Metric {
				temp[formatName(name)] = string(value)
			}
			queryResultsMap = append(queryResultsMap, temp)
		}

		proRuleResult = append(proRuleResult, queryResultsMap)
	}

	marshal, err := json.Marshal(proRuleResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func (o *prometheusInspect) GetResult(ctx context.Context, c *kube.KubernetesClient, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var prometheus [][]map[string]string

	err := json.Unmarshal(result.BinaryData[constant.Data], &prometheus)
	if err != nil {
		return err
	}
	if prometheus == nil {
		return nil
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
	inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.PrometheusResult = prometheus
	//err = c.Create(ctx, &inspectResult)
	_, err = c.VersionClientSet.KubeeyeV1alpha2().RESTClient().Post().Resource("inspectresults").Body(&inspectResult).DoRaw(ctx)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
}

func formatName(name model.LabelName) string {
	return strings.Trim(string(name), "_")
}
