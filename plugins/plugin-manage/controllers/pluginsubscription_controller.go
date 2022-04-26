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
	"github.com/kubesphere/kubeeye/pkg/expend"
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/plugins/plugin-manage/api/v1alpha1"
	"github.com/kubesphere/kubeeye/plugins/plugin-manage/pkg"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PluginSubscriptionReconciler reconciles a PluginSubscription object
type PluginSubscriptionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=pluginsubscriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=pluginsubscriptions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=pluginsubscriptions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PluginSubscription object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PluginSubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logs := log.FromContext(ctx)

	// TODO(user): your logic here
	pluginSub := &kubeeyev1alpha1.PluginSubscription{}
	err := r.Get(ctx, req.NamespacedName, pluginSub)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			logs.Info("Cluster resource not found. Ignoring since object must be deleted ", "name", req.String())
			return ctrl.Result{}, nil
		}
		logs.Error(err, "Get pluginSub object failed for ", req.String())
		return ctrl.Result{}, err
	}

	finalizer := "kubeeye.kubesphere.io/plugin"
	if !pluginSub.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted, uninstall plugin
		pluginSub.Status.Install = pkg.PluginUninstalling
		if err := r.Status().Update(ctx, pluginSub); err != nil {
			if kubeErr.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				logs.Error(err, "unexpected error when update status")
				return ctrl.Result{}, err
			}
		}

		if err := expend.InstallOrUninstallPlugin(ctx, pluginSub.GetNamespace(), pluginSub.GetName(), false); err != nil {
			return ctrl.Result{}, err
		}
		pluginSub.Status.Install = pkg.PluginUninstalled
		if err := r.Status().Update(ctx, pluginSub); err != nil {
			if kubeErr.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				logs.Error(err, "unexpected error when update status")
				return ctrl.Result{}, err
			}
		}
		logs.Info("plugin uninstalled successfully")
		pluginSub.ObjectMeta.Finalizers = removeString(pluginSub.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, pluginSub); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object.
	//append finalizer
	if !containsString(pluginSub.ObjectMeta.Finalizers, finalizer) {
		pluginSub.ObjectMeta.Finalizers = append(pluginSub.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, pluginSub); err != nil {
			return ctrl.Result{}, err
		}
	}
	// if plugin is not installed, install it
	if pluginSub.Status.Install == pkg.PluginUninstalled || pluginSub.Status.Install == "" {
		// update status to installing
		pluginSub.Status.Install = pkg.PluginInstalling
		if err := r.Status().Update(ctx, pluginSub); err != nil {
			if kubeErr.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				logs.Error(err, "unexpected error when update status")
				return ctrl.Result{}, err
			}
		}

		if err := expend.InstallOrUninstallPlugin(ctx, pluginSub.GetNamespace(), pluginSub.GetName(), true); err != nil {
			logs.Error(err, "plugin installed failed")
			return ctrl.Result{}, err
		}

		// update plugin installed status
		go func(pluginSub *kubeeyev1alpha1.PluginSubscription) {
			if expend.IsPluginPodRunning(pluginSub.GetNamespace(), pluginSub.GetName()) {
				pluginSub.Status.Install = pkg.PluginIntalled
				pluginSub.Status.Enabled = pluginSub.Spec.Enabled
				if err := r.Status().Update(ctx, pluginSub); err != nil {
					if kubeErr.IsConflict(err) {
						logs.Error(err, "IsConflict")
						return
					} else {
						logs.Error(err, "unexpected error when update status")
						return
					}
				}
				logs.Info("plugin installed successfully.")
			}
		}(pluginSub)
	}

	if pluginSub.Status.Enabled != pluginSub.Spec.Enabled {
		pluginSub.Status.Enabled = pluginSub.Spec.Enabled
		if err := r.Status().Update(ctx, pluginSub); err != nil {
			if kubeErr.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			} else {
				logs.Error(err, "unexpected error when update status")
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PluginSubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: pkg.MaxConcurrentReconciles,
		}).
		For(&kubeeyev1alpha1.PluginSubscription{}).
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
