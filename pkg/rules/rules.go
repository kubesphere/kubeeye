package rules

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"
)

func GetRules(ctx context.Context, task types.NamespacedName, client versioned.Interface) map[string][]byte {

	_, err := client.KubeeyeV1alpha2().InspectTasks().Get(ctx, task.Name, metav1.GetOptions{})
	if err != nil {
		if kubeErr.IsNotFound(err) {
			fmt.Printf("rego ruleFiles not found .\n")
			return nil
		}
		fmt.Printf("Failed to Get rego ruleFiles.\n")
		return nil
	}
	return nil
}

func MergeRule(task *kubeeyev1alpha2.InspectTask, rules ...kubeeyev1alpha2.InspectRule) (*kubeeyev1alpha2.InspectRuleSpec, error) {
	ruleSpec := &kubeeyev1alpha2.InspectRuleSpec{}
	//for _, rule := range rules {
	//	if rule.Spec.Opas != nil {
	//		opas, err := RuleArrayDeduplication[kubeeyev1alpha2.OpaRule](append(ruleSpec.Opas, rule.Spec.Opas...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.Opas = opas
	//	}
	//	if rule.Spec.Prometheus != nil {
	//		for _, pro := range rule.Spec.Prometheus {
	//			if "" != rule.Spec.PrometheusEndpoint && len(rule.Spec.PrometheusEndpoint) > 0 {
	//				pro.Endpoint = rule.Spec.PrometheusEndpoint
	//			}
	//			_, b, _ := utils.ArrayFinds(ruleSpec.Prometheus, func(m kubeeyev1alpha2.PrometheusRule) bool {
	//				return m.Name == pro.Name
	//			})
	//			if !b {
	//				ruleSpec.Prometheus = append(ruleSpec.Prometheus, pro)
	//			}
	//		}
	//	}
	//	if rule.Spec.FileChange != nil && len(rule.Spec.FileChange) > 0 {
	//		fileChange, err := RuleArrayDeduplication[kubeeyev1alpha2.FileChangeRule](append(ruleSpec.FileChange, rule.Spec.FileChange...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.FileChange = fileChange
	//	}
	//	if rule.Spec.Sysctl != nil {
	//		sysctl, err := RuleArrayDeduplication[kubeeyev1alpha2.SysRule](append(ruleSpec.Sysctl, rule.Spec.Sysctl...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.Sysctl = sysctl
	//	}
	//	if rule.Spec.NodeInfo != nil {
	//
	//		nodeInfo, err := RuleArrayDeduplication[kubeeyev1alpha2.NodeInfo](append(ruleSpec.NodeInfo, rule.Spec.NodeInfo...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.NodeInfo = nodeInfo
	//	}
	//	if rule.Spec.Systemd != nil {
	//
	//		systemd, err := RuleArrayDeduplication[kubeeyev1alpha2.SysRule](append(ruleSpec.Systemd, rule.Spec.Systemd...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.Systemd = systemd
	//	}
	//	if rule.Spec.FileFilter != nil {
	//		fileFilter, err := RuleArrayDeduplication[kubeeyev1alpha2.FileFilterRule](append(ruleSpec.FileFilter, rule.Spec.FileFilter...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.FileFilter = fileFilter
	//	}
	//	if rule.Spec.CustomCommand != nil {
	//		command, err := RuleArrayDeduplication[kubeeyev1alpha2.CustomCommandRule](append(ruleSpec.CustomCommand, rule.Spec.CustomCommand...))
	//		if err != nil {
	//			return nil, err
	//		}
	//		ruleSpec.CustomCommand = command
	//	}
	//
	//	ruleSpec.Component = rule.Spec.Component
	//}
	return ruleSpec, nil
}

func RuleArrayDeduplication[T any](obj interface{}) []T {
	maps, err := utils.ArrayStructToArrayMap(obj)
	if err != nil {
		return nil
	}
	var newMaps []map[string]interface{}
	for _, m := range maps {
		_, b, _ := utils.ArrayFinds(newMaps, func(m1 map[string]interface{}) bool {
			return m["name"] == m1["name"]
		})
		if !b {
			newMaps = append(newMaps, m)
		}
	}
	return utils.MapToStruct[T](newMaps...)

}

