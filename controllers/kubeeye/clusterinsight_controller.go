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
	kubeeyepluginsv1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeyeplugins/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/audit"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/plugins"
	"github.com/pkg/errors"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
)

// ClusterInsightReconciler reconciles a ClusterInsight object
type ClusterInsightReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var AuditComplete = 100

// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=clusterinsights/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=nodes;namespaces;events,verbs=get;list
// +kubebuilder:rbac:groups=batch,resources=*,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the ClusterInsight object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterInsightReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logs := log.FromContext(ctx).WithValues("clusterInsight", req.NamespacedName)
	clusterInsight := &kubeeyev1alpha1.ClusterInsight{}

	// get the clusterInsight to determine whether the CRD is created.
	if err := r.Get(ctx, req.NamespacedName, clusterInsight); err != nil {
		if kubeErr.IsNotFound(err) {
			logs.Info("Cluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}
	if !IsTime(clusterInsight.Spec.AuditPeriod) {
		clusterInsight.Spec.AuditPeriod = "0h"
	}

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

	insightName := clusterInsight.ObjectMeta.Name
	ch := make(chan bool)
	if clusterInsight.Status.AuditPercent == 0 || clusterInsight.Status.AuditPercent == 100 {
		{
			kubeeyePlugins := &kubeeyepluginsv1alpha1.PluginSubscriptionList{}
			if err := r.List(ctx, kubeeyePlugins); err != nil {
				logs.Info("Plugins not found")
			}

			// get the list of plugins with result not-ready
			resultNotReadyPlugins := plugins.NotReadyPluginsList(clusterInsight.Status.PluginsResults, kubeeyePlugins)

			// trigger plugins audit tasks
			if len(resultNotReadyPlugins) != 0 {
				plugins.TriggerPluginsAudit(logs, resultNotReadyPlugins)

			}
		}

		t := time.NewTimer(time.Second * 5)

		defer close(ch)

		go func(t *time.Timer) {
			defer t.Stop()
			for {
				select {
				case <-t.C:
					if clusterInsight.Status.AuditPercent == AuditComplete {
						time.Sleep(500 * time.Millisecond)
					}
					percent, ok := audit.AuditPercent.Load(insightName)
					var auditPercent *audit.PercentOutput
					if !ok {
						clusterInsight.Status.AuditPercent = 0
					} else {
						auditPercent = percent.(*audit.PercentOutput)
						clusterInsight.Status.AuditPercent = auditPercent.AuditPercent
					}

					if err := r.Status().Update(ctx, clusterInsight); err != nil {
						if kubeErr.IsConflict(err) {
							return
						} else {
							logs.Error(err, "update CR failed")
							return
						}
					}
					t.Reset(time.Second * 5)
				case <-ch:
					return
				}
			}
		}(t)

		//if clusterInsight.Status.AuditResults == nil {
		if clusterInsight.Status.AuditPercent == 0 {
			logs.Info("Starting kubeeye audit")
			// exec cluster audit
			K8SResources, validationResultsChan := audit.ValidationResults(ctx, clients, "", insightName)

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

			clusterInsight.Status.AuditPercent = AuditComplete
			ch <- true
			audit.AuditPercent.Delete(insightName)

			// update clusterInsight CR
			defer func() {
				if err := r.Status().Update(ctx, clusterInsight); err != nil {
					if kubeErr.IsConflict(err) {
						reterr = err
					} else {
						reterr = errors.Wrap(err, "unexpected error when update status")
					}
				}
			}()

			logs.Info("Cluster audit completed")

		}
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
		Watches(
			&source.Kind{Type: &kubeeyepluginsv1alpha1.PluginSubscription{}},
			handler.EnqueueRequestsFromMapFunc(r.PluginSubscriptionToClusterInsight(context.TODO())),
		).
		For(&kubeeyev1alpha1.ClusterInsight{}).
		Complete(r)
}

func (r *ClusterInsightReconciler) PluginSubscriptionToClusterInsight(ctx context.Context) handler.MapFunc {
	logs := ctrl.LoggerFrom(ctx)
	return func(o client.Object) []reconcile.Request {
		result := []ctrl.Request{}

		c, ok := o.(*kubeeyepluginsv1alpha1.PluginSubscription)
		if !ok {
			logs.Error(errors.Errorf("expected a PluginSubscription but got a %T", o), "failed to get ClusterInsight for PluginSubscription")
		}

		clusterInsight := &kubeeyev1alpha1.ClusterInsightList{}
		if err := r.List(ctx, clusterInsight, client.InNamespace(c.Namespace)); err != nil {
			logs.Error(err, "failed to list ClusterInsight")
		}

		for _, item := range clusterInsight.Items {
			resourceKey := client.ObjectKey{Namespace: item.Namespace, Name: item.Name}
			result = append(result, ctrl.Request{NamespacedName: resourceKey})
		}
		return result
	}
}

func IsTime(period string) bool {
	if '0' <= period[0] && period[0] <= '9' {
		switch string(period[len(period)-1]) {
		case "ns", "us", "ms", "s", "m", "h":
			return true
		default:
			return false
		}
	}
	return false
}
