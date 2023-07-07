package template

import (
	"context"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/kube"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

func GetJobConfig(ctx context.Context, client *kube.KubernetesClient) *conf.JobConfig {

	kubeeyeCm, err := client.ClientSet.CoreV1().ConfigMaps("kubeeye-system").Get(ctx, "kubeeye-config", metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get kubeeye config, kubeeye config file do not exist. err:%s", err)
		return nil
	}
	config := kubeeyeCm.Data["config"]
	var KubeEyeConfig conf.KubeeyeConfig
	err = yaml.Unmarshal([]byte(config), &KubeEyeConfig)
	if err != nil {
		klog.Errorf("failed to unmarshal kubeeye config. err:%s ", err)
		return nil
	}
	return KubeEyeConfig.Job
}

func InspectJobsTemplate(ctx context.Context, client *kube.KubernetesClient, jobName string, inspectTask *kubeeyev1alpha2.InspectTask, nodeName string, nodeSelector map[string]string, taskType string) *v1.Job {

	jobConfig := GetJobConfig(ctx, client)
	if jobConfig == nil {
		klog.Error("Unable to get jobConfig")
		return nil
	}

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
			Namespace:       "kubeeye-system",
			OwnerReferences: []metav1.OwnerReference{ownerRef},
			Labels:          map[string]string{constant.LabelResultName: taskType},
		},
		Spec: v1.JobSpec{
			BackoffLimit: jobConfig.BackLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "inspect-job-pod",
					Namespace:   "kubeeye-system",
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
