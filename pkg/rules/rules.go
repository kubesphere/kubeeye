package rules

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
	"time"
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

func MergeRule(rules []kubeeyev1alpha2.InspectRule) (ruleSpec kubeeyev1alpha2.InspectRuleSpec) {
	for _, rule := range rules {
		if rule.Spec.Opas != nil && len(rule.Spec.Opas) > 0 {
			ruleSpec.Opas = append(ruleSpec.Opas, rule.Spec.Opas...)
			ruleSpec.Opas = RuleArrayDeduplication[kubeeyev1alpha2.OpaRule](ruleSpec.Opas)
		}
		if rule.Spec.Prometheus != nil {
			for _, pro := range rule.Spec.Prometheus {
				if "" != rule.Spec.PrometheusEndpoint && len(rule.Spec.PrometheusEndpoint) > 0 {
					pro.Endpoint = rule.Spec.PrometheusEndpoint
				}
				_, b, _ := utils.ArrayFinds(ruleSpec.Prometheus, func(m kubeeyev1alpha2.PrometheusRule) bool {
					return m.Name == pro.Name
				})
				if !b {
					ruleSpec.Prometheus = append(ruleSpec.Prometheus, pro)
				}
			}
		}
		if rule.Spec.FileChange != nil && len(rule.Spec.FileChange) > 0 {
			ruleSpec.FileChange = append(ruleSpec.FileChange, rule.Spec.FileChange...)
			ruleSpec.FileChange = RuleArrayDeduplication[kubeeyev1alpha2.FileChangeRule](ruleSpec.FileChange)
		}
		if rule.Spec.Sysctl != nil && len(rule.Spec.Sysctl) > 0 {
			ruleSpec.Sysctl = append(ruleSpec.Sysctl, rule.Spec.Sysctl...)
			ruleSpec.Sysctl = RuleArrayDeduplication[kubeeyev1alpha2.SysRule](ruleSpec.Sysctl)
		}
		if rule.Spec.Systemd != nil && len(rule.Spec.Systemd) > 0 {
			ruleSpec.Systemd = append(ruleSpec.Systemd, rule.Spec.Systemd...)
			ruleSpec.Systemd = RuleArrayDeduplication[kubeeyev1alpha2.SysRule](ruleSpec.Systemd)
		}
		if rule.Spec.FileFilter != nil && len(rule.Spec.FileFilter) > 0 {
			ruleSpec.FileFilter = append(ruleSpec.FileFilter, rule.Spec.FileFilter...)
			ruleSpec.FileFilter = RuleArrayDeduplication[kubeeyev1alpha2.FileFilterRule](ruleSpec.FileFilter)
		}
		ruleSpec.Component = rule.Spec.Component
	}
	return ruleSpec
}

type MapType interface {
	kubeeyev1alpha2.SysRule | kubeeyev1alpha2.OpaRule | kubeeyev1alpha2.PrometheusRule | kubeeyev1alpha2.FileChangeRule | kubeeyev1alpha2.FileFilterRule
}

func StructToMap(obj interface{}) []map[string]interface{} {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var result []map[string]interface{}
	err = json.Unmarshal(marshal, &result)
	if err != nil {
		return nil
	}
	return result
}

func MapToStruct[T MapType](maps []map[string]interface{}) []T {
	var result []T
	marshal, err := json.Marshal(maps)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(marshal, &result)
	if err != nil {
		return nil
	}
	return result
}
func RuleArrayDeduplication[T MapType](obj interface{}) []T {
	maps := StructToMap(obj)
	var newMaps []map[string]interface{}
	for _, m := range maps {
		_, b, _ := utils.ArrayFinds(newMaps, func(m1 map[string]interface{}) bool {
			return m["name"] == m1["name"]
		})
		if !b {
			newMaps = append(newMaps, m)
		}
	}
	toStruct := MapToStruct[T](newMaps)
	return toStruct
}

func AllocationOpa(rule []kubeeyev1alpha2.OpaRule, taskName string) *kubeeyev1alpha2.JobRule {
	if rule == nil {
		return nil
	}

	jobRule := &kubeeyev1alpha2.JobRule{
		JobName:  fmt.Sprintf("%s-%s", taskName, constant.Opa),
		RuleType: constant.Opa,
	}

	opa, err := json.Marshal(rule)
	if err != nil {
		klog.Errorf("Failed to marshal  opa rule. err:%s", err)
		return nil
	}
	jobRule.RunRule = opa
	return jobRule
}

func AllocationComponent(components *string, taskName string) *kubeeyev1alpha2.JobRule {

	jobRule := &kubeeyev1alpha2.JobRule{
		JobName:  fmt.Sprintf("%s-%s", taskName, constant.Component),
		RuleType: constant.Component,
	}

	opa, err := json.Marshal(components)
	if err != nil {
		klog.Errorf("Failed to marshal  opa rule. err:%s", err)
		return nil
	}
	jobRule.RunRule = opa
	return jobRule
}

