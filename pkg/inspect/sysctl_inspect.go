package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/prometheus/procfs"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type sysctlInspect struct {
}

func init() {
	RuleOperatorMap[constant.Sysctl] = &sysctlInspect{}
}

func (o *sysctlInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	jobName := fmt.Sprintf("%s-%s", task.Name, constant.Sysctl)

	var sysRules []kubeeyev1alpha2.SysRule

	_ = json.Unmarshal(task.Spec.Rules[constant.Sysctl], &sysRules)

	nodeData, filterData := utils.ArrayFilter(sysRules, func(v kubeeyev1alpha2.SysRule) bool {
		return v.NodeName != nil
	})

	nodeNameRule, nodeNameStatus := mergeSysRule(nodeData, nodeName)
	if nodeNameStatus {
		for key, v := range nodeNameRule {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, v[0].Name), task, key, nil, constant.Sysctl)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(v)
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}

	}

	nodeSelectorData, residueData := utils.ArrayFilter(filterData, func(v kubeeyev1alpha2.SysRule) bool {
		return v.NodeSelector != nil
	})
	nodeSelectorRule, nodeSelectorStatus := mergeSysRule(nodeSelectorData, nodeSelector)
	if nodeSelectorStatus {
		for k, v := range nodeSelectorRule {
			labelsMap, _ := labels.ConvertSelectorToLabelsMap(k)
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, k), task, "", labelsMap, constant.Sysctl)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(v)
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}
	}

	if len(residueData) > 0 {
		nodeAll, err := clients.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, nodeItem := range nodeAll.Items {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, nodeItem.Name), task, nodeItem.Name, nil, constant.Sysctl)
			if err != nil {
				klog.Errorf("Failed to create Jobs template for name:%s,err:%s", err, err)
				return nil, err
			}
			createJob, err := clients.ClientSet.BatchV1().Jobs(task.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
				return nil, err
			}
			marshal, _ := json.Marshal(filterData)

			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}
	}

	return jobNames, nil
}

