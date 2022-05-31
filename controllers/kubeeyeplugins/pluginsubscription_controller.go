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

package kubeeyeplugins

import (
	"context"
	"fmt"
	"time"

	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/expend"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeeyepluginsv1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeyeplugins/v1alpha1"
)

// PluginSubscriptionReconciler reconciles a PluginSubscription object
type PluginSubscriptionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubeeyeplugins.kubesphere.io,resources=pluginsubscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeeyeplugins.kubesphere.io,resources=pluginsubscriptions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeeyeplugins.kubesphere.io,resources=pluginsubscriptions/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups="",resources=namespaces;services;deployments;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete


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
	logs := log.FromContext(ctx).WithValues("pluginSubscription", req.NamespacedName)

	plugin := &kubeeyepluginsv1alpha1.PluginSubscription{}
	if err := r.Get(ctx, req.NamespacedName, plugin); err != nil {
		if kubeErr.IsNotFound(err) {
			logs.Info(fmt.Sprintf("plugin %s not found. Ignoring since object must be deleted", plugin.Name))
			return ctrl.Result{}, nil
		}
	}

	if plugin.Spec.Enabled && plugin.Status.State == conf.PluginInstalling {
		logs.Info(fmt.Sprintf("check plugin %s health", plugin.Name))
		state, err := expend.PluginHealth(plugin)
		if err != nil || state == "" {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, errors.Wrap(nil, fmt.Sprintf("plugin %s not ready", plugin.Name))
		}
		logs.Info(fmt.Sprintf("plugin %s installation complete", plugin.Name))
		plugin.Status.State = conf.PluginInstalled

	}

	if plugin.Spec.Enabled && (plugin.Status.State == "" || plugin.Status.State == conf.PluginUninstalled) {
		logs.Info(fmt.Sprintf("starting install plugin %s", plugin.Name))
		// get plugin configmap
		pluginConfigMap := &corev1.ConfigMap{}
		pluginNamespacedName := types.NamespacedName{Namespace: conf.KubeeyeNameSpace, Name: plugin.Name}
		if err := r.Get(ctx, pluginNamespacedName, pluginConfigMap); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, errors.Wrap(err, fmt.Sprintf("get plugin %s configmap failed", plugin.Name))
		}
		pluginResources := pluginConfigMap.Data[plugin.Name]

		if err := expend.PluginsInstaller(ctx, plugin.Name, pluginResources); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, errors.Wrap(err, fmt.Sprintf("failed to install plugin %s", plugin.Name))
		}
		logs.Info(fmt.Sprintf("installing plugin %s", plugin.Name))
		plugin.Status.State = conf.PluginInstalling
	}

	if !plugin.Spec.Enabled && plugin.Status.State == conf.PluginInstalled {
		logs.Info(fmt.Sprintf("starting uninstall plugin %s", plugin.Name))
		pluginConfigMap := &corev1.ConfigMap{}
		pluginNamespacedName := types.NamespacedName{Namespace: conf.KubeeyeNameSpace, Name: plugin.Name}
		if err := r.Get(ctx, pluginNamespacedName, pluginConfigMap); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, errors.Wrap(err, fmt.Sprintf("get plugin %s configmap failed", plugin.Name))
		}
		pluginResources := pluginConfigMap.Data[plugin.Name]
		if err := expend.PluginsUninstaller(ctx, plugin.Name, pluginResources); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, errors.Wrap(err, fmt.Sprintf("failed to uninstall plugin %s", plugin.Name))
		}
		logs.Info(fmt.Sprintf("plugin %s uninstall complete", plugin.Name))
		plugin.Status.State = conf.PluginUninstalled
	}

	if err := r.Status().Update(ctx, plugin); err != nil {
		if kubeErr.IsConflict(err) {
			return ctrl.Result{}, err
		} else {
			logs.Error(err, "unexpected error when update status")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	return ctrl.Result{RequeueAfter: 24 * time.Hour}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PluginSubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyepluginsv1alpha1.PluginSubscription{}).
		Complete(r)
}
