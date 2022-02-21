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
	"github.com/kubesphere/kubeeye/pkg/kube"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	log := log.FromContext(ctx)

	clusterInsight := &kubeeyev1alpha1.ClusterInsight{}

	if err := r.Get(ctx, req.NamespacedName, clusterInsight); err != nil {
		if kubeErr.IsNotFound(err) {
			log.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	// get kubernetes cluster config
	kubeConfig, err := rest.InClusterConfig()
	//kubeConfig, err := config.GetConfig()
	if err != nil {
		log.Error(err, "failed to get cluster config")
		return ctrl.Result{}, err
	}

	// get kubernetes cluster clients
	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		log.Error(err, "Failed to load cluster clients")
		return ctrl.Result{}, err
	}

	log.Info("Starting cluster audit")
	K8SResources, validationResultsChan := audit.ValidationResults(ctx, clients, "")

	// set cluster info
	clusterInsight.Status.ClusterInfo = setClusterInfo(K8SResources)

	// clear clusterInsight.Status.AuditResults
	clusterInsight.Status.AuditResults = kubeeyev1alpha1.AuditResult{}

	// fill clear clusterInsight.Status.AuditResults
	for receiver := range validationResultsChan {
		for _, results := range receiver.Results {
			if results.ResultInfos != nil {
				clusterInsight.Status.AuditResults.Results = append(clusterInsight.Status.AuditResults.Results, results)
			}
		}
	}

	// update clusterInsight CR
	if err := r.Status().Update(ctx, clusterInsight); err != nil {
		log.Error(err, "Update CR Status failed")
		return ctrl.Result{}, err
	}

	log.Info("Cluster audit completed")

	// If auditPeriod is not set, set the default value to 24h
	if clusterInsight.Spec.AuditPeriod == "" {
		clusterInsight.Spec.AuditPeriod = "24h"
	}

	reconcilePeriod, err := time.ParseDuration(clusterInsight.Spec.AuditPeriod)
	if err != nil {
		log.Error(err, "AuditPeriod setting is invalid")
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

func setClusterInfo(k8SResource kube.K8SResource) (ClusterInfo kubeeyev1alpha1.ClusterInfo) {
	ClusterInfo.ClusterVersion = k8SResource.ServerVersion
	ClusterInfo.NodesCount = k8SResource.NodesCount
	ClusterInfo.NamespacesCount = k8SResource.NameSpacesCount
	ClusterInfo.NamespacesList = k8SResource.NameSpacesList
	ClusterInfo.WorkloadsCount = k8SResource.WorkloadsCount
	return ClusterInfo
}
