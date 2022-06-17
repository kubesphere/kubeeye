package kubeeye

import (
	"time"

	kubeeyeclientset "github.com/kubesphere/kubeeye/client/clientset/versioned"
	kubeeyeinformers "github.com/kubesphere/kubeeye/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which kubesphere needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	KubeeyeSharedInformerFactory() kubeeyeinformers.SharedInformerFactory
	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	k8SInformerFactory     k8sinformers.SharedInformerFactory
	kubeeyeInformerFactory kubeeyeinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, kubeeyeclient kubeeyeclientset.Interface) InformerFactory {
	factory := &informerFactories{}
	if client != nil {
		factory.k8SInformerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}
	if kubeeyeclient != nil {
		factory.kubeeyeInformerFactory = kubeeyeinformers.NewSharedInformerFactory(kubeeyeclient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.k8SInformerFactory
}

func (f *informerFactories) KubeeyeSharedInformerFactory() kubeeyeinformers.SharedInformerFactory {
	return f.kubeeyeInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.k8SInformerFactory != nil {
		f.k8SInformerFactory.Start(stopCh)
	}
	if f.kubeeyeInformerFactory != nil {
		f.kubeeyeInformerFactory.Start(stopCh)
	}

}