func Allocation(rule interface{}, taskName string, ruleType string) (*kubeeyev1alpha2.JobRule, error) {

	toMap, err := utils.ArrayStructToArrayMap(rule)
	if err != nil {
		klog.Errorf("Failed to convert rule to map. err:%s", err)
		return nil, err
	}
	if toMap == nil && ruleType != constant.Component {
		return nil, fmt.Errorf("failed to Allocation rule for empty")
	}

	marshalRule, err := json.Marshal(toMap)
	if err != nil {
		return nil, err
	}

	return &kubeeyev1alpha2.JobRule{
		JobName:  fmt.Sprintf("%s-%s-%s", taskName, ruleType, rand.String(5)),
		RuleType: ruleType,
		RunRule:  marshalRule,
	}, nil
}

func AllocationRule(rule interface{}, taskName string, allNode []corev1.Node, ctlOrTem string) ([]kubeeyev1alpha2.JobRule, error) {

	toMap, err := utils.ArrayStructToArrayMap(rule)
	if err != nil {
		klog.Errorf("Failed to convert rule to map. err:%s", err)
		return nil, err
	}

	nodeData, filterData := utils.ArrayFilter(toMap, func(v map[string]interface{}) bool {
		return v["nodeName"] != nil || v["nodeSelector"] != nil
	})
	var jobRules []kubeeyev1alpha2.JobRule
	nodeNameMergeMap := mergeNodeRule(nodeData)

	for _, v := range nodeNameMergeMap {
		jobRule := kubeeyev1alpha2.JobRule{
			JobName:  fmt.Sprintf("%s-%s-%s", taskName, ctlOrTem, rand.String(5)),
			RuleType: ctlOrTem,
		}
		jobRule.RunRule, err = json.Marshal(v)
		if err != nil {
			klog.Errorf("Failed to marshal  fileChange rule. err:%s", err)
			return nil, err
		}

		jobRules = append(jobRules, jobRule)
	}

	if len(filterData) > 0 {
		for _, item := range allNode {
			jobRule := kubeeyev1alpha2.JobRule{
				JobName:  fmt.Sprintf("%s-%s-%s", taskName, ctlOrTem, rand.String(5)),
				RuleType: ctlOrTem,
			}

			for i := range filterData {
				filterData[i]["nodeName"] = &item.Name
			}
			jobRule.RunRule, err = json.Marshal(filterData)
			if err != nil {
				klog.Errorf("Failed to marshal  fileChange rule. err:%s", err)
				return nil, err
			}

			jobRules = append(jobRules, jobRule)
		}
	}

	return jobRules, nil
}

func mergeNodeRule(rule []map[string]interface{}) map[string][]map[string]interface{} {
	var mergeMap = make(map[string][]map[string]interface{})
	for _, m := range rule {
		nnv, nnvExist := m["nodeName"]
		nsv, nsvExist := m["nodeSelector"]
		if nnvExist {
			mergeMap[nnv.(string)] = append(mergeMap[nnv.(string)], m)
		} else if nsvExist {
			convertMap := utils.MapValConvert[string](nsv.(map[string]interface{}))
			formatLabels := labels.FormatLabels(convertMap)
			mergeMap[formatLabels] = append(mergeMap[formatLabels], m)
		}
	}
	return mergeMap
}

func ParseRules(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask, ruleGroup []kubeeyev1alpha2.InspectRule) ([]kubeeyev1alpha2.JobRule, map[string]int, error) {

	nodes := kube.GetNodes(ctx, clients.ClientSet)
	ruleSpec, err := MergeRule(task, ruleGroup...)
	if err != nil {
		return nil, nil, err
	}
	var inspectRuleTotal = make(map[string]int)
	var executeRule []kubeeyev1alpha2.JobRule
	component, err := Allocation(ruleSpec.Component, task.Name, constant.Component)
	if err == nil {
		executeRule = append(executeRule, *component)
		inspectRuleTotal[constant.Component] = TotalServiceNum(ctx, clients, ruleSpec.Component)
	}
	opa, err := Allocation(ruleSpec.Opas, task.Name, constant.Opa)
	if err == nil {
		executeRule = append(executeRule, *opa)
		inspectRuleTotal[constant.Opa] = len(ruleSpec.Opas)
	}
	prometheus, err := Allocation(ruleSpec.Prometheus, task.Name, constant.Prometheus)
	if err == nil {
		executeRule = append(executeRule, *prometheus)
		inspectRuleTotal[constant.Prometheus] = len(ruleSpec.Prometheus)
	}
	if len(nodes) > 0 {
		change, err := AllocationRule(ruleSpec.FileChange, task.Name, nodes, constant.FileChange)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, change...)
		inspectRuleTotal[constant.FileChange] = len(ruleSpec.FileChange)

		sysctl, err := AllocationRule(ruleSpec.Sysctl, task.Name, nodes, constant.Sysctl)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, sysctl...)
		inspectRuleTotal[constant.Sysctl] = len(ruleSpec.Sysctl)

		nodeInfo, err := AllocationRule(ruleSpec.NodeInfo, task.Name, nodes, constant.NodeInfo)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, nodeInfo...)
		inspectRuleTotal[constant.NodeInfo] = len(ruleSpec.NodeInfo)

		systemd, err := AllocationRule(ruleSpec.Systemd, task.Name, nodes, constant.Systemd)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, systemd...)
		inspectRuleTotal[constant.Systemd] = len(ruleSpec.Systemd)

		fileFilter, err := AllocationRule(ruleSpec.FileFilter, task.Name, nodes, constant.FileFilter)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, fileFilter...)
		inspectRuleTotal[constant.FileFilter] = len(ruleSpec.FileFilter)

		customCommand, err := AllocationRule(ruleSpec.CustomCommand, task.Name, nodes, constant.CustomCommand)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, customCommand...)
		inspectRuleTotal[constant.CustomCommand] = len(ruleSpec.CustomCommand)

	}
	return executeRule, inspectRuleTotal, nil
}

