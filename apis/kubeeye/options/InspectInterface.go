package options

import (
	"context"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
	v12 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InspectInterface interface {
	CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error)
	RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...v1.OwnerReference) ([]byte, error)
	GetResult(ctx context.Context, c client.Client, jobs *v12.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error
}
