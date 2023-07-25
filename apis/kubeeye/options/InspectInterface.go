package options

import (
	"context"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/kube"
	v12 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InspectInterface interface {
	CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error)
	RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...v1.OwnerReference) ([]byte, error)
	GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error)
}

type InspectType struct {
	Clients        *kube.KubernetesClient
	JobRule        *kubeeyev1alpha2.JobRule
	Task           *kubeeyev1alpha2.InspectTask
	CurrentJobName string
	Jobs           *v12.Job
	Result         *corev1.ConfigMap
	OwnerRef       []v1.OwnerReference
}
