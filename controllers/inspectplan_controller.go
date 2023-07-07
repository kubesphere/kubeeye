/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/rules"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InspectPlanReconciler reconciles a InspectPlan object
type InspectPlanReconciler struct {
	client.Client
	K8sClient *kube.KubernetesClient
	Scheme    *runtime.Scheme
}

const Finalizers = "kubeeye.finalizers.kubesphere.io"

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectplans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectplans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectplans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectPlan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *InspectPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	inspectPlan := &kubeeyev1alpha2.InspectPlan{}
	err := r.Get(ctx, req.NamespacedName, inspectPlan)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			klog.Infof("inspect plan is not found;name:%s,namespect:%s\n", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		klog.Error("failed to get inspect plan.\n", err)
		return ctrl.Result{}, err
	}

	if inspectPlan.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, inspectPlan.Finalizers); !b {
			inspectPlan.Finalizers = append(inspectPlan.Finalizers, Finalizers)
			err = r.Client.Update(ctx, inspectPlan)
			if err != nil {
				klog.Error("Failed to  add finalizers for inspect plan .\n", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, inspectPlan.Finalizers)
		inspectPlan.Finalizers = newFinalizers.([]string)
		klog.Info("inspect plan is being deleted.")
		err = r.Client.Update(ctx, inspectPlan)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers.\n", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if inspectPlan.Spec.Timeout == "" {
		inspectPlan.Spec.Timeout = "10m"
	}

	if inspectPlan.Spec.Schedule == nil {
		taskName, err := r.createInspectTask(inspectPlan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}
		klog.Info("create a new inspect task.", taskName)
		r.removeTask(ctx, inspectPlan)
		if err = r.updateStatus(ctx, inspectPlan, time.Now(), taskName); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	schedule, err := cron.ParseStandard(*inspectPlan.Spec.Schedule)
	if err != nil {
		klog.Error("Unparseable schedule.\n", err)
		return ctrl.Result{}, nil
	}
	if inspectPlan.Spec.Suspend {
		klog.Info("inspect plan suspend")
		return ctrl.Result{}, nil
	}
	now := time.Now()
	scheduledTime := nextScheduledTimeDuration(schedule, inspectPlan.Status.LastScheduleTime.Time)
	if inspectPlan.Status.LastScheduleTime.Add(*scheduledTime).Before(now) { // if the scheduled time has arrived, create Audit task
		taskName, err := r.createInspectTask(inspectPlan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}
		klog.Info("create a new inspect task.", taskName)
		r.removeTask(ctx, inspectPlan)
		inspectPlan.Status.NextScheduleTime = metav1.Time{Time: schedule.Next(now)}
		if err = r.updateStatus(ctx, inspectPlan, now, taskName); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	} else {
		nextScheduledTime := nextScheduledTimeDuration(schedule, now)
		return ctrl.Result{RequeueAfter: *nextScheduledTime}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectPlan{}).
		Complete(r)
}

// nextScheduledTimeDuration returns the time duration to requeue based on
// the schedule and given time. It adds a 100ms padding to the next requeue to account
// for Network Time Protocol(NTP) time skews.
func nextScheduledTimeDuration(sched cron.Schedule, now time.Time) *time.Duration {
	nextTime := sched.Next(now).Add(100 * time.Millisecond).Sub(now)
	return &nextTime
}

func (r *InspectPlanReconciler) createInspectTask(inspectPlan *kubeeyev1alpha2.InspectPlan, ctx context.Context) (string, error) {
	ownerController := true
	taskName := fmt.Sprintf("%s-%s", inspectPlan.Name, time.Now().Format("20060102-15-04"))
	scanRule, ruleTotal, err := r.scanRules(ctx, taskName, inspectPlan)
	if err != nil {
		return "", err
	}

	var inspectTask kubeeyev1alpha2.InspectTask
	inspectTask.Labels = map[string]string{constant.LabelName: inspectPlan.Name}
	inspectTask.OwnerReferences = []metav1.OwnerReference{{
		APIVersion:         inspectPlan.APIVersion,
		Kind:               inspectPlan.Kind,
		Name:               inspectPlan.Name,
		UID:                inspectPlan.UID,
		Controller:         &ownerController,
		BlockOwnerDeletion: &ownerController,
	}}
	inspectTask.Spec.Timeout = inspectPlan.Spec.Timeout
	inspectTask.Spec.Rules = scanRule
	inspectTask.Spec.InspectRuleTotal = ruleTotal
	inspectTask.Name = taskName
	err = r.Client.Create(ctx, &inspectTask)
	if err != nil {
		klog.Error("create inspectTask failed", err)
		return "", err
	}

	return inspectTask.Name, nil
}

func (r *InspectPlanReconciler) scanRules(ctx context.Context, taskName string, inspectPlan *kubeeyev1alpha2.InspectPlan) ([]kubeeyev1alpha2.JobRule, map[string]int, error) {
	if utils.IsEmptyString(inspectPlan.Spec.Tag) && inspectPlan.Spec.RuleNames == nil {
		return nil, nil, fmt.Errorf("failed to get tags and rule names")
	}

	ruleLists, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules(v1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{constant.LabelRuleTag: inspectPlan.Spec.Tag}),
	})
	if err != nil {
		if kubeErr.IsNotFound(err) {
			klog.Error("failed get to inspectrules not found.", err)
			return nil, nil, err
		}
		klog.Error("failed get to inspectrules.", err)
		return nil, nil, err
	}
	if ruleLists.Items == nil || len(ruleLists.Items) == 0 {
		klog.Errorf("failed to  rules not found to tag:%s , check whether it exists", inspectPlan.Spec.Tag)
		return nil, nil, fmt.Errorf("failed to  rules not found to tag:%s , check whether it exists", inspectPlan.Spec.Tag)
	}

	nodes := r.GetNodes(ctx)
	ruleSpec := rules.MergeRule(ruleLists.Items)

	var inspectRuleTotal = make(map[string]int)
	var executeRule []kubeeyev1alpha2.JobRule

	component := rules.AllocationComponent(ruleSpec.Component, taskName)
	executeRule = append(executeRule, *component)
	componentRuleNumber := 0
	if ruleSpec.Component == nil {
		inspectComponent, _ := inspect.GetInspectComponent(ctx, r.K8sClient, "")
		componentRuleNumber = len(inspectComponent)
	} else {
		componentRuleNumber = len(strings.Split(*ruleSpec.Component, "|"))
	}

	inspectRuleTotal[constant.Component] = componentRuleNumber
	opa := rules.AllocationOpa(ruleSpec.Opas, taskName)
	if opa != nil {
		executeRule = append(executeRule, *opa)
		inspectRuleTotal[constant.Opa] = len(ruleSpec.Opas)
	}
	prometheus := rules.AllocationPrometheus(ruleSpec.Prometheus, taskName)
	if prometheus != nil {
		executeRule = append(executeRule, *prometheus)
		inspectRuleTotal[constant.Prometheus] = len(ruleSpec.Prometheus)
	}
	if nodes != nil && len(nodes) > 0 {
		change := rules.AllocationRule(ruleSpec.FileChange, taskName, nodes, constant.FileChange)
		if change != nil && len(change) > 0 {
			executeRule = append(executeRule, change...)
			inspectRuleTotal[constant.FileChange] = len(ruleSpec.FileChange)
		}
		sysctl := rules.AllocationRule(ruleSpec.Sysctl, taskName, nodes, constant.Sysctl)
		if sysctl != nil && len(sysctl) > 0 {
			executeRule = append(executeRule, sysctl...)
			inspectRuleTotal[constant.Sysctl] = len(ruleSpec.Sysctl)
		}
		systemd := rules.AllocationRule(ruleSpec.Systemd, taskName, nodes, constant.Systemd)
		if systemd != nil && len(systemd) > 0 {
			executeRule = append(executeRule, systemd...)
			inspectRuleTotal[constant.Systemd] = len(ruleSpec.Systemd)
		}
		fileFilter := rules.AllocationRule(ruleSpec.FileFilter, taskName, nodes, constant.FileFilter)
		if fileFilter != nil && len(fileFilter) > 0 {
			executeRule = append(executeRule, fileFilter...)
			inspectRuleTotal[constant.FileFilter] = len(ruleSpec.FileFilter)
		}
	}
	return executeRule, inspectRuleTotal, nil
}

func (r *InspectPlanReconciler) GetNodes(ctx context.Context) []v1.Node {
	nodeAll, err := r.K8sClient.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("failed to get nodes", err)
		return nil
	}
	return nodeAll.Items
}

func (r *InspectPlanReconciler) removeTask(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan) {
	if plan.Spec.MaxTasks > 0 {
		for len(plan.Status.TaskNames) > plan.Spec.MaxTasks-1 {
			err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().RESTClient().Delete().Resource("inspecttasks").Name(plan.Status.TaskNames[0]).Do(ctx).Error()
			if err == nil || kubeErr.IsNotFound(err) {
				plan.Status.TaskNames = plan.Status.TaskNames[1:]
			} else {
				klog.Error("Failed to delete inspect task", err)
			}

		}
	}
}
func (r *InspectPlanReconciler) updateStatus(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan, now time.Time, taskName string) error {
	plan.Status.LastScheduleTime = metav1.Time{Time: now}
	plan.Status.LastTaskName = taskName
	plan.Status.TaskNames = append(plan.Status.TaskNames, taskName)
	err := r.Status().Update(ctx, plan)
	if err != nil {
		klog.Error("failed to update inspect plan.", err)
		return err
	}
	return nil
}
