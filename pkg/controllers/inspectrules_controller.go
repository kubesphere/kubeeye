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
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/utils"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// InspectRulesReconciler reconciles a Insights object
type InspectRulesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Insights object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *InspectRulesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	inspectRules := &kubeeyev1alpha2.InspectRule{}

	err := r.Get(ctx, req.NamespacedName, inspectRules)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Error(err, "failed to get inspect ruleFiles")
		return ctrl.Result{}, err
	}

	if inspectRules.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, inspectRules.Finalizers); !b {
			inspectRules.Finalizers = append(inspectRules.Finalizers, Finalizers)
			err = r.Client.Update(ctx, inspectRules)
			if err != nil {
				klog.Info("Failed to inspect plan add finalizers", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, inspectRules.Finalizers)
		inspectRules.Finalizers = newFinalizers.([]string)
		klog.Info("inspect ruleFiles is being deleted")
		err = r.Client.Update(ctx, inspectRules)
		if err != nil {
			klog.Info("Failed to inspect plan add finalizers")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if inspectRules.Status.State == "" {
		inspectRules.Status.State = kubeeyev1alpha2.StartImport
		inspectRules.Status.StartImportTime = &v1.Time{Time: time.Now()}
		err = r.Status().Update(ctx, inspectRules)
		if err != nil {
			klog.Error(err, "failed to update inspect ruleFiles")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if inspectRules.Status.State == kubeeyev1alpha2.ImportComplete {
		return ctrl.Result{}, nil
	}

	var levelCount = make(map[kubeeyev1alpha2.Level]*int)

	if inspectRules.Spec.Opas != nil {
		ComputeLevel(inspectRules.Spec.Opas, levelCount)
	}
	if inspectRules.Spec.Prometheus != nil {
		ComputeLevel(inspectRules.Spec.Prometheus, levelCount)
	}
	if inspectRules.Spec.FileChange != nil {
		ComputeLevel(inspectRules.Spec.FileChange, levelCount)
	}
	if inspectRules.Spec.Sysctl != nil {
		ComputeLevel(inspectRules.Spec.Sysctl, levelCount)
	}
	if inspectRules.Spec.Systemd != nil {
		ComputeLevel(inspectRules.Spec.Systemd, levelCount)
	}
	if inspectRules.Spec.FileFilter != nil {
		ComputeLevel(inspectRules.Spec.FileFilter, levelCount)
	}
	if inspectRules.Spec.CustomCommand != nil {
		ComputeLevel(inspectRules.Spec.CustomCommand, levelCount)
	}
	if inspectRules.Spec.NodeInfo != nil {
		ComputeLevel(inspectRules.Spec.NodeInfo, levelCount)
	}
	if inspectRules.Spec.ServiceConnect != nil {
		ComputeLevel(inspectRules.Spec.ServiceConnect, levelCount)
	}

	inspectRules.Status.EndImportTime = &v1.Time{Time: time.Now()}
	inspectRules.Status.State = kubeeyev1alpha2.ImportComplete
	inspectRules.Status.LevelCount = levelCount
	err = r.Status().Update(ctx, inspectRules)
	if err != nil {
		klog.Error(err, "failed to update inspect ruleFiles")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectRule{}).
		Complete(r)
}

func ComputeLevel(data interface{}, mapLevel map[kubeeyev1alpha2.Level]*int) {

	maps, err := utils.ArrayStructToArrayMap(data)
	if err != nil {
		return
	}
	Autoincrement := func(level kubeeyev1alpha2.Level) *int {
		if mapLevel[level] == nil {
			mapLevel[level] = new(int)
		}
		*mapLevel[level]++
		return mapLevel[level]
	}
	for _, m := range maps {
		v, ok := m["level"]
		if !ok {
			mapLevel[kubeeyev1alpha2.DangerLevel] = Autoincrement(kubeeyev1alpha2.DangerLevel)
		} else {
			l := v.(string)
			mapLevel[kubeeyev1alpha2.Level(l)] = Autoincrement(kubeeyev1alpha2.Level(l))
		}

	}

}
