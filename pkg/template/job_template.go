package template

import (
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InspectJobsTemplate(jobConfig *conf.JobConfig, jobName string, inspectTask *kubeeyev1alpha2.InspectTask, nodeName string, nodeSelector map[string]string, taskType string) *v1.Job {

	var ownerController = true
	ownerRef := metav1.OwnerReference{
		APIVersion:         inspectTask.APIVersion,
		Kind:               inspectTask.Kind,
		Name:               inspectTask.Name,
		UID:                inspectTask.UID,
		Controller:         &ownerController,
		BlockOwnerDeletion: &ownerController,
	}

	var mountPropagation = corev1.MountPropagationHostToContainer
	inspectJob := &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName,
			Namespace:       constant.DefaultNamespace,
			OwnerReferences: []metav1.OwnerReference{ownerRef},
			Labels:          map[string]string{constant.LabelRuleType: taskType},
		},
		Spec: v1.JobSpec{
			BackoffLimit:            jobConfig.BackLimit,
			TTLSecondsAfterFinished: jobConfig.AutoDelTime,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "inspect-job-pod",
					Namespace:   constant.DefaultNamespace,
					Annotations: map[string]string{"container.apparmor.security.beta.kubernetes.io/inspect-task-kubeeye": "unconfined"},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "inspect-task-kubeeye",
						Image:   jobConfig.Image,
						Command: []string{"ke"},
						Args:    []string{"create", "job", taskType, "--task-name", inspectTask.Name, "--result-name", jobName},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "proc",
							ReadOnly:  true,
							MountPath: "/hosts/proc",
						}, {
							Name:      "sys",
							ReadOnly:  true,
							MountPath: "/hosts/sys",
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
						ImagePullPolicy: corev1.PullPolicy(jobConfig.ImagePullPolicy),
						Resources:       jobConfig.Resources,
					}},
					HostNetwork:        true,
					HostPID:            true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
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
