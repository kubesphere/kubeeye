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
	"sync"
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
	Instance  sync.Map
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
	// Every time a plan operation is triggered, it checks how many plans are associated with the rule

	if err != nil {
		if kubeErr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Error("failed to get inspect plan.\n", err)
		return ctrl.Result{}, err
	}

	if inspectPlan.DeletionTimestamp.IsZero() {
		if _, ok := utils.ArrayFind(Finalizers, inspectPlan.Finalizers); !ok {
			inspectPlan.Finalizers = append(inspectPlan.Finalizers, Finalizers)
			inspectPlan.Annotations = utils.MergeMap(inspectPlan.Annotations, map[string]string{constant.AnnotationJoinRuleNum: strconv.Itoa(len(inspectPlan.Spec.RuleNames))})
			r.updateAddRuleReferNum(ctx, inspectPlan)
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
		r.updateSubRuleReferNum(ctx, inspectPlan)
		err = r.Client.Update(ctx, inspectPlan)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers.\n", err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if inspectPlan.Spec.Once != nil {
		if !utils.IsEmptyString(inspectPlan.Status.LastTaskName) {
			return ctrl.Result{}, nil
		}
		if !inspectPlan.Spec.Once.After(time.Now()) {
			taskName, err := r.createInspectTask(inspectPlan, ctx)
			if err != nil {
				klog.Error("failed to create InspectTask.", err)
				return ctrl.Result{}, err
			}

			if err = r.updateStatus(ctx, inspectPlan, time.Now(), taskName); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		nextScheduledTime := inspectPlan.Spec.Once.Sub(time.Now())
		return ctrl.Result{RequeueAfter: nextScheduledTime}, nil
	}

	if inspectPlan.Spec.Schedule == nil {
		if !utils.IsEmptyString(inspectPlan.Status.LastTaskName) {
			return ctrl.Result{}, nil
		}
		taskName, err := r.createInspectTask(inspectPlan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}
		if err = r.updateStatus(ctx, inspectPlan, time.Now(), taskName); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	if inspectPlan.Spec.Suspend {
		klog.Info("inspect plan suspend")
		return ctrl.Result{}, nil
	}
	schedule, err := cron.ParseStandard(*inspectPlan.Spec.Schedule)
	if err != nil {
		klog.Error("Unparseable schedule.\n", err)
		return ctrl.Result{}, nil
	}
	now := time.Now()
	scheduledTime := nextScheduledTimeDuration(schedule, inspectPlan.Status.LastScheduleTime.Time)
	if inspectPlan.Status.LastScheduleTime.Add(*scheduledTime).Before(now) {
		taskName, err := r.createInspectTask(inspectPlan, ctx)
		if err != nil {
			klog.Error("failed to create InspectTask.", err)
			return ctrl.Result{}, err
		}

		inspectPlan.Status.NextScheduleTime = &metav1.Time{Time: schedule.Next(now)}
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
	r.removeTask(ctx, inspectPlan)
	inspectTask := kubeeyev1alpha2.InspectTask{
		ObjectMeta: metav1.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", inspectPlan.Name, time.Now().Format("20060102-15-04")),
			Labels: map[string]string{constant.LabelPlanName: inspectPlan.Name},
			Annotations: map[string]string{constant.AnnotationInspectType: func() string {
				if inspectPlan.Spec.Schedule == nil {
					return string(kubeeyev1alpha2.InspectTypeInstant)
				}
				return string(kubeeyev1alpha2.InspectTypeTiming)
			}()},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         inspectPlan.APIVersion,
				Kind:               inspectPlan.Kind,
				Name:               inspectPlan.Name,
				UID:                inspectPlan.UID,
				Controller:         &ownerController,
				BlockOwnerDeletion: &ownerController,
			}},
		},
		Spec: kubeeyev1alpha2.InspectTaskSpec{
			RuleNames:   inspectPlan.Spec.RuleNames,
			ClusterName: inspectPlan.Spec.ClusterName,
			Timeout: func() string {
				if inspectPlan.Spec.Timeout == "" {
					return "10m"
				}
				return inspectPlan.Spec.Timeout
			}(),
			InspectPolicy: func() kubeeyev1alpha2.Policy {
				if inspectPlan.Spec.Once == nil && inspectPlan.Spec.Schedule != nil {
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

func (r *InspectPlanReconciler) updateAddRuleReferNum(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan) {

	for _, v := range plan.Spec.RuleNames {
		rule, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(ctx, v, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "Failed to get inspectRules")
			continue
		}
		rule.Labels = utils.MergeMap(rule.Labels, map[string]string{fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", plan.Name): plan.Name})
		num := 0
		n, ok := rule.Annotations[constant.AnnotationJoinPlanNum]
		if ok {
			num, _ = strconv.Atoi(n)
			num++
		}
		rule.Annotations = utils.MergeMap(rule.Annotations, map[string]string{constant.AnnotationJoinPlanNum: strconv.Itoa(num)})

		_, err = r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Update(ctx, rule, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err, "Failed to update inspectRules")
			continue
		}
		plan.Labels = utils.MergeMap(plan.Labels, map[string]string{fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", v): v})

	}

}

func (r *InspectPlanReconciler) updateSubRuleReferNum(ctx context.Context, plan *kubeeyev1alpha2.InspectPlan) {

	for _, v := range plan.Spec.RuleNames {
		rule, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(ctx, v, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "Failed to get inspectRules")
			continue
		}
		delete(rule.Labels, fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", plan.Name))
		//rule.Labels = utils.MergeMap(rule.Labels, map[string]string{fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", plan.Name): plan.Name})
		num := 0
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
			continue
		}
		delete(plan.Labels, fmt.Sprintf("%s/%s", "kubeeye.kubesphere.io", v))

	}

}
