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
	"time"

	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/audit"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/kube"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
)

// AuditTaskReconciler reconciles a AuditTask object
type AuditTaskReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Audit      *audit.Audit
	K8sClients *kube.KubernetesClient
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=audittasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=audittasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=audittasks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuditTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *AuditTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName(req.NamespacedName.String())

	auditTask := &kubeeyev1alpha2.AuditTask{}
	err := r.Get(ctx, req.NamespacedName, auditTask)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			delete(r.Audit.TaskOnceMap, req.NamespacedName)
			delete(r.Audit.TaskResults, auditTask.Name)
			logger.Error(err, "audit task is not found")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "failed to get audit task")
		return ctrl.Result{}, err
	}

	if !auditTask.DeletionTimestamp.IsZero() {
		delete(r.Audit.TaskOnceMap, req.NamespacedName)
		delete(r.Audit.TaskResults, auditTask.Name)
		logger.Info("audit task is being deleted")
		return ctrl.Result{}, nil
	}

	if auditTask.Status.StartTimestamp.IsZero() { // if Audit task have not start, trigger kubeeye and plugin

		// start Audit
		r.Audit.AddTaskToQueue(req.NamespacedName)

		auditTask.Status.StartTimestamp = &metav1.Time{Time: time.Now()}
		auditTask.Status.Phase = kubeeyev1alpha2.PhaseRunning
		// get cluster info : ClusterVersion, NodesCount, NamespaceCount
		auditTask.Status.ClusterInfo, err = r.getClusterInfo(ctx)
		if err != nil {
			logger.Error(err, "failed to get cluster info")
			return ctrl.Result{}, err
		}
		logger.Info("start task ", "name", req.Name)
	} else {
		if auditTask.Status.Phase == kubeeyev1alpha2.PhaseSucceeded || auditTask.Status.Phase == kubeeyev1alpha2.PhaseFailed {
			// remove from processing queue
			delete(r.Audit.TaskResults, auditTask.Name)
			return ctrl.Result{}, nil
		} else {
			resultMap, ok := r.Audit.TaskResults[auditTask.Name]
			if !ok {
				r.Audit.AddTaskToQueue(req.NamespacedName)
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
			var results []kubeeyev1alpha2.AuditResult
			completed := 0
			for _, auditor := range auditTask.Spec.Auditors {
				if result, ok := resultMap[string(auditor)]; ok {
					results = append(results, *result)
					if result.Phase == kubeeyev1alpha2.PhaseSucceeded {
						completed++
					}
				}
			}
			auditTask.Status.AuditResults = results
			auditTask.Status.CompleteItemCount = completed
		}

		timeout, err := time.ParseDuration(auditTask.Spec.Timeout)
		if err != nil {
			timeout = constant.DefaultTimeout
		}
		if auditTask.Status.CompleteItemCount == len(auditTask.Spec.Auditors) {
			auditTask.Status.Phase = kubeeyev1alpha2.PhaseSucceeded
			auditTask.Status.EndTimestamp = &metav1.Time{Time: time.Now()}
		} else if auditTask.Status.StartTimestamp.Add(timeout).Before(time.Now()) {
			auditTask.Status.Phase = kubeeyev1alpha2.PhaseFailed
		}
	}

	err = r.Status().Update(ctx, auditTask)
	if err != nil && !kubeErr.IsNotFound(err) {
		logger.Error(err, "failed to update audit task")
		return ctrl.Result{RequeueAfter: 60 * time.Second}, err
	}
	return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
}

func (r *AuditTaskReconciler) getClusterInfo(ctx context.Context) (kubeeyev1alpha2.ClusterInfo, error) {
	var clusterInfo kubeeyev1alpha2.ClusterInfo
	versionInfo, err := r.K8sClients.ClientSet.Discovery().ServerVersion()
	if err != nil {
		klog.Error(err, "Failed to get Kubernetes serverVersion.\n")
	}
	var serverVersion string
	if versionInfo != nil {
		serverVersion = versionInfo.Major + "." + versionInfo.Minor
	}
	_, nodesCount, err := kube.GetObjectCounts(ctx, r.K8sClients, conf.Nodes, conf.NoGroup)
	if err != nil {
		klog.Error(err, "Failed to get node number.")
	}
	_, namespacesCount, err := kube.GetObjectCounts(ctx, r.K8sClients, conf.Namespaces, conf.NoGroup)
	if err != nil {
		klog.Error(err, "Failed to get ns number.")
	}
	clusterInfo = kubeeyev1alpha2.ClusterInfo{ClusterVersion: serverVersion, NodesCount: nodesCount, NamespacesCount: namespacesCount}
	return clusterInfo, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *AuditTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.AuditTask{}).
		Complete(r)
}
