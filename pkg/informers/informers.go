package informers

import (
	kubeeyeClient "github.com/kubesphere/kubeeye/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/clients/informers/externalversions"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type InformerFactory interface {
	KubeEyeInformerFactory() externalversions.SharedInformerFactory
	KubernetesInformerFactory() informers.SharedInformerFactory
	Start(stopCh <-chan struct{})
	ForResources(keEyeGver map[schema.GroupVersion][]string, k8sEyeGver map[schema.GroupVersion][]string)
}

type informerFactory struct {
	kubeEyeInformerFactory    externalversions.SharedInformerFactory
	kubernetesInformerFactory informers.SharedInformerFactory
}

func NewInformerFactory(k8sClient kubernetes.Interface, kubeEyeClient kubeeyeClient.Interface) InformerFactory {
	info := &informerFactory{}
	if k8sClient != nil {
		info.kubernetesInformerFactory = informers.NewSharedInformerFactory(k8sClient, constant.DefaultTimeout)
	}
	if kubeEyeClient != nil {
		info.kubeEyeInformerFactory = externalversions.NewSharedInformerFactory(kubeEyeClient, constant.DefaultTimeout)
	}
	return info
}

func (i *informerFactory) KubeEyeInformerFactory() externalversions.SharedInformerFactory {
	return i.kubeEyeInformerFactory
}

func (i *informerFactory) KubernetesInformerFactory() informers.SharedInformerFactory {
	return i.kubernetesInformerFactory
}

func (i *informerFactory) Start(stopCh <-chan struct{}) {
	if i.kubernetesInformerFactory != nil {
		i.kubernetesInformerFactory.Start(stopCh)

	}
	if i.kubeEyeInformerFactory != nil {
		i.kubeEyeInformerFactory.Start(stopCh)
	}
}

func (i *informerFactory) ForResources(keEyeGver map[schema.GroupVersion][]string, k8sEyeGver map[schema.GroupVersion][]string) {

	if i.kubeEyeInformerFactory != nil && keEyeGver != nil {
		for groupVersion, resources := range keEyeGver {
			for _, resource := range resources {
				_, err := i.kubeEyeInformerFactory.ForResource(groupVersion.WithResource(resource))
				if err != nil {
					klog.Error(err)
				}
			}

		}

	}
	if i.kubernetesInformerFactory != nil && k8sEyeGver != nil {
		for groupVersion, resources := range k8sEyeGver {
			for _, resource := range resources {
				_, err := i.kubernetesInformerFactory.ForResource(groupVersion.WithResource(resource))
				if err != nil {
					klog.Error(err)
				}
			}
		}
	}

}
