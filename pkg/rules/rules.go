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

func MergeRule(rules ...kubeeyev1alpha2.InspectRule) (*kubeeyev1alpha2.InspectRuleSpec, error) {
	ruleSpec := &kubeeyev1alpha2.InspectRuleSpec{}
	for _, rule := range rules {
		if rule.Spec.Opas != nil {
			opas, err := RuleArrayDeduplication[kubeeyev1alpha2.OpaRule](append(ruleSpec.Opas, rule.Spec.Opas...))
			if err != nil {
				return nil, err
			}
			ruleSpec.Opas = opas
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
			fileChange, err := RuleArrayDeduplication[kubeeyev1alpha2.FileChangeRule](append(ruleSpec.FileChange, rule.Spec.FileChange...))
			if err != nil {
				return nil, err
			}
			ruleSpec.FileChange = fileChange
		}
		if rule.Spec.Sysctl != nil {
			sysctl, err := RuleArrayDeduplication[kubeeyev1alpha2.SysRule](append(ruleSpec.Sysctl, rule.Spec.Sysctl...))
			if err != nil {
				return nil, err
			}
			ruleSpec.Sysctl = sysctl
		}
		if rule.Spec.NodeInfo != nil {

			nodeInfo, err := RuleArrayDeduplication[kubeeyev1alpha2.NodeInfo](append(ruleSpec.NodeInfo, rule.Spec.NodeInfo...))
			if err != nil {
				return nil, err
			}
			ruleSpec.NodeInfo = nodeInfo
		}
		if rule.Spec.Systemd != nil {

			systemd, err := RuleArrayDeduplication[kubeeyev1alpha2.SysRule](append(ruleSpec.Systemd, rule.Spec.Systemd...))
			if err != nil {
				return nil, err
			}
			ruleSpec.Systemd = systemd
		}
		if rule.Spec.FileFilter != nil {
			fileFilter, err := RuleArrayDeduplication[kubeeyev1alpha2.FileFilterRule](append(ruleSpec.FileFilter, rule.Spec.FileFilter...))
			if err != nil {
				return nil, err
			}
			ruleSpec.FileFilter = fileFilter
		}
		if rule.Spec.CustomCommand != nil {
			command, err := RuleArrayDeduplication[kubeeyev1alpha2.CustomCommandRule](append(ruleSpec.CustomCommand, rule.Spec.CustomCommand...))
			if err != nil {
				return nil, err
			}
			ruleSpec.CustomCommand = command
		}

		ruleSpec.Component = rule.Spec.Component
	}
	return ruleSpec, nil
}

func RuleArrayDeduplication[T any](obj interface{}) ([]T, error) {
	maps, err := utils.StructToMap(obj)
	if err != nil {
		return nil, err
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
	toStruct := utils.MapToStruct[T](newMaps...)
	return toStruct, nil
}

func Allocation(rule interface{}, taskName string, ruleType string) (*kubeeyev1alpha2.JobRule, error) {

	toMap, err := utils.StructToMap(rule)
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
		JobName:  fmt.Sprintf("%s-%s", taskName, ruleType),
		RuleType: ruleType,
		RunRule:  marshalRule,
	}, nil
}

func AllocationRule(rule interface{}, taskName string, allNode []corev1.Node, ctlOrTem string) ([]kubeeyev1alpha2.JobRule, error) {

	toMap, err := utils.StructToMap(rule)
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
			JobName:  fmt.Sprintf("%s-%s-%d", taskName, ctlOrTem, time.Now().Unix()),
			RuleType: ctlOrTem,
		}
		fileChange, err := json.Marshal(v)
		if err != nil {
			klog.Errorf("Failed to marshal  fileChange rule. err:%s", err)
			return nil, err
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
				return nil, err
			}
			jobRule.RunRule = sysMarshal
			jobRules = append(jobRules, jobRule)
		}
	}

	return jobRules, nil
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

func ParseRules(ctx context.Context, clients *kube.KubernetesClient, taskName string, ruleGroup []kubeeyev1alpha2.InspectRule) ([]kubeeyev1alpha2.JobRule, map[string]int, error) {

	nodes := kube.GetNodes(ctx, clients.ClientSet)
	ruleSpec, err := MergeRule(ruleGroup...)
	if err != nil {
		return nil, nil, err
	}
	var inspectRuleTotal = make(map[string]int)
	var executeRule []kubeeyev1alpha2.JobRule
	component, err := Allocation(ruleSpec.Component, taskName, constant.Component)
	if err == nil {
		executeRule = append(executeRule, *component)
		inspectRuleTotal[constant.Component] = TotalServiceNum(ctx, clients, ruleSpec.Component)
	}
	opa, err := Allocation(ruleSpec.Opas, taskName, constant.Opa)
	if err == nil {
		executeRule = append(executeRule, *opa)
		inspectRuleTotal[constant.Opa] = len(ruleSpec.Opas)
	}
	prometheus, err := Allocation(ruleSpec.Prometheus, taskName, constant.Prometheus)
	if err == nil {
		executeRule = append(executeRule, *prometheus)
		inspectRuleTotal[constant.Prometheus] = len(ruleSpec.Prometheus)
	}
	if len(nodes) > 0 {
		change, err := AllocationRule(ruleSpec.FileChange, taskName, nodes, constant.FileChange)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, change...)
		inspectRuleTotal[constant.FileChange] = len(ruleSpec.FileChange)

		sysctl, err := AllocationRule(ruleSpec.Sysctl, taskName, nodes, constant.Sysctl)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, sysctl...)
		inspectRuleTotal[constant.Sysctl] = len(ruleSpec.Sysctl)

		nodeInfo, err := AllocationRule(ruleSpec.NodeInfo, taskName, nodes, constant.NodeInfo)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, nodeInfo...)
		inspectRuleTotal[constant.NodeInfo] = len(ruleSpec.NodeInfo)

		systemd, err := AllocationRule(ruleSpec.Systemd, taskName, nodes, constant.Systemd)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, systemd...)
		inspectRuleTotal[constant.Systemd] = len(ruleSpec.Systemd)

		fileFilter, err := AllocationRule(ruleSpec.FileFilter, taskName, nodes, constant.FileFilter)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, fileFilter...)
		inspectRuleTotal[constant.FileFilter] = len(ruleSpec.FileFilter)

		customCommand, err := AllocationRule(ruleSpec.CustomCommand, taskName, nodes, constant.CustomCommand)
		if err != nil {
			return nil, nil, err
		}
		executeRule = append(executeRule, customCommand...)
		inspectRuleTotal[constant.CustomCommand] = len(ruleSpec.CustomCommand)

	}
	return executeRule, inspectRuleTotal, nil
}

func TotalServiceNum(ctx context.Context, clients *kube.KubernetesClient, component *string) int {
	componentRuleNumber := 0
	if component == nil {
		services, err := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Failed to list services. err:%s", err)
			return componentRuleNumber
		}
		componentRuleNumber = len(services.Items)
	} else {
		componentRuleNumber = len(strings.Split(*component, "|"))
	}
	return componentRuleNumber
}
