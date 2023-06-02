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
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	_ = ValidationResults(ctx, clients, types.NamespacedName{}, nil)

	// Set the output mode, support default output JSON and CSV.
	//switch output {
	//case "JSON", "json", "Json":
	//	if err := JSONOutput(validationResultsChan); err != nil {
	//		return err
	//	}
	//case "CSV", "csv", "Csv":
	//	if err := CSVOutput(); err != nil {
	//		return err
	//	}
	//default:
	//	if err := defaultOutput(validationResultsChan); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func ValidationResults(ctx context.Context, kubernetesClient *kube.KubernetesClient, taskName types.NamespacedName, auditResult *kubeeyev1alpha2.Result) interface{} {
	// get kubernetes resources and put into the channel.
	klog.Info("starting get kubernetes resources")
	//k8sResources := kube.GetK8SResources(ctx, kubernetesClient)
	klog.Info("getting  Rego rules")
	getRules := rules.GetRules(ctx, taskName, kubernetesClient.VersionClientSet)

	for key, rule := range getRules {
		if key == constant.Opa {
			var opaRules []kubeeyev1alpha2.OpaRule
			err := json.Unmarshal(rule, &opaRules)
			if err != nil {
				fmt.Printf("unmarshal opaRule failed,err:%s\n", err)
				continue
			}
			var RegoRules []string
			for i := range opaRules {
				RegoRules = append(RegoRules, *opaRules[i].Rule)
			}

			//return VailOpaRulesResult(ctx, k8sResources, RegoRules)
		} else if key == constant.Prometheus {
			var proRules []kubeeyev1alpha2.PrometheusRule
			err := json.Unmarshal(rule, &proRules)
			if err != nil {
				fmt.Printf("unmarshal opaRule failed,err:%s\n", err)
				continue
			}
			//return PrometheusRulesResult(ctx, proRules)
		}
	}

	return nil
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

func JobInspect(ctx context.Context, taskName string, taskNamespace string, resultName string, clients *kube.KubernetesClient, ruleType string) error {
	task, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks(taskNamespace).Get(ctx, taskName, v1.GetOptions{})
	apiVersion := clients.VersionClientSet.KubeeyeV1alpha2().RESTClient().APIVersion().String()
	if err != nil || task.Spec.Rules == nil {
		klog.Error(err, ",Failed to get file Change rule not found")
		return err
	}

	var ownerRefBol = true
	ownerRef := v1.OwnerReference{
		APIVersion:         apiVersion,
		Kind:               "InspectTask",
		Name:               task.Name,
		UID:                task.UID,
		Controller:         &ownerRefBol,
		BlockOwnerDeletion: &ownerRefBol,
	}
	var result []byte

	inspectInterface, status := RuleOperatorMap[ruleType]
	if status {
		result, err = inspectInterface.RunInspect(ctx, task, clients, resultName, ownerRef)
	}
	if err != nil {
		return err
	}

	if result == nil {
		klog.Info("There are no new issues")
		return nil
	}
	resultConfigMap := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:            resultName,
			Namespace:       task.Namespace,
			OwnerReferences: []v1.OwnerReference{ownerRef},
			Labels:          map[string]string{constant.LabelName: task.Name, constant.LabelConfigType: constant.Result},
		},
		BinaryData: map[string][]byte{constant.Result: result},
	}
	_, err = clients.ClientSet.CoreV1().ConfigMaps(task.Namespace).Create(ctx, resultConfigMap, v1.CreateOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("create configMap failed. err:%s", err))
	}

	return nil
}