func TotalServiceNum(ctx context.Context, clients *kube.KubernetesClient, component *kubeeyev1alpha2.ComponentRule) (componentRuleNumber int) {

	if component != nil && component.IncludeComponent != nil {
		componentRuleNumber = len(component.IncludeComponent)
		return componentRuleNumber
	}

	services, err := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list services. err:%s", err)
		return componentRuleNumber
	}
	inspectData, _ := utils.ArrayFilter(services.Items, func(v corev1.Service) bool {
		if component == nil {
			return v.Spec.ClusterIP != "None"
		}
		_, isExist := utils.ArrayFind(v.Name, component.ExcludeComponent)
		return !isExist && v.Spec.ClusterIP != "None"
	})
	componentRuleNumber = len(inspectData)

	return componentRuleNumber
}

type ExecuteRule struct {
	KubeClient              *kube.KubernetesClient
	Task                    *kubeeyev1alpha2.InspectTask
	clusterInspectRuleMap   map[string]string
	clusterInspectRuleNames []string
	ruleTotal               map[string]int
}

func NewExecuteRuleOptions(clients *kube.KubernetesClient, Task *kubeeyev1alpha2.InspectTask) *ExecuteRule {
	clusterInspectRuleNames := []string{constant.Opa, constant.Prometheus, constant.Component}
	clusterInspectRuleMap := map[string]string{
		"opas":          constant.Opa,
		"prometheus":    constant.Prometheus,
		"component":     constant.Component,
		"fileChange":    constant.FileChange,
		"sysctl":        constant.Sysctl,
		"systemd":       constant.Systemd,
		"fileFilter":    constant.FileFilter,
		"customCommand": constant.CustomCommand,
		"nodeInfo":      constant.NodeInfo,
	}
	return &ExecuteRule{
		KubeClient:              clients,
		Task:                    Task,
		clusterInspectRuleNames: clusterInspectRuleNames,
		clusterInspectRuleMap:   clusterInspectRuleMap,
	}
}

func (e *ExecuteRule) SetRuleSchedule(rules []kubeeyev1alpha2.InspectRule) (newRules []kubeeyev1alpha2.InspectRule) {
	for _, r := range e.Task.Spec.RuleNames {
		_, isExist, rule := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.InspectRule) bool {
			return r.Name == m.Name
		})
		if isExist {
			if !utils.IsEmptyValue(r.NodeName) || r.NodeSelector != nil {
				toMap := utils.StructToMap(rule.Spec)
				if toMap != nil {
					for _, v := range toMap {
						switch val := v.(type) {
						case []interface{}:
							for i := range val {
								m := val[i].(map[string]interface{})
								_, nnExist := m["nodeName"]
								_, nsExist := m["nodeSelector"]
								if !nnExist && !nsExist {
									m["nodeName"] = r.NodeName
									m["nodeSelector"] = r.NodeSelector
								}
							}
						}
					}
					rule.Spec = utils.MapToStruct[kubeeyev1alpha2.InspectRuleSpec](toMap)[0]
				}

			}
			newRules = append(newRules, rule)
		}
	}
	return newRules
}

