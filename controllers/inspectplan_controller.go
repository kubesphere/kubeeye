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
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"strconv"
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

var controller_log = ctrl.Log.WithName("controller_log")

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
			controller_log.Error(err, "inspect plan is not found")
			return ctrl.Result{}, nil
		}
		controller_log.Error(err, "failed to get Audit plan")
		return ctrl.Result{}, err
	}

	if !inspectPlan.DeletionTimestamp.IsZero() {
		controller_log.Info("audit plan is being deleted")
		return ctrl.Result{}, nil
	}

	if inspectPlan.Spec.Suspend {
		controller_log.Info("audit plan suspend")
		return ctrl.Result{}, nil
	}

	schedule, err := cron.ParseStandard(inspectPlan.Spec.Schedule)
	if err != nil {
		controller_log.Error(err, "Unparseable schedule")
		return ctrl.Result{}, nil
	}

	if inspectPlan.Spec.Timeout == "" {
		inspectPlan.Spec.Timeout = "10m"
	}

	now := time.Now()
	scheduledTime := nextScheduledTimeDuration(schedule, inspectPlan.Status.LastScheduleTime.Time)
	if inspectPlan.Status.LastScheduleTime.Add(*scheduledTime).Before(now) { // if the scheduled time has arrived, create Audit task

		taskName, err := r.createInspectTask(inspectPlan, ctx)
		if err != nil {
			controller_log.Error(err, "failed to create Inspect task")
			return ctrl.Result{}, err
		}
		controller_log.Info("create a new inspect task ", "task name", taskName)
		inspectPlan.Status.LastScheduleTime = metav1.Time{Time: now}
		inspectPlan.Status.LastTaskName = taskName
		inspectPlan.Status.NextScheduleTime = metav1.Time{Time: schedule.Next(now)}
		err = r.Status().Update(ctx, inspectPlan)
		if err != nil {
			controller_log.Error(err, "failed to update audit plan")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
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
	var inspectTask kubeeyev1alpha2.InspectTask
	rules, err := r.scanRules(inspectPlan, ctx)
	if err != nil {
		return "", err
	}
	inspectTask.Name = fmt.Sprintf("%s-%s", inspectPlan.Name, strconv.Itoa(int(time.Now().Unix())))
	inspectTask.Namespace = inspectPlan.Namespace
	inspectTask.Labels = inspectPlan.Labels
	inspectTask.Spec.Auditors = inspectPlan.Spec.Auditors
	inspectTask.Spec.Timeout = inspectPlan.Spec.Timeout
	inspectTask.Spec.Rules = rules
	err = r.Client.Create(ctx, &inspectTask)
	if err != nil {
		return "", err
	}
	return inspectTask.Name, nil
}

func (r *InspectPlanReconciler) scanRules(inspectPlan *kubeeyev1alpha2.InspectPlan, ctx context.Context) ([]map[string]string, error) {
	if len(inspectPlan.Spec.Tags) == 0 && len(inspectPlan.Spec.RuleNames) == 0 {
		return nil, errors.New("Failed to get tags and rule names")
	}

	list, err := r.K8sClient.VersionClientSet.KubeeyeV1alpha2().InspectRules(v1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		if kubeErr.IsNotFound(err) {
			controller_log.Error(err, "failed get to inspectrules not found")
			return nil, err
		}
		controller_log.Error(err, "failed get to inspectrules")
		return nil, err
	}
	var resultRules []map[string]string

	for _, item := range list.Items {
		for _, rules := range item.Spec.Rules {
			for _, tag := range inspectPlan.Spec.Tags {
				if utils.ArrayFind(tag, rules.Tags) && (rules.Opa != "" || rules.Prometheus != "") {
					ruleType := constant.Prometheus
					rule := rules.Prometheus
					if rules.Opa != "" {
						ruleType = constant.Opa
						rule = rules.Opa
					}
					resultRules = append(resultRules, map[string]string{constant.RuleType: ruleType, constant.Rules: rule})
				}
			}

		}
	}

	return resultRules, nil
}
