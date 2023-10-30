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
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"sort"
	"strconv"
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

	plan := &kubeeyev1alpha2.InspectPlan{}
	err := r.Get(ctx, req.NamespacedName, plan)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Error("failed to get inspect plan.\n", err)
		return ctrl.Result{}, err
	}

	if plan.DeletionTimestamp.IsZero() {
		if _, ok := utils.ArrayFind(Finalizers, plan.Finalizers); !ok {
			plan.Finalizers = append(plan.Finalizers, Finalizers)
			err = r.Client.Update(ctx, plan)
			if err != nil {
				klog.Error("Failed to  add finalizers for inspect plan .\n", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		if d, action := r.GetUpdatePlanRule(plan); d != nil {
			plan.Annotations = utils.MergeMap(plan.Annotations, map[string]string{constant.AnnotationJoinRuleNum: strconv.Itoa(len(plan.Spec.RuleNames))})
			if action {
				r.updateAddRuleReferNum(ctx, d, plan)
			} else {
				r.updateSubRuleReferNum(ctx, d, plan)
			}

			err = r.Client.Update(ctx, plan)
			if err != nil {
				klog.Error("Failed to update rule refer quantity .\n", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, plan.Finalizers)
		plan.Finalizers = newFinalizers.([]string)
		klog.Info("inspect plan is being deleted.")
		r.updateSubRuleReferNum(ctx, plan.Spec.RuleNames, plan)
		err = r.Client.Update(ctx, plan)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers.\n", err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if plan.Status.LastTaskStatus.IsEmpty() {
		plan.Status.LastTaskStatus = kubeeyev1alpha2.PhasePending
		err = r.Status().Update(ctx, plan)
		if err != nil {
			klog.Error("failed to update InspectPlan  last task status.", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if plan.Spec.Suspend {
		klog.Info("inspect plan suspend")
		return ctrl.Result{}, nil
	}

	if plan.Spec.Once != nil {
		if !utils.IsEmptyValue(plan.Status.LastTaskName) {
			return ctrl.Result{}, nil
		}
		if !plan.Spec.Once.After(time.Now()) {
			taskName, err := r.createInspectTask(plan, ctx)
			if err != nil {
				klog.Error("failed to create InspectTask.", err)
				return ctrl.Result{}, err
			}

			if err = r.updateStatus(ctx, plan, time.Now(), taskName); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		//nextScheduledTime := inspectPlan.Spec.Once.Sub(time.Now())
		nextScheduledTime := time.Until(plan.Spec.Once.Time)

		return ctrl.Result{RequeueAfter: nextScheduledTime}, nil
	}

	if plan.Spec.Schedule == nil {
		if !utils.IsEmptyValue(plan.Status.LastTaskName) {
			return ctrl.Result{}, nil
		}
		taskName, err := r.createInspectTask(plan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}
		if err = r.updateStatus(ctx, plan, time.Now(), taskName); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	schedule, err := cron.ParseStandard(*plan.Spec.Schedule)
	if err != nil {
		klog.Error("Unparseable schedule.\n", err)
		return ctrl.Result{}, nil
	}
	now := time.Now()
	scheduledTime := nextScheduledTimeDuration(schedule, plan.Status.LastScheduleTime)
	if plan.Status.LastScheduleTime == nil || plan.Status.LastScheduleTime.Add(*scheduledTime).Before(now) {
		taskName, err := r.createInspectTask(plan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}

		plan.Status.NextScheduleTime = &metav1.Time{Time: schedule.Next(now)}
		if err = r.updateStatus(ctx, plan, now, taskName); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	} else {
		nextScheduledTime := nextScheduledTimeDuration(schedule, &metav1.Time{Time: now})
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
func nextScheduledTimeDuration(sched cron.Schedule, now *metav1.Time) *time.Duration {
	LastScheduleTime := time.Time{}
	if now != nil {
		LastScheduleTime = now.Time
	}
	nextTime := sched.Next(LastScheduleTime).Add(100 * time.Millisecond).Sub(LastScheduleTime)
	return &nextTime
}

func (r *InspectPlanReconciler) createInspectTask(plan *kubeeyev1alpha2.InspectPlan, ctx context.Context) (string, error) {
	ownerController := true
	r.removeTask(ctx, plan)
	inspectTask := kubeeyev1alpha2.InspectTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", plan.Name, time.Now().Format("20060102-15-04")),
			Labels: map[string]string{constant.LabelPlanName: plan.Name},
			Annotations: map[string]string{constant.AnnotationInspectType: func() string {
				if plan.Spec.Schedule == nil {
					return string(kubeeyev1alpha2.InspectTypeInstant)
				}
				return string(kubeeyev1alpha2.InspectTypeTiming)
			}()},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         plan.APIVersion,
				Kind:               plan.Kind,
				Name:               plan.Name,
				UID:                plan.UID,
				Controller:         &ownerController,
				BlockOwnerDeletion: &ownerController,
			}},
		},
		Spec: kubeeyev1alpha2.InspectTaskSpec{
			RuleNames:   plan.Spec.RuleNames,
			ClusterName: plan.Spec.ClusterName,
			Timeout: func() string {
				if plan.Spec.Timeout == "" {
					return "10m"
				}
				return plan.Spec.Timeout
			}(),
			InspectPolicy: func() kubeeyev1alpha2.Policy {
				if plan.Spec.Once == nil && plan.Spec.Schedule != nil {
					return kubeeyev1alpha2.CyclePolicy
				}
				return kubeeyev1alpha2.SinglePolicy
			}(),
		},
	}

	err := r.Client.Create(ctx, &inspectTask)
	if err != nil {
		return "", err
	}
	klog.Info("create a new inspect task.", inspectTask.Name)
	return inspectTask.Name, nil
}

func (r *InspectPlanReconciler) removeTask(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan) {
	if plan.Spec.MaxTasks > 0 {
		tasks, err := r.getInspectTaskForLabel(ctx, plan.Name)
		if err != nil {
			klog.Error("Failed to get inspect task for label", err)
		}
		if len(tasks) > plan.Spec.MaxTasks {
			for _, task := range tasks[:len(tasks)-plan.Spec.MaxTasks] {
				err = r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectTasks().Delete(ctx, task.Name, metav1.DeleteOptions{})
				if err == nil || kubeErr.IsNotFound(err) {
					klog.Info("delete inspect task", task.Name)
				} else {
					klog.Error("Failed to delete inspect task", err)
				}
			}
			plan.Status.TaskNames = ConvertTaskStatus(tasks[len(tasks)-plan.Spec.MaxTasks:])
		}

	}
}
func (r *InspectPlanReconciler) updateStatus(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan, now time.Time, taskName string) error {
	plan.Status.LastScheduleTime = &metav1.Time{Time: now}
	plan.Status.LastTaskName = taskName
	plan.Status.LastTaskStatus = kubeeyev1alpha2.PhasePending
	plan.Status.TaskNames = append(plan.Status.TaskNames, kubeeyev1alpha2.TaskNames{
		Name:       taskName,
		TaskStatus: kubeeyev1alpha2.PhasePending,
	})
	err := r.Status().Update(ctx, plan)
	if err != nil {
		klog.Error("failed to update inspect plan.", err)
		return err
	}
	return nil
}

func (r *InspectPlanReconciler) getInspectTaskForLabel(ctx context.Context, planName string) ([]kubeeyev1alpha2.InspectTask, error) {
	list, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectTasks().List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{constant.LabelPlanName: planName}),
	})

	if err != nil {
		return nil, err
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].CreationTimestamp.Before(&list.Items[j].CreationTimestamp)
	})
	return list.Items, nil
}

func ConvertTaskStatus(tasks []kubeeyev1alpha2.InspectTask) (taskStatus []kubeeyev1alpha2.TaskNames) {

	for _, t := range tasks {
		if t.Status.EndTimestamp.IsZero() {
			taskStatus = append(taskStatus, kubeeyev1alpha2.TaskNames{
				Name:       t.Name,
				TaskStatus: kubeeyev1alpha2.PhasePending,
			})
		} else {
			taskStatus = append(taskStatus, kubeeyev1alpha2.TaskNames{
				Name:       t.Name,
				TaskStatus: GetStatus(&t),
			})
		}

	}

	return

}

func (r *InspectPlanReconciler) updateAddRuleReferNum(ctx context.Context, ruleNames []kubeeyev1alpha2.InspectRuleNames, plan *kubeeyev1alpha2.InspectPlan) {

	for _, v := range ruleNames {
		rule, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(ctx, v.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "Failed to get inspectRules")
			continue
		}
		rule.Labels = utils.MergeMap(rule.Labels, map[string]string{fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", plan.Name): plan.Name})
		num := 1
		n, ok := rule.Annotations[constant.AnnotationJoinPlanNum]
		if ok {
			num, err = strconv.Atoi(n)
			if err != nil {
				klog.Error(err, "Failed to strconv.Atoi")
			} else {
				num++
			}

		}
		rule.Annotations = utils.MergeMap(rule.Annotations, map[string]string{constant.AnnotationJoinPlanNum: strconv.Itoa(num)})

		_, err = r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Update(ctx, rule, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err, "Failed to update inspectRules")
			continue
		}
		plan.Labels = utils.MergeMap(plan.Labels, map[string]string{fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", v.Name): v.Name})

	}

}

func (r *InspectPlanReconciler) updateSubRuleReferNum(ctx context.Context, ruleNames []kubeeyev1alpha2.InspectRuleNames, plan *kubeeyev1alpha2.InspectPlan) {

	for _, v := range ruleNames {
		rule, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(ctx, v.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "Failed to get inspectRules")
			continue
		}
		delete(rule.Labels, fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", plan.Name))
		num := 1
		n, ok := rule.Annotations[constant.AnnotationJoinPlanNum]
		if ok {
			num, _ = strconv.Atoi(n)
			if num > 0 {
				num--
			}
		}
		rule.Annotations = utils.MergeMap(rule.Annotations, map[string]string{constant.AnnotationJoinPlanNum: strconv.Itoa(num)})

		_, err = r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Update(ctx, rule, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err, "Failed to update inspectRules")
		}

	}

}

func (r *InspectPlanReconciler) GetUpdatePlanRule(plan *kubeeyev1alpha2.InspectPlan) ([]kubeeyev1alpha2.InspectRuleNames, bool) {
	var newLabel = make(map[string]string)
	var newRules []kubeeyev1alpha2.InspectRuleNames
	for key := range plan.Labels {
		if strings.HasPrefix(key, "kubeeye.kubesphere.io/") {
			newLabel[key] = plan.Labels[key]
		}
	}
	for _, n := range plan.Spec.RuleNames {
		if len(newLabel) > 0 {
			delete(newLabel, fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", n.Name))
		} else {
			newRules = append(newRules, kubeeyev1alpha2.InspectRuleNames{Name: n.Name})
		}
	}
	if len(newLabel) > 0 {
		for _, v := range newLabel {
			newRules = append(newRules, kubeeyev1alpha2.InspectRuleNames{Name: v})
		}
		return newRules, false
	}

	return newRules, true

}