func (e *ExecuteRule) SetPrometheusEndpoint(allRule []kubeeyev1alpha2.InspectRule) []kubeeyev1alpha2.InspectRule {
	for i := range allRule {
		if !utils.IsEmptyValue(allRule[i].Spec.PrometheusEndpoint) && allRule[i].Spec.Prometheus != nil {
			for p := range allRule[i].Spec.Prometheus {
				if utils.IsEmptyValue(allRule[i].Spec.Prometheus[p].Endpoint) {
					allRule[i].Spec.Prometheus[p].Endpoint = allRule[i].Spec.PrometheusEndpoint
				}
			}
		}
	}
	return allRule
}

func (e *ExecuteRule) MergeRule(allRule []kubeeyev1alpha2.InspectRule) (kubeeyev1alpha2.InspectRuleSpec, error) {
	var newRuleSpec kubeeyev1alpha2.InspectRuleSpec
	var newSpec = make(map[string][]interface{})
	ruleTotal := map[string]int{constant.Component: 0}
	for _, rule := range e.SetPrometheusEndpoint(e.SetRuleSchedule(allRule)) {
		if rule.Spec.Component != nil && newRuleSpec.Component == nil {
			newRuleSpec.Component = rule.Spec.Component
		}
		toMap := utils.StructToMap(rule.Spec)
		for k, v := range toMap {
			switch val := v.(type) {
			case []interface{}:
				newSpec[k] = RuleArrayDeduplication[interface{}](append(newSpec[k], val...))
				ruleTotal[e.clusterInspectRuleMap[k]] = len(newSpec[k])
			}
		}
	}
	ruleTotal[constant.Component] = TotalServiceNum(context.TODO(), e.KubeClient, newRuleSpec.Component)

	marshal, err := json.Marshal(newSpec)
	if err != nil {
		return newRuleSpec, err
	}
	err = json.Unmarshal(marshal, &newRuleSpec)
	if err != nil {
		return newRuleSpec, err
	}
	e.ruleTotal = ruleTotal
	return newRuleSpec, nil
}

func (e *ExecuteRule) GenerateJob(ctx context.Context, rulesSpec kubeeyev1alpha2.InspectRuleSpec) (jobs []kubeeyev1alpha2.JobRule) {

	toMap := utils.StructToMap(rulesSpec)
	nodes := kube.GetNodes(ctx, e.KubeClient.ClientSet)
	for key, v := range toMap {
		mapV, mapExist := e.clusterInspectRuleMap[key]
		if mapExist {
			_, exist := utils.ArrayFind(mapV, e.clusterInspectRuleNames)
			if exist {
				allocation, err := Allocation(v, e.Task.Name, mapV)
				if err == nil {
					jobs = append(jobs, *allocation)
				}
			} else {
				allocationRule, err := AllocationRule(v, e.Task.Name, nodes, mapV)
				if err == nil {
					jobs = append(jobs, allocationRule...)
				}
			}
		}
	}
	_, exist, _ := utils.ArrayFinds(jobs, func(m kubeeyev1alpha2.JobRule) bool {
		return m.RuleType == constant.Component
	})
	if !exist {
		component, err := Allocation(nil, e.Task.Name, constant.Component)
		if err == nil {
			jobs = append(jobs, *component)
		}
	}

	return jobs
}

func (e *ExecuteRule) CreateInspectRule(ctx context.Context, ruleGroup []kubeeyev1alpha2.JobRule) ([]kubeeyev1alpha2.JobRule, error) {
	r := sortRuleOpaToAtLast(ruleGroup)
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	_, err = e.KubeClient.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Get(ctx, e.Task.Name, metav1.GetOptions{})
	if err == nil {
		_ = e.KubeClient.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Delete(ctx, e.Task.Name, metav1.DeleteOptions{})
	}
	// create temp inspect rule
	configMapTemplate := template.BinaryConfigMapTemplate(e.Task.Name, constant.DefaultNamespace, marshal, true, map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"})
	_, err = e.KubeClient.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Create(ctx, configMapTemplate, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func sortRuleOpaToAtLast(rule []kubeeyev1alpha2.JobRule) []kubeeyev1alpha2.JobRule {

	finds, b, OpaRule := utils.ArrayFinds(rule, func(i kubeeyev1alpha2.JobRule) bool {
		return i.RuleType == constant.Opa
	})
	if b {
		rule = append(rule[:finds], rule[finds+1:]...)
		rule = append(rule, OpaRule)
	}

	return rule
}

func (e *ExecuteRule) GetRuleTotal() map[string]int {
	return e.ruleTotal
}
