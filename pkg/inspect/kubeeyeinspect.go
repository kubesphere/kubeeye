package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubeeye/apis/kubeeye/options"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
	certexp   = "data.kubeeye_certexpiration"
)

var RuleOperatorMap = make(map[string]options.InspectInterface)

type PercentOutput struct {
	TotalAuditCount   int
	CurrentAuditCount int
	AuditPercent      int
}
type OutputType string

func CalculateScore(fmResultss []kubeeyev1alpha2.ResourceResult, k8sResources kube.K8SResource) (scoreInfo kubeeyev1alpha2.ScoreInfo) {
	var countDanger int
	var countWarning int
	var countIgnore int

	for _, fmResult := range fmResultss {
		for _, item := range fmResult.ResultItems {
			if item.Level == "warning" {
				countWarning++
			} else if item.Level == "danger" {
				countDanger++
			} else if item.Level == "ignore" {
				countIgnore++
			}
		}
	}

	total := k8sResources.WorkloadsCount*20 + (len(k8sResources.Roles.Items)+len(k8sResources.ClusterRoles.Items))*3 + len(k8sResources.Events.Items) + len(k8sResources.Nodes.Items) + 1
	countSuccess := total - countDanger - countWarning - countIgnore
	totalWeight := countSuccess*2 + countDanger*2 + countWarning
	scoreInfo.Score = countSuccess * 2 * 100 / totalWeight
	scoreInfo.Total = total
	scoreInfo.Dangerous = countDanger
	scoreInfo.Warning = countWarning
	scoreInfo.Ignore = countIgnore
	scoreInfo.Passing = countSuccess

	return scoreInfo
}

func JobInspect(ctx context.Context, taskName string, resultName string, clients *kube.KubernetesClient, ruleType string) error {
	var jobRule []kubeeyev1alpha2.JobRule

	rule, err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"})})
	if err != nil {
		klog.Errorf("Failed to get  inspect Rule. err:%s", err)
		return err
	}
	for _, item := range rule.Items {
		var tempRule []kubeeyev1alpha2.JobRule
		data := item.BinaryData[constant.Data]
		err = json.Unmarshal(data, &tempRule)
		jobRule = append(jobRule, tempRule...)
	}

	var result []byte
	inspectInterface, status := RuleOperatorMap[ruleType]
	if status {
		result, err = inspectInterface.RunInspect(ctx, jobRule, clients, resultName)
	}
	if err != nil {
		return err
	}
	node := findJobRunNode(ctx, resultName, clients.ClientSet)
	resultConfigMap := template.BinaryConfigMapTemplate(resultName, constant.DefaultNamespace, result, true, map[string]string{constant.LabelTaskName: taskName, constant.LabelNodeName: node, constant.LabelRuleType: ruleType})
	_, err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Create(ctx, resultConfigMap, v1.CreateOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("create configMap failed. err:%s", err))
	}

	return nil
}

func findJobRunNode(ctx context.Context, jobName string, c kubernetes.Interface) string {
	pods, err := c.CoreV1().Pods(constant.DefaultNamespace).List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{"job-name": jobName})})
	if err != nil {
		klog.Error(err)
		return ""
	}

	return pods.Items[0].Spec.NodeName
}
