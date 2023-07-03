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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// InspectTaskLister helps list InspectTasks.
// All objects returned here must be treated as read-only.
type InspectTaskLister interface {
	// List lists all InspectTasks in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.InspectTask, err error)
	// InspectTasks returns an object that can list and get InspectTasks.
	InspectTasks(namespace string) InspectTaskNamespaceLister
	InspectTaskListerExpansion
}

// inspectTaskLister implements the InspectTaskLister interface.
type inspectTaskLister struct {
	indexer cache.Indexer
}

// NewInspectTaskLister returns a new InspectTaskLister.
func NewInspectTaskLister(indexer cache.Indexer) InspectTaskLister {
	return &inspectTaskLister{indexer: indexer}
}

// List lists all InspectTasks in the indexer.
func (s *inspectTaskLister) List(selector labels.Selector) (ret []*v1alpha2.InspectTask, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.InspectTask))
	})
	return ret, err
}

// InspectTasks returns an object that can list and get InspectTasks.
func (s *inspectTaskLister) InspectTasks(namespace string) InspectTaskNamespaceLister {
	return inspectTaskNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// InspectTaskNamespaceLister helps list and get InspectTasks.
// All objects returned here must be treated as read-only.
type InspectTaskNamespaceLister interface {
	// List lists all InspectTasks in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.InspectTask, err error)
	// Get retrieves the InspectTask from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.InspectTask, error)
	InspectTaskNamespaceListerExpansion
}

// inspectTaskNamespaceLister implements the InspectTaskNamespaceLister
// interface.
type inspectTaskNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all InspectTasks in the indexer for a given namespace.
func (s inspectTaskNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.InspectTask, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.InspectTask))
	})
	return ret, err
}

// Get retrieves the InspectTask from the indexer for a given namespace and name.
func (s inspectTaskNamespaceLister) Get(name string) (*v1alpha2.InspectTask, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("inspecttask"), name)
	}
	return obj.(*v1alpha2.InspectTask), nil
}