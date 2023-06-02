package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/kubesphere/event-rule-engine/visitor"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type systemdInspect struct {
}

func init() {
	RuleOperatorMap[constant.Systemd] = &systemdInspect{}
}

func (o *systemdInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	jobName := fmt.Sprintf("%s-%s", task.Name, constant.Systemd)

	var sysRules []kubeeyev1alpha2.SysRule

	_ = json.Unmarshal(task.Spec.Rules[constant.Systemd], &sysRules)

	nodeData, filterData := utils.ArrayFilter(sysRules, func(v kubeeyev1alpha2.SysRule) bool {
		return v.NodeName != nil
	})

	nodeNameRule, nodeNameStatus := mergeSysRule(nodeData, nodeName)
	if nodeNameStatus {
		for key, v := range nodeNameRule {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, v[0].Name), task, key, nil, constant.Systemd)
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
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, NodeName: key, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}

	}

	nodeSelectorData, residueData := utils.ArrayFilter(filterData, func(v kubeeyev1alpha2.SysRule) bool {
		return v.NodeSelector != nil
	})
	nodeSelectorRule, nodeSelectorStatus := mergeSysRule(nodeSelectorData, nodeSelector)
	if nodeSelectorStatus {
		for k, v := range nodeSelectorRule {
			labelsMap, _ := labels.ConvertSelectorToLabelsMap(k)
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, k), task, "", labelsMap, constant.Systemd)
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
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: createJob.Name, NodeName: k, RunRule: marshal, Phase: kubeeyev1alpha2.PhaseRunning})
		}
	}

	if len(residueData) > 0 {
		nodeAll, err := clients.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, nodeItem := range nodeAll.Items {
			job, err := template.InspectJobsTemplate(fmt.Sprintf("%s-%s", jobName, nodeItem.Name), task, nodeItem.Name, nil, constant.Systemd)
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

func (o *systemdInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	var nodeResult []kubeeyev1alpha2.NodeResultItem

	_, exist, phase := utils.ArrayFinds(task.Status.JobPhase, func(m kubeeyev1alpha2.JobPhase) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var systemd []kubeeyev1alpha2.SysRule
		err := json.Unmarshal(phase.RunRule, &systemd)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		conn, err := dbus.NewWithContext(ctx)
		if err != nil {
			return nil, err
		}
		unitsContext, err := conn.ListUnitsContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range systemd {
			var ctl kubeeyev1alpha2.NodeResultItem
			ctl.Name = r.Name
			for _, status := range unitsContext {
				if status.Name == fmt.Sprintf("%s.service", r.Name) {
					ctl.Value = &status.ActiveState

					if r.Rule != nil {

						if _, err := visitor.CheckRule(*r.Rule); err != nil {
							sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
							klog.Error(sprintf)
							ctl.Value = &sprintf
						} else {
							err, res := visitor.EventRuleEvaluate(map[string]interface{}{r.Name: status.ActiveState}, *r.Rule)
							if err != nil {
								sprintf := fmt.Sprintf("err:%s", err.Error())
								klog.Error(sprintf)
								ctl.Value = &sprintf
							} else {
								ctl.Assert = &res
							}
						}
					}
					break
				}
			}
			if ctl.Value == nil {
				errVal := fmt.Sprintf("name:%s to does not exist", r.Name)
				ctl.Value = &errVal
			}
			nodeResult = append(nodeResult, ctl)
		}
	}

	marshal, err := json.Marshal(nodeResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil

}

func (o *systemdInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Namespace: task.Namespace,
		Name:      fmt.Sprintf("%s-nodeinfo", task.Name),
	}, &inspectResult)
	var nodeInfoResult []kubeeyev1alpha2.NodeResultItem
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &nodeInfoResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
		return err
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
			inspectResult.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{jobs.Spec.Template.Spec.NodeName: {SystemdResult: nodeInfoResult}}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	infoResult, ok := inspectResult.Spec.NodeInfoResult[jobs.Spec.Template.Spec.NodeName]
	if ok {
		infoResult.SystemdResult = append(infoResult.SystemdResult, nodeInfoResult...)
	} else {
		infoResult.SystemdResult = nodeInfoResult
	}

	inspectResult.Spec.NodeInfoResult[jobs.Spec.Template.Spec.NodeName] = infoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}
