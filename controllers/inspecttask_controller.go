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
	"github.com/kubesphere/kubeeye/pkg/utils"
	"time"

	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InspectTaskReconciler reconciles a InspectTask object
type InspectTaskReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	K8sClients *kube.KubernetesClient
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources="*",verbs=get;list;watch
//+kubebuilder:rbac:groups="apps",resources="*",verbs=get;list
//+kubebuilder:rbac:groups="batch",resources="*",verbs=get;list;create
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources="*",verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *InspectTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := log.FromContext(ctx).WithName(req.NamespacedName.String())

	inspectTask := &kubeeyev1alpha2.InspectTask{}
	err := r.Get(ctx, req.NamespacedName, inspectTask)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			klog.Infof("inspect task is not found;name:%s,namespect:%s\n", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		klog.Error("failed to get inspect task. ", err)
		return ctrl.Result{}, err
	}

	if inspectTask.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, inspectTask.Finalizers); !b {
			inspectTask.Finalizers = append(inspectTask.Finalizers, Finalizers)
			err = r.Client.Update(ctx, inspectTask)
			if err != nil {
				klog.Error("Failed to inspect plan add finalizers", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, inspectTask.Finalizers)
		inspectTask.Finalizers = newFinalizers.([]string)
		klog.Infof("inspect ruleFiles is being deleted")
		err = r.Client.Update(ctx, inspectTask)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers. ", err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	// if Audit task have not start, trigger kubeeye and plugin
	if inspectTask.Status.StartTimestamp.IsZero() {
		inspectTask.Status.StartTimestamp = &metav1.Time{Time: time.Now()}
		inspectTask.Status.ClusterInfo, err = r.getClusterInfo(ctx)
		if err != nil {
			klog.Error("failed to get cluster info. ", err)
			return ctrl.Result{}, err
		}
		JobPhase, err := r.createJobsInspect(ctx, inspectTask)
		if err != nil {
			return ctrl.Result{}, err
		}
		inspectTask.Status.JobPhase = JobPhase
		klog.Infof("%s start task ", req.Name)
		err = r.Status().Update(ctx, inspectTask)
		if err != nil {
			klog.Error("failed to update inspect task. ", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	} else {
		_, complete := r.IsComplete(inspectTask.Status.JobPhase)
		if complete {
			klog.Infof("all job finished for taskName:%s", inspectTask.Name)
			return ctrl.Result{}, nil
		}
		updateStatus := false
		for i, job := range inspectTask.Status.JobPhase {
			if job.Phase != kubeeyev1alpha2.PhaseRunning {
				continue
			}

			jobInfo, err := r.K8sClients.ClientSet.BatchV1().Jobs(inspectTask.Namespace).Get(ctx, job.JobName, metav1.GetOptions{})
			if err != nil {
				klog.Error(err)
				inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
				updateStatus = true
				continue
			}
			if jobInfo.Status.CompletionTime != nil && !jobInfo.Status.CompletionTime.IsZero() && jobInfo.Status.Active == 0 {
				updateStatus = true
				configs, err := r.K8sClients.ClientSet.CoreV1().ConfigMaps(inspectTask.Namespace).Get(ctx, job.JobName, metav1.GetOptions{})
				if err != nil {
					klog.Error(err)
					inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
					continue
				}
				inspectInterface, status := inspect.RuleOperatorMap[jobInfo.Labels[constant.LabelResultName]]
				if status {
					klog.Infof("starting get %s result data", job.JobName)
					err = inspectInterface.GetResult(ctx, r.Client, jobInfo, configs, inspectTask)
					if err != nil {
						klog.Error(err)
						inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
						continue
					}
				}
				inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseSucceeded
				_ = r.K8sClients.ClientSet.CoreV1().ConfigMaps(inspectTask.Namespace).Delete(ctx, job.JobName, metav1.DeleteOptions{})
			}
		}

		timeout, err := time.ParseDuration(inspectTask.Spec.Timeout)
		if err != nil {
			timeout = constant.DefaultTimeout
		}
		if inspectTask.Status.StartTimestamp.Add(timeout).Before(time.Now()) {
			for i, job := range inspectTask.Status.JobPhase {
				if job.Phase == kubeeyev1alpha2.PhaseRunning {
					inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
					var DeletePro = metav1.DeletePropagationBackground
					err := r.K8sClients.ClientSet.BatchV1().Jobs(inspectTask.Namespace).Delete(ctx, job.JobName, metav1.DeleteOptions{PropagationPolicy: &DeletePro})
					if err != nil {
						klog.Errorf("failed to delete jobs for jobName:%s,%s", job.JobName, err)
						continue
					}
				}
			}
		}
		if updateStatus {
			err = r.Status().Update(ctx, inspectTask)
			if err != nil && !kubeErr.IsNotFound(err) {
				klog.Error("failed to update inspect task. ", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
}

func (r *InspectTaskReconciler) getClusterInfo(ctx context.Context) (kubeeyev1alpha2.ClusterInfo, error) {
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
func (r *InspectTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectTask{}).
		Complete(r)
}

func (r *InspectTaskReconciler) createJobsInspect(ctx context.Context, inspectTask *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	for key := range inspectTask.Spec.Rules {
		inspectInterface, status := inspect.RuleOperatorMap[key]
		if status {
			task, err := inspectInterface.CreateJobTask(ctx, r.K8sClients, inspectTask)
			if err != nil {
				klog.Error("create job error")
				continue
			}
			jobNames = append(jobNames, task...)
		} else {
			klog.Errorf("%s not found", key)
		}

	}

	return jobNames, nil
}

func (r *InspectTaskReconciler) IsComplete(JobPhase []kubeeyev1alpha2.JobPhase) ([]kubeeyev1alpha2.JobPhase, bool) {
	var Jobs []kubeeyev1alpha2.JobPhase
	for _, job := range JobPhase {
		if job.Phase == kubeeyev1alpha2.PhaseRunning {
			Jobs = append(Jobs, job)
		}
	}
	return Jobs, len(Jobs) == 0
}
