package template

import (
	"github.com/kubesphere/kubeeye/pkg/constant"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GeneratorJobTemplate(Job JobTemplateOptions) *v1.Job {
	var ownerController = true
	mountPropagation := corev1.MountPropagationHostToContainer
	return &v1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Job.JobName,
			Namespace: constant.DefaultNamespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         Job.Task.APIVersion,
				Kind:               Job.Task.Kind,
				Name:               Job.Task.Name,
				UID:                Job.Task.UID,
				Controller:         &ownerController,
				BlockOwnerDeletion: &ownerController,
			}},
			Labels: map[string]string{constant.LabelRuleType: Job.RuleType},
		},
		Spec: v1.JobSpec{
			BackoffLimit:            Job.JobConfig.BackLimit,
			TTLSecondsAfterFinished: Job.JobConfig.AutoDelTime,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "inspect-job-pod",
					Namespace:   constant.DefaultNamespace,
					Annotations: map[string]string{"container.apparmor.security.beta.kubernetes.io/inspect-task-kubeeye": "unconfined"},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "inspect-task-kubeeye",
						Image:   Job.JobConfig.Image,
						Command: []string{"ke"},
						Args:    []string{"create", "job", "--job-type", Job.RuleType, "--task-name", Job.Task.Name, "--result-name", Job.JobName},
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
						ImagePullPolicy: corev1.PullPolicy(Job.JobConfig.ImagePullPolicy),
						Resources:       Job.JobConfig.Resources,
					}},
					HostNetwork:        true,
					HostPID:            true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					ServiceAccountName: "kubeeye-inspect-job",
					NodeName:           Job.NodeName,
					NodeSelector:       Job.NodeSelector,
					RestartPolicy:      corev1.RestartPolicyNever,
					Tolerations: []corev1.Toleration{
						{
							Key:      "",
							Operator: corev1.TolerationOpExists,
							Value:    "",
							Effect:   "",
						},
					},
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

}
