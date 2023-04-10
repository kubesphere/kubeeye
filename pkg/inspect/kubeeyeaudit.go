package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/rules"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
	certexp   = "data.kubeeye_certexpiration"
)

type PercentOutput struct {
	TotalAuditCount   int
	CurrentAuditCount int
	AuditPercent      int
}
type OutputType string

func AuditCluster(ctx context.Context, kubeConfigPath string, additionalregoruleputh string, output OutputType) error {
	kubeConfig, err := kube.GetKubeConfig(kubeConfigPath)
	if err != nil {
		return errors.Wrap(err, "Failed to load config file")
	}

	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		return err
	}

	_, validationResultsChan, _ := ValidationResults(ctx, clients, types.NamespacedName{}, additionalregoruleputh)

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		if err := JSONOutput(validationResultsChan); err != nil {
			return err
		}
	case "CSV", "csv", "Csv":
		if err := CSVOutput(validationResultsChan); err != nil {
			return err
		}
	default:
		if err := defaultOutput(validationResultsChan); err != nil {
			return err
		}
	}
	return nil
}

func ValidationResults(ctx context.Context, kubernetesClient *kube.KubernetesClient, taskName types.NamespacedName, additionalregoruleputh string) (kube.K8SResource, <-chan []kubeeyev1alpha2.ResourceResult, *PercentOutput) {
	// get kubernetes resources and put into the channel.
	klog.Info("starting get kubernetes resources")

	k8sResources := kube.GetK8SResources(ctx, kubernetesClient)

	auditPercent := &PercentOutput{}

	if k8sResources.Deployments != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Deployments.Items)
	}
	if k8sResources.StatefulSets != nil {
		auditPercent.TotalAuditCount += len(k8sResources.StatefulSets.Items)
	}
	if k8sResources.DaemonSets != nil {
		auditPercent.TotalAuditCount += len(k8sResources.DaemonSets.Items)
	}
	if k8sResources.Jobs != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Jobs.Items)
	}
	if k8sResources.CronJobs != nil {
		auditPercent.TotalAuditCount += len(k8sResources.CronJobs.Items)
	}
	if k8sResources.Roles != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Roles.Items)
	}
	if k8sResources.ClusterRoles != nil {
		auditPercent.TotalAuditCount += len(k8sResources.ClusterRoles.Items)
	}
	if k8sResources.Nodes != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Nodes.Items)
	}
	if k8sResources.Events != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Events.Items)
	}
	auditPercent.TotalAuditCount++
	auditPercent.CurrentAuditCount = auditPercent.TotalAuditCount

	klog.Info("getting and merging the Rego ruleFiles")
	getRules, ruleType := rules.GetRules(ctx, taskName, kubernetesClient.VersionClientSet)
	var RulesValidateChan <-chan []kubeeyev1alpha2.ResourceResult
	if ruleType == constant.Opa {
		var opaRules []kubeeyev1alpha2.OpaRule
		err := json.Unmarshal(getRules, &opaRules)
		if err != nil {
			fmt.Printf("unmarshal opaRule failed,err:%s\n", err)
		}
		var RegoRules []string
		for i := range opaRules {
			RegoRules = append(RegoRules, opaRules[i].Rule)
		}
		regoRulesChan := rules.MergeRegoRules(ctx, RegoRules, rules.GetAdditionalRegoRulesfiles(additionalregoruleputh))
		RulesValidateChan = MergeRegoRulesValidate(ctx, regoRulesChan,
			RegoRulesValidate(workloads, k8sResources, auditPercent),
			RegoRulesValidate(rbac, k8sResources, auditPercent),
			RegoRulesValidate(events, k8sResources, auditPercent),
			RegoRulesValidate(nodes, k8sResources, auditPercent),
			RegoRulesValidate(certexp, k8sResources, auditPercent),
		)
	}

	klog.Info("starting inspect kubernetes resources")

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	klog.Info("get inspect results")
	return k8sResources, RulesValidateChan, auditPercent
}

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
