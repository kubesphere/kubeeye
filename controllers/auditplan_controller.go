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
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
)

// AuditPlanReconciler reconciles a AuditPlan object
type AuditPlanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=auditplans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=auditplans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=auditplans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuditPlan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AuditPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName(req.NamespacedName.String())

	auditPlan := &kubeeyev1alpha2.AuditPlan{}
	err := r.Get(ctx, req.NamespacedName, auditPlan)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			logger.Error(err, "audit plan is not found")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get Audit plan")
		return ctrl.Result{}, err
	}

	if !auditPlan.DeletionTimestamp.IsZero() {
		logger.Info("audit plan is being deleted")
		return ctrl.Result{}, nil
	}

	if auditPlan.Spec.Suspend {
		logger.Info("audit plan suspend")
		return ctrl.Result{}, nil
	}

	schedule, err := cron.ParseStandard(auditPlan.Spec.Schedule)
	if err != nil {
		logger.Error(err, "Unparseable schedule")
		return ctrl.Result{}, nil
	}

	if auditPlan.Spec.Timeout == "" {
		auditPlan.Spec.Timeout = "10m"
	}

	var scheduleStart time.Time
	// if the auditplan have just been created
	// we use its creation time as the start time to determine the next schedule
	if auditPlan.Status.LastScheduleTime.IsZero() {
		scheduleStart = auditPlan.CreationTimestamp.Time
	} else {
		scheduleStart = auditPlan.Status.LastScheduleTime.Time
	}

	missedCount, lastMissedScheduleTime, nextScheduleTime := getNextScheduleTime(schedule, scheduleStart)

	// no missed schedule, wait for the next schedule time to come
	if missedCount == 0 || lastMissedScheduleTime.IsZero() {
		return ctrl.Result{RequeueAfter: time.Until(nextScheduleTime)}, nil
	}

	if lastMissedScheduleTime.Add(time.Minute).Before(time.Now()) {
		logger.Info("last schedule time is missed for too long, skip and wait for the next schedule", "missedCount", missedCount)
		return ctrl.Result{RequeueAfter: time.Until(nextScheduleTime)}, nil
	}

	// the last missed schedule is right on time
	taskName, err := r.createAuditTask(auditPlan, ctx)
	if err != nil {
		logger.Error(err, "failed to create Audit task")
		return ctrl.Result{}, err
	}
	logger.Info("create a new audit task ", "task name", taskName)
	auditPlan.Status.LastScheduleTime = metav1.Time{Time: time.Now()}
	auditPlan.Status.LastTaskName = taskName
	err = r.Status().Update(ctx, auditPlan)
	if err != nil {
		logger.Error(err, "failed to update audit plan")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuditPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.AuditPlan{}).
		Complete(r)
}

// nextScheduledTimeDuration returns the time duration to requeue based on
// the schedule and given time. It adds a 100ms padding to the next requeue to account
// for Network Time Protocol(NTP) time skews.
// if controller has
func getNextScheduleTime(sched cron.Schedule, scheduleStart time.Time) (missedCount int, lastMissedScheduleTime time.Time, nextScheduleTime time.Time) {
	now := time.Now()
	for t := sched.Next(scheduleStart); !t.After(now); t = sched.Next(t) {
		lastMissedScheduleTime = t
		missedCount++
	}
	return missedCount, lastMissedScheduleTime, sched.Next(now).Add(100 * time.Millisecond)
}

func (r *AuditPlanReconciler) createAuditTask(auditPlan *kubeeyev1alpha2.AuditPlan, ctx context.Context) (string, error) {
	var auditTask kubeeyev1alpha2.AuditTask
	auditTask.Name = fmt.Sprintf("%s-%s", auditPlan.Name, strconv.Itoa(int(time.Now().Unix())))
	auditTask.Namespace = auditPlan.Namespace
	auditTask.Labels = auditPlan.Labels
	auditTask.Spec.Auditors = auditPlan.Spec.Auditors
	auditTask.Spec.Timeout = auditPlan.Spec.Timeout
	err := r.Client.Create(ctx, &auditTask)
	if err != nil {
		return "", err
	}
	return auditTask.Name, nil
}