func (o *sysctlInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var nodeInfoResult kubeeyev1alpha2.NodeInfoResult

	fs, err := procfs.NewFS(constant.DefaultProcPath)
	if err != nil {
		return nil, err
	}

	meminfo, err := fs.Meminfo()
	if err != nil {
		return nil, err
	}
	totalMemory := *meminfo.MemTotal
	freeMemory := *meminfo.MemFree + *meminfo.Buffers + *meminfo.Cached
	usedMemory := totalMemory - freeMemory
	memoryUsage := float64(usedMemory) / float64(totalMemory)
	memoryFree := float64(freeMemory) / float64(totalMemory)
	nodeInfoResult.NodeInfo = map[string]string{"memoryUsage": fmt.Sprintf("%.2f", memoryUsage*100), "memoryIdle": fmt.Sprintf("%.2f", memoryFree*100)}
	avg, err := fs.LoadAvg()
	if err != nil {
		klog.Errorf(" failed to get loadavg,err:%s", err)
	} else {
		nodeInfoResult.NodeInfo["load1"] = fmt.Sprintf("%.2f", avg.Load1)
		nodeInfoResult.NodeInfo["load5"] = fmt.Sprintf("%.2f", avg.Load5)
		nodeInfoResult.NodeInfo["load15"] = fmt.Sprintf("%.2f", avg.Load15)
	}

	stat, err := fs.Stat()
	if err != nil {
		klog.Error(err)
	} else {
		totalUsage := 0.0
		totalIdle := 0.0
		for _, cpuStat := range stat.CPU {
			totalUsage += cpuStat.System + cpuStat.User + cpuStat.Nice
			totalIdle += cpuStat.Idle
		}

		usage := totalUsage / (totalUsage + totalIdle)
		idle := totalIdle / (totalUsage + totalIdle)
		nodeInfoResult.NodeInfo["cpuUsage"] = fmt.Sprintf("%.2f", usage*100)
		nodeInfoResult.NodeInfo["cpuIdle"] = fmt.Sprintf("%.2f", idle*100)
	}
	_, exist, phase := utils.ArrayFinds(task.Status.JobPhase, func(m kubeeyev1alpha2.JobPhase) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var sysctl []kubeeyev1alpha2.SysRule
		err := json.Unmarshal(phase.RunRule, &sysctl)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		for _, sysRule := range sysctl {
			ctlRule, err := fs.SysctlStrings(sysRule.Name)
			klog.Infof("name:%s,value:%s", sysRule.Name, ctlRule)
			var ctl kubeeyev1alpha2.NodeResultItem
			ctl.Name = sysRule.Name
			if err != nil {
				errVal := fmt.Sprintf("name:%s to does not exist", sysRule.Name)
				ctl.Value = &errVal
			} else {
				val := strings.Join(ctlRule, ",")
				ctl.Value = &val
				if sysRule.Rule != nil {
					if _, err := visitor.CheckRule(*sysRule.Rule); err != nil {
						sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
						klog.Error(sprintf)
						ctl.Value = &sprintf
					} else {
						err, res := visitor.EventRuleEvaluate(map[string]interface{}{sysRule.Name: ctlRule[0]}, *sysRule.Rule)
						if err != nil {
							sprintf := fmt.Sprintf("err:%s", err.Error())
							klog.Error(sprintf)
							ctl.Value = &sprintf
						} else {
							ctl.Assert = &res
						}

					}

				}
			}
			nodeInfoResult.SysctlResult = append(nodeInfoResult.SysctlResult, ctl)
		}

	}

	marshal, err := json.Marshal(nodeInfoResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *sysctlInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	runNodeName := findJobRunNode(ctx, jobs, c)
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Namespace: task.Namespace,
		Name:      fmt.Sprintf("%s-nodeinfo", task.Name),
	}, &inspectResult)
	var nodeInfoResult kubeeyev1alpha2.NodeInfoResult
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &nodeInfoResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
	}
	if err != nil {
		if kubeErr.IsNotFound(err) {
			var ownerRefBol = true
			resultRef := metav1.OwnerReference{
				APIVersion:         task.APIVersion,
				Kind:               task.Kind,
				Name:               task.Name,
				UID:                task.UID,
				Controller:         &ownerRefBol,
				BlockOwnerDeletion: &ownerRefBol,
			}
			inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
			inspectResult.Name = fmt.Sprintf("%s-nodeinfo", task.Name)
			inspectResult.Namespace = task.Namespace
			inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
			inspectResult.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{runNodeName: nodeInfoResult}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	infoResult, ok := inspectResult.Spec.NodeInfoResult[runNodeName]
	if ok {
		infoResult.NodeInfo = mergeMap(infoResult.NodeInfo, nodeInfoResult.NodeInfo)
		//infoResult.FileChangeResult = append(infoResult.FileChangeResult, nodeInfoResult.FileChangeResult...)
		infoResult.SysctlResult = append(infoResult.SysctlResult, nodeInfoResult.SysctlResult...)
		//infoResult.SystemdResult = append(infoResult.SystemdResult, nodeInfoResult.SystemdResult...)
	} else {
		infoResult = nodeInfoResult
	}

	inspectResult.Spec.NodeInfoResult[runNodeName] = infoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}

func mergeMap(map1 map[string]string, map2 map[string]string) map[string]string {
	if map1 == nil {
		return map2
	}
	for k, v := range map2 {
		map1[k] = v
	}
	return map1
}

func mergeSysRule(rule []kubeeyev1alpha2.SysRule, types mergeType) (map[string][]kubeeyev1alpha2.SysRule, bool) {
	var mergeNodeMap = make(map[string][]kubeeyev1alpha2.SysRule)
	exists := false
	switch types {
	case nodeName:
		for _, changeRule := range rule {
			mergeNodeMap[*changeRule.NodeName] = append(mergeNodeMap[*changeRule.NodeName], changeRule)
			exists = true
		}
		break
	case nodeSelector:
		for _, changeRule := range rule {
			formatLabels := labels.FormatLabels(changeRule.NodeSelector)
			mergeNodeMap[formatLabels] = append(mergeNodeMap[formatLabels], changeRule)
		}
		break
	}

	return mergeNodeMap, exists
}
