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
// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	time "time"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	versioned "github.com/kubesphere/kubeeye/clients/clientset/versioned"
	internalinterfaces "github.com/kubesphere/kubeeye/clients/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/kubesphere/kubeeye/clients/listers/kubeeye/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AuditPlanInformer provides access to a shared informer and lister for
// AuditPlans.
type AuditPlanInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.AuditPlanLister
}

type auditPlanInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAuditPlanInformer constructs a new informer for AuditPlan type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAuditPlanInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAuditPlanInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAuditPlanInformer constructs a new informer for AuditPlan type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAuditPlanInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubeeyeV1alpha2().AuditPlans(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubeeyeV1alpha2().AuditPlans(namespace).Watch(context.TODO(), options)
			},
		},
		&kubeeyev1alpha2.AuditPlan{},
		resyncPeriod,
		indexers,
	)
}

func (f *auditPlanInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAuditPlanInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *auditPlanInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubeeyev1alpha2.AuditPlan{}, f.defaultInformer)
}

func (f *auditPlanInformer) Lister() v1alpha2.AuditPlanLister {
	return v1alpha2.NewAuditPlanLister(f.Informer().GetIndexer())
}