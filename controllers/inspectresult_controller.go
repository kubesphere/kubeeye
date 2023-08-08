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
	"encoding/json"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/utils"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"os"
	"path"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InspectResultReconciler reconciles a InspectResult object
type InspectResultReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectResult object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *InspectResultReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result := &kubeeyev1alpha2.InspectResult{}
	err := r.Get(ctx, req.NamespacedName, result)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			klog.Infof("inspect rule is not found;name:%s\n", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if result.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, result.Finalizers); !b {
			result.Finalizers = append(result.Finalizers, Finalizers)
			err = r.Client.Update(ctx, result)
			if err != nil {
				klog.Error("Failed to inspect plan add finalizers", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, result.Finalizers)
		result.Finalizers = newFinalizers.([]string)
		klog.Infof("inspect task is being deleted")
		err = os.Remove(path.Join(constant.ResultPath, result.Name))
		if err != nil {
			klog.Error(err, "failed to delete file")
		}
		err = r.Client.Update(ctx, result)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers. ", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if result.Status.Complete {
		return ctrl.Result{}, nil
	}

	taskName := result.GetLabels()[constant.LabelTaskName]

	task := &kubeeyev1alpha2.InspectTask{}
	err = r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: taskName}, task)
	if err != nil {
		klog.Error("Failed to get inspect task", err)
		return ctrl.Result{}, err
	}
	startTime := result.GetAnnotations()[constant.AnnotationStartTime]
	endTime := result.GetAnnotations()[constant.AnnotationEndTime]

	parseStart, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}
	parseEnd, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}

	result.Status.Policy = task.Spec.InspectPolicy
	result.Status.Duration = parseEnd.Sub(parseStart).String()
	result.Status.TaskStartTime = startTime
	result.Status.TaskEndTime = endTime
	result.Status.Complete = true
	countLevelNum, err := r.CountLevelNum(result.Name)
	if err != nil {
		klog.Error("Failed to count level num", err)
		return ctrl.Result{}, err
	}
	result.Status.Level = countLevelNum

	err = r.Client.Status().Update(ctx, result)
	if err != nil {
		klog.Error("Failed to update inspect result status", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectResultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectResult{}).
		Complete(r)
}

func (r *InspectResultReconciler) CountLevelNum(resultName string) (map[kubeeyev1alpha2.Level]*int, error) {
	file, err := os.ReadFile(resultName)
	if err != nil {
		return nil, err
	}

	var result kubeeyev1alpha2.InspectResult

	err = json.Unmarshal(file, &result)
	if err != nil {
		return nil, err
	}

	levelTotal := make(map[kubeeyev1alpha2.Level]*int)
	for _, v := range result.Spec.NodeInfoResult {
		for _, item := range v.FileChangeResult {
			computeLevel(item.Level, levelTotal)
		}
		for _, item := range v.FileFilterResult {
			computeLevel(item.Level, levelTotal)
		}
		for _, item := range v.SysctlResult {
			computeLevel(item.Level, levelTotal)
		}

		for _, item := range v.SystemdResult {
			computeLevel(item.Level, levelTotal)
		}
	}
	levelTotal[kubeeyev1alpha2.DangerLevel] = &result.Spec.OpaResult.Dangerous
	levelTotal[kubeeyev1alpha2.WarningLevel] = &result.Spec.OpaResult.Warning
	levelTotal[kubeeyev1alpha2.IgnoreLevel] = &result.Spec.OpaResult.Ignore
	for _, item := range result.Spec.PrometheusResult {
		for _, m := range item {
			computeLevel(kubeeyev1alpha2.Level(m["level"]), levelTotal)
		}
	}

	for _, item := range result.Spec.ComponentResult {
		computeLevel(item.Level, levelTotal)
	}
	return levelTotal, nil
}

func computeLevel(v kubeeyev1alpha2.Level, m map[kubeeyev1alpha2.Level]*int) {
	Autoincrement := func(level kubeeyev1alpha2.Level) *int {
		if m[level] == nil {
			m[level] = new(int)
		}
		*m[level]++
		return m[level]
	}
	if v == "" {
		m[kubeeyev1alpha2.DangerLevel] = Autoincrement(kubeeyev1alpha2.DangerLevel)
	} else {
		m[v] = Autoincrement(v)
	}

}
