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

package kubeeye

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/kubesphere/kubeeye/pkg/audit"
	"github.com/kubesphere/kubeeye/pkg/expend"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/plugins"
	"github.com/kubesphere/kubeeye/plugins/plugin-manage/pkg"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
)

// ClusterInsightReconciler reconciles a ClusterInsight object
type ClusterInsightReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list
// +kubebuilder:rbac:groups=batch,resources=*,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=*,verbs=get;list
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the ClusterInsight object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterInsightReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logs := log.FromContext(ctx)
	clusterInsight := &kubeeyev1alpha1.ClusterInsight{}

	// get the clusterInsight to determine whether the CRD is created.
	if err := r.Get(ctx, req.NamespacedName, clusterInsight); err != nil {
		if kubeErr.IsNotFound(err) {
			logs.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	pluginsResults  := clusterInsight.Status.PluginsResults
	var pluginsList []string
	var notreadyplugins []string

	var kubeConfig *rest.Config
	// get kubernetes cluster config
	kubeConfig, err := kube.GetKubeConfigInCluster()
	if err != nil {
		return ctrl.Result{}, err
	}

	// get kubernetes cluster clients
	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		logs.Error(err, "Failed to load cluster clients")
		return ctrl.Result{}, err
	}

	logs.Info("Starting cluster audit")

	// get plugins list
	pluginsList, err = expend.ListCRDResources(ctx, clients.DynamicClient, clusterInsight.GetNamespace())
	if err != nil {
		logs.Info( "Plugins not found")
	}

	// get not-ready plugins list
	notreadyplugins = plugins.NotReadyPluginsList(pluginsResults, pluginsList)
	// exec plugins by goroutine
	var pluginName string
	if len(notreadyplugins) != 0 {
		logs.Info( "Starting plugin audit")

		pluginName = plugins.RandomPluginName(notreadyplugins)
		// Get a plugin name for plugin audit
		go plugins.PluginsAudit(logs, pluginName)
	}

	{
		logs.Info("Starting kubeeye audit")
		// exec cluster audit
		K8SResources, validationResultsChan := audit.ValidationResults(ctx, clients, "")

		// get cluster info
		clusterInfo := setClusterInfo(K8SResources)

		// fill clusterInsight.Status.ClusterInfo
		clusterInsight.Status.ClusterInfo = clusterInfo

		// format result
		fmResults := formatResults(validationResultsChan)

		// fill clusterInsight.Status.AuditResults
		clusterInsight.Status.AuditResults = fmResults

		// get score
		scoreInfo := CalculateScore(fmResults, K8SResources)

		// fill
		clusterInsight.Status.ScoreInfo = scoreInfo
	}

	// get plugins results
	if len(notreadyplugins) != 0 {
		logs.Info( "get plugins result")
		select {
		case res := <-kube.PluginResultChan:
			pluginsResults = plugins.MergePluginsResults(pluginsResults, res)
		case <- time.After(30 * time.Second):
			logs.Info("plugins results not found")
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	// set PluginsResults in clusterInsight
	if len(pluginsResults) != 0 {
		clusterInsight.Status.PluginsResults = pluginsResults
	}

	// update clusterInsight CR
	if err := r.Status().Update(ctx, clusterInsight); err != nil {
		if kubeErr.IsConflict(err) {
			return ctrl.Result{}, err
		} else {
			logs.Error(err, "unexpected error when update status")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	if len(notreadyplugins) !=0 {
		logs.Info("plugins with unready state， retry plugins audit")
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	logs.Info("Cluster audit completed")

	// If auditPeriod is not set, set the default value to 24h
	if clusterInsight.Spec.AuditPeriod == "" {
		clusterInsight.Spec.AuditPeriod = "24h"
	}

	reconcilePeriod, err := time.ParseDuration(clusterInsight.Spec.AuditPeriod)
	if err != nil {
		logs.Error(err, "AuditPeriod setting is invalid")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: reconcilePeriod}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterInsightReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: pkg.MaxConcurrentReconciles,
		}).
		For(&kubeeyev1alpha1.ClusterInsight{}).
		Complete(r)
}
