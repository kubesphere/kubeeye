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
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// InspectRulesReconciler reconciles a Insights object
type InspectRulesReconciler struct {
	client.Client
	k8sClient kube.KubernetesClient
	Scheme    *runtime.Scheme
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

	inspectRules := &kubeeyev1alpha2.InspectRules{}

	err := r.Get(ctx, req.NamespacedName, inspectRules)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			fmt.Printf("inspect ruleFiles is not found;name:%s,namespect:%s\n", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		controller_log.Error(err, "failed to get inspect ruleFiles")
		return ctrl.Result{}, err
	}

	if inspectRules.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, inspectRules.Finalizers); !b {
			inspectRules.Finalizers = append(inspectRules.Finalizers, Finalizers)
			err = r.Client.Update(ctx, inspectRules)
			if err != nil {
				controller_log.Info("Failed to inspect plan add finalizers")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, inspectRules.Finalizers)
		inspectRules.Finalizers = newFinalizers.([]string)
		controller_log.Info("inspect ruleFiles is being deleted")
		err = r.Client.Update(ctx, inspectRules)
		if err != nil {
			controller_log.Info("Failed to inspect plan add finalizers")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if inspectRules.Status.State == "" {
		inspectRules.Status.State = kubeeyev1alpha2.StartImport
		err = r.Status().Update(ctx, inspectRules)
		if err != nil {
			controller_log.Error(err, "failed to update inspect ruleFiles")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if inspectRules.Status.State == kubeeyev1alpha2.ImportSuccess {
		controller_log.Info("import inspect ruleFiles success")
		return ctrl.Result{}, nil
	}
	controller_log.Info("starting inspect ruleFiles")
	copyInspectRules := inspectRules.DeepCopy()

	total := 0
	if inspectRules.Spec.Opas != nil {
		total += len(*inspectRules.Spec.Opas)
	}
	if inspectRules.Spec.Prometheus != nil {
		total += len(*inspectRules.Spec.Prometheus)
	}
	if inspectRules.Spec.FileChange != nil {
		total += len(*inspectRules.Spec.FileChange)
	}

	copyInspectRules.Status.ImportTime = v1.Time{Time: time.Now()}
	copyInspectRules.Status.State = kubeeyev1alpha2.ImportSuccess
	copyInspectRules.Status.RuleCount = total
	err = r.Status().Update(ctx, copyInspectRules)
	if err != nil {
		controller_log.Error(err, "failed to update inspect ruleFiles")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectRules{}).
		Complete(r)
}
