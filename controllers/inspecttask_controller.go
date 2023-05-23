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
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
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
	Audit      *inspect.Audit
	K8sClients *kube.KubernetesClient
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources="*",verbs=get;list
//+kubebuilder:rbac:groups="apps",resources="*",verbs=get;list
//+kubebuilder:rbac:groups="batch",resources="*",verbs=get;list
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
		//JobPhase, err := r.createJobsInspect(ctx, inspectTask)
		//if err != nil {
		//	return ctrl.Result{}, err
		//}
		//inspectTask.Status.JobPhase = JobPhase
		klog.Infof("%s start task ", req.Name)
	} else {
		if r.IsComplete(inspectTask) {
			klog.Info("all job finished")
			return ctrl.Result{}, nil
		}

		for i, job := range inspectTask.Status.JobPhase {
			if job.Phase != kubeeyev1alpha2.PhaseRunning {
				continue
			}

			jobs, err := r.K8sClients.ClientSet.BatchV1().Jobs(inspectTask.Namespace).Get(ctx, job.JobName, metav1.GetOptions{})
			if err != nil {
				klog.Error(err)
				continue
			}
			if jobs.Status.CompletionTime != nil && !jobs.Status.CompletionTime.IsZero() && jobs.Status.Active == 0 {
				configs, err := r.K8sClients.ClientSet.CoreV1().ConfigMaps(inspectTask.Namespace).Get(ctx, job.JobName, metav1.GetOptions{})
				if err != nil {
					klog.Error(err)
					continue
				}
				switch jobs.Labels[constant.LabelResultName] {
				case constant.Opa:
					break
				case constant.Prometheus:
					klog.Info("进来了")
					break
				case constant.FileChange:
					err = inspect.GetFileChangeResult(ctx, r.Client, jobs, configs, inspectTask)
					if err != nil {
						return ctrl.Result{}, err
					}
					break
				default:
					klog.Error("Unable to get results")
					break
				}
				inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseSucceeded
				err = r.K8sClients.ClientSet.CoreV1().ConfigMaps(inspectTask.Namespace).Delete(ctx, job.JobName, metav1.DeleteOptions{})
				if err != nil {
					klog.Errorf("failed to delete result for configMap:%s", job.JobName)
					continue
				}
			}
			if jobs.Status.Conditions != nil && jobs.Status.Conditions[0].Type == v1.JobFailed {
				inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
			}
		}

		timeout, err := time.ParseDuration(inspectTask.Spec.Timeout)
		if err != nil {
			timeout = constant.DefaultTimeout
		}
		if inspectTask.Status.StartTimestamp.Add(timeout).Before(time.Now()) {
			for i, job := range inspectTask.Status.JobPhase {
				inspectTask.Status.JobPhase[i].Phase = kubeeyev1alpha2.PhaseFailed
				var DeletePro = metav1.DeletePropagationBackground
				err := r.K8sClients.ClientSet.BatchV1().Jobs(inspectTask.Namespace).Delete(ctx, job.JobName, metav1.DeleteOptions{PropagationPolicy: &DeletePro})
				if err != nil {
					klog.Errorf("failed to delete jobs for jobName:%s", job.JobName)
					continue
				}
			}
		}

	}

	err = r.Status().Update(ctx, inspectTask)
	if err != nil && !kubeErr.IsNotFound(err) {
		klog.Error("failed to update inspect task. ", err)
		return ctrl.Result{RequeueAfter: 60 * time.Second}, err
	}
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
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
	var name = fmt.Sprintf("inspect-job-%s", strconv.Itoa(int(time.Now().Unix())))
	var jobNames []kubeeyev1alpha2.JobPhase
	for key := range inspectTask.Spec.Rules {
		if key == constant.Prometheus || key == constant.Opa {
			jobName, err := r.inspectJobsTemplate(ctx, name, inspectTask, "", constant.Prometheus)
			if err != nil {
				klog.Errorf("Failed to create Jobs for node name:%s,err:%s", err, err)
				return nil, err
			}
			jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: jobName, Phase: kubeeyev1alpha2.PhaseRunning})
		}
		if key == constant.FileChange {
			nodes, err := r.K8sClients.ClientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				klog.Error("Failed to get nodes info", err)
				return nil, err
			}
			for _, node := range nodes.Items {
				jobName, err := r.inspectJobsTemplate(ctx, fmt.Sprintf("%s-%s", name, node.Name), inspectTask, node.Name, constant.FileChange)
				if err != nil {
					klog.Errorf("Failed to create Jobs for node name:%s,err:%s", err, err)
					return nil, err
				}
				jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: jobName, Phase: kubeeyev1alpha2.PhaseRunning})
			}

		}
	}

	return jobNames, nil
}

func (r *InspectTaskReconciler) inspectJobsTemplate(ctx context.Context, jobName string, inspectTask *kubeeyev1alpha2.InspectTask, nodeName string, taskType string) (string, error) {
	var ownerController = true
	ownerRef := metav1.OwnerReference{
		APIVersion:         inspectTask.APIVersion,
		Kind:               inspectTask.Kind,
		Name:               inspectTask.Name,
		UID:                inspectTask.UID,
		Controller:         &ownerController,
		BlockOwnerDeletion: &ownerController,
	}
	var resetBack int32 = 5
	var autoDelTime int32 = 30
	inspectJob := v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       inspectTask.Namespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
			Labels:          map[string]string{constant.LabelResultName: taskType},
		},
		Spec: v1.JobSpec{
			BackoffLimit:            &resetBack,
			TTLSecondsAfterFinished: &autoDelTime,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Name: "inspect-job-pod", Namespace: inspectTask.Namespace},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "inspect-task-kubeeye",
						Image:   "jw008/kubeeye:amd64",
						Command: []string{"inspect"},
						Args:    []string{taskType, "--task-name", inspectTask.Name, "--task-namespace", inspectTask.Namespace, "--result-name", jobName},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "root-path",
							ReadOnly:  true,
							MountPath: "/host",
						}},
						ImagePullPolicy: "Always",
					}},
					ServiceAccountName: "kubeeye-controller-manager",
					NodeName:           nodeName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{{
						Name: "root-path",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/",
							},
						},
					}},
				},
			},
		},
	}
	err := r.Create(ctx, &inspectJob)

	if err != nil {
		klog.Error(err)
		return "", err
	}
	return inspectJob.Name, nil
}

func (r *InspectTaskReconciler) IsComplete(task *kubeeyev1alpha2.InspectTask) bool {
	for _, job := range task.Status.JobPhase {
		if job.Phase == kubeeyev1alpha2.PhaseRunning {
			return false
		}
	}

	return true
}
