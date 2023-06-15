package template

import (
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InspectJobsTemplate(jobName string, inspectTask *kubeeyev1alpha2.InspectTask, nodeName string, nodeSelector map[string]string, taskType string) *v1.Job {

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
	var autoDelTime int32 = 60
	var mountPropagation = corev1.MountPropagationHostToContainer
	var RunAsNonRoot = true
	inspectJob := &v1.Job{
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
				ObjectMeta: metav1.ObjectMeta{
					Name:        "inspect-job-pod",
					Namespace:   inspectTask.Namespace,
					Annotations: map[string]string{"container.apparmor.security.beta.kubernetes.io/inspect-task-kubeeye": "unconfined"},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "inspect-task-kubeeye",
						Image:   "jw008/kubeeye:dev",
						Command: []string{"ke"},
						Args:    []string{"create", "job", taskType, "--task-name", inspectTask.Name, "--task-namespace", inspectTask.Namespace, "--result-name", jobName},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "proc",
							ReadOnly:  true,
							MountPath: "/host/proc",
						}, {
							Name:      "sys",
							ReadOnly:  true,
							MountPath: "/host/sys",
						}, {
							Name:             "root",
							ReadOnly:         true,
							MountPath:        constant.RootPathPrefix,
							MountPropagation: &mountPropagation,
						}, {
							Name:      "system-socket",
							ReadOnly:  true,
							MountPath: "/var/run/dbus/system_bus_socket",
						}},
						ImagePullPolicy: "Always",
						Resources: corev1.ResourceRequirements{
							Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1000m"), corev1.ResourceMemory: resource.MustParse("512Mi")},
							Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("500m"), corev1.ResourceMemory: resource.MustParse("256Mi")},
						},
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:           &RunAsNonRoot,
							ReadOnlyRootFilesystem: &RunAsNonRoot,
						},
					}},
					ServiceAccountName: "kubeeye-controller-manager",
					NodeName:           nodeName,
					NodeSelector:       nodeSelector,
					RestartPolicy:      corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{{
						Name: "root",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/",
							},
						},
					}, {
						Name: "proc",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/proc",
							},
						},
					}, {
						Name: "sys",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/sys",
							},
						},
					}, {
						Name: "system-socket",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/run/dbus/system_bus_socket",
							},
						},
					}},
				},
			},
		},
	}

	return inspectJob
}