func AllocationPrometheus(rule []kubeeyev1alpha2.PrometheusRule, taskName string) *kubeeyev1alpha2.JobRule {
	if rule == nil {
		return nil
	}

	jobRule := &kubeeyev1alpha2.JobRule{
		JobName:  fmt.Sprintf("%s-%s", taskName, constant.Prometheus),
		RuleType: constant.Prometheus,
	}

	prometheus, err := json.Marshal(rule)
	if err != nil {
		klog.Errorf("Failed to marshal  prometheus rule. err:%s", err)
		return nil
	}
	jobRule.RunRule = prometheus
	return jobRule
}
func AllocationRule(rule interface{}, taskName string, allNode []corev1.Node, ctlOrTem string) []kubeeyev1alpha2.JobRule {
	if rule == nil {
		return nil
	}
	toMap := StructToMap(rule)

	nodeData, filterData := utils.ArrayFilter(toMap, func(v map[string]interface{}) bool {
		return v["nodeName"] != nil || v["nodeSelector"] != nil
	})
	var jobRules []kubeeyev1alpha2.JobRule
	nodeNameMergeMap := mergeNodeRule(nodeData)

	for _, v := range nodeNameMergeMap {
		jobRule := kubeeyev1alpha2.JobRule{
			JobName:  fmt.Sprintf("%s-%s-%d", taskName, ctlOrTem, time.Now().Unix()),
			RuleType: ctlOrTem,
		}
		fileChange, err := json.Marshal(v)
		if err != nil {
			klog.Errorf("Failed to marshal  fileChange rule. err:%s", err)
			return nil
		}
		jobRule.RunRule = fileChange
		jobRules = append(jobRules, jobRule)
	}

	if len(filterData) > 0 {
		for _, item := range allNode {
			jobRule := kubeeyev1alpha2.JobRule{
				JobName:  fmt.Sprintf("%s-%s-%s", taskName, ctlOrTem, item.Name),
				RuleType: ctlOrTem,
			}

			for i := range filterData {
				filterData[i]["nodeName"] = &item.Name
			}
			sysMarshal, err := json.Marshal(filterData)
			if err != nil {
				klog.Errorf("Failed to marshal  fileChange rule. err:%s", err)
				return nil
			}
			jobRule.RunRule = sysMarshal
			jobRules = append(jobRules, jobRule)
		}
	}

	return jobRules
}

func mergeNodeRule(rule []map[string]interface{}) map[string][]map[string]interface{} {
	var mergeMap = make(map[string][]map[string]interface{})
	for _, m := range rule {
		for k, v := range m {
			if k == "nodeName" {
				mergeMap[v.(string)] = append(mergeMap[v.(string)], m)
			} else if k == "nodeSelector" {
				formatLabels := labels.FormatLabels(v.(map[string]string))
				mergeMap[formatLabels] = append(mergeMap[formatLabels], m)
			}
		}
	}
	return mergeMap
}

func ParseRules(ctx context.Context, clients *kube.KubernetesClient, taskName string, ruleGroup []kubeeyev1alpha2.InspectRule) ([]kubeeyev1alpha2.JobRule, map[string]int) {

	nodes := kube.GetNodes(ctx, clients.ClientSet)
	ruleSpec := MergeRule(ruleGroup)
	var inspectRuleTotal = make(map[string]int)
	var executeRule []kubeeyev1alpha2.JobRule

	component := AllocationComponent(ruleSpec.Component, taskName)
	executeRule = append(executeRule, *component)
	componentRuleNumber := 0
	if ruleSpec.Component == nil {
		services, _ := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		componentRuleNumber = len(services.Items)
	} else {
		componentRuleNumber = len(strings.Split(*ruleSpec.Component, "|"))
	}
	inspectRuleTotal[constant.Component] = componentRuleNumber
	opa := AllocationOpa(ruleSpec.Opas, taskName)
	if opa != nil {
		executeRule = append(executeRule, *opa)
		inspectRuleTotal[constant.Opa] = len(ruleSpec.Opas)
	}
	prometheus := AllocationPrometheus(ruleSpec.Prometheus, taskName)
	if prometheus != nil {
		executeRule = append(executeRule, *prometheus)
		inspectRuleTotal[constant.Prometheus] = len(ruleSpec.Prometheus)

	}
	if len(nodes) > 0 {
		change := AllocationRule(ruleSpec.FileChange, taskName, nodes, constant.FileChange)
		if len(change) > 0 {
			executeRule = append(executeRule, change...)
			inspectRuleTotal[constant.FileChange] = len(ruleSpec.FileChange)
		}
		sysctl := AllocationRule(ruleSpec.Sysctl, taskName, nodes, constant.Sysctl)
		if len(sysctl) > 0 {
			executeRule = append(executeRule, sysctl...)
			inspectRuleTotal[constant.Sysctl] = len(ruleSpec.Sysctl)
		}
		systemd := AllocationRule(ruleSpec.Systemd, taskName, nodes, constant.Systemd)
		if len(systemd) > 0 {
			executeRule = append(executeRule, systemd...)
			inspectRuleTotal[constant.Systemd] = len(ruleSpec.Systemd)

		}
		fileFilter := AllocationRule(ruleSpec.FileFilter, taskName, nodes, constant.FileFilter)
		if len(fileFilter) > 0 {
			executeRule = append(executeRule, fileFilter...)
			inspectRuleTotal[constant.FileFilter] = len(ruleSpec.FileFilter)
		}
	}
	return executeRule, inspectRuleTotal
}

func UpdateRuleReferNum(ctx context.Context, clients *kube.KubernetesClient) error {
	listRule, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().List(ctx, metav1.ListOptions{})
	listPlan, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error(err, "Failed to list inspectRules and inspectPlan")
		return err
	}
	for _, item := range listRule.Items {
		filter, _ := utils.ArrayFilter(listPlan.Items, func(v kubeeyev1alpha2.InspectPlan) bool {
			return v.Spec.Tag == item.GetLabels()[constant.LabelRuleGroup]
		})
		item.Annotations = labels.Merge(item.Annotations, map[string]string{constant.AnnotationRuleJoinNum: strconv.Itoa(len(filter))})
		_, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Update(ctx, &item, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err, "Failed to update inspectRules refer num")
			return err
		}
	}

	return nil
}
