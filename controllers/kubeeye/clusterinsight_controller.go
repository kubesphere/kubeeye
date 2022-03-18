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
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
)

// ClusterInsightReconciler reconciles a ClusterInsight object
type ClusterInsightReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list
//+kubebuilder:rbac:groups=batch,resources=*,verbs=get;list
//+kubebuilder:rbac:groups=apps,resources=*,verbs=get;list
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=get;list

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

	err := r.Get(ctx, req.NamespacedName, clusterInsight)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			logs.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	// manage plugins
	// manage NPD plugins,if NPD enabled is set to true,NPD will be installed.
	if clusterInsight.Spec.Plugins.NPDState.Enabled {
		if err := expend.InstallNPD(ctx, ""); err != nil {
			logs.Error(err, "Install npd failed with error: %v")
			return ctrl.Result{}, err
		}
	} else {
		if err := expend.UninstallNPD(ctx,""); err != nil {
			logs.Error(err, "Uninstall npd failed with error: %v")
			return ctrl.Result{}, err
		}
	}

	// manage KubeBench plugins,if KubeBench enabled is set to true,KubeBench will be installed.
	if clusterInsight.Spec.Plugins.KubeBenchState.Enabled {
		if err := expend.InstallKubeBench(ctx, ""); err != nil {
			logs.Error(err, "Install KubeBench failed with error: %v")
			return ctrl.Result{}, err
		}
	} else {
		if err := expend.UninstallKubeBench(ctx, ""); err != nil {
			logs.Error(err, "Uninstall KubeBench failed with error: %v")
			return ctrl.Result{}, err
		}
	}

	var kubeConfig *rest.Config
	// get kubernetes cluster config
	kubeConfig, err = rest.InClusterConfig()
	if err != nil {
		kubeConfig, err = config.GetConfig()
		if err != nil {
			logs.Error(err, "failed to get cluster config")
			return ctrl.Result{}, err
		}
	}

	// get kubernetes cluster clients
	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		logs.Error(err, "Failed to load cluster clients")
		return ctrl.Result{}, err
	}

	logs.Info("Starting cluster audit")
	K8SResources, validationResultsChan := audit.ValidationResults(ctx, clients, "")

	// get cluster info
	clusterInfo := setClusterInfo(K8SResources)

	// fill clusterInsight.Status.ClusterInfo
	clusterInsight.Status.ClusterInfo = clusterInfo

	// clear clusterInsight.Status.AuditResults
	clusterInsight.Status.AuditResults = []kubeeyev1alpha1.AuditResults{}

	//format result
	fmResults := formatResults(validationResultsChan)

	// fill clusterInsight.Status.AuditResults
	clusterInsight.Status.AuditResults = fmResults

	// get score
	scoreInfo := CalculateScore(fmResults, K8SResources)

	// fill
	clusterInsight.Status.ScoreInfo = scoreInfo

	// update clusterInsight CR
	if err := r.Status().Update(ctx, clusterInsight); err != nil {
		logs.Error(err, "Update CR Status failed")
		return ctrl.Result{}, err
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
		For(&kubeeyev1alpha1.ClusterInsight{}).
		Complete(r)
}