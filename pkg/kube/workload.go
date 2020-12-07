// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type GenericWorkload struct {
	Kind               string
	PodSpec            kubeAPICoreV1.PodSpec
	ObjectMeta         kubeAPIMetaV1.Object
	OriginalObjectJSON []byte
}

func NewGenericWorkload(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericWorkload, error) {
	workload, err := newGenericWorkload(ctx, podResource, dynamicClient, restMapper, objectCache)
	if err != nil {
		return workload, err
	}
	if len(workload.OriginalObjectJSON) == 0 {
		return NewGenericWorkloadFromPod(podResource, podResource)
	}
	return workload, err
}
func NewGenericWorkloadFromPod(podResource kubeAPICoreV1.Pod, originalObject interface{}) (GenericWorkload, error) {
	workload := GenericWorkload{
		Kind:       "Pod",
		PodSpec:    podResource.Spec,
		ObjectMeta: podResource.ObjectMeta.GetObjectMeta(),
	}
	if originalObject != nil {
		bytes, err := json.Marshal(originalObject)
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	}
	return workload, nil
}
func newGenericWorkload(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericWorkload, error) {
	workload, err := NewGenericWorkloadFromPod(podResource, nil)
	if err != nil {
		return workload, err
	}
	// If an owner exists then set the name to the workload.
	// This allows us to handle CRDs creating Workloads or DeploymentConfigs in OpenShift.
	owners := workload.ObjectMeta.GetOwnerReferences()
	lastKey := ""
	for len(owners) > 0 {
		if len(owners) > 1 {
			logrus.Warn("More than 1 owner found")
		}
		firstOwner := owners[0]
		if firstOwner.Kind == "Node" {
			break
		}
		workload.Kind = firstOwner.Kind
		key := fmt.Sprintf("%s/%s/%s", firstOwner.Kind, workload.ObjectMeta.GetNamespace(), firstOwner.Name)
		lastKey = key
		abstractObject, ok := objectCache[key]
		if !ok {
			err = cacheAllObjectsOfKind(ctx, firstOwner.APIVersion, firstOwner.Kind, dynamicClient, restMapper, objectCache)
			if err != nil {
				logrus.Warnf("Error caching objects of Kind %s %v", firstOwner.Kind, err)
				break
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				logrus.Errorf("Cache missed %s again", key)
				break
			}
		}

		objMeta, err := meta.Accessor(&abstractObject)
		if err != nil {
			logrus.Warnf("Error retrieving parent metadata %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload, err
		}
		workload.ObjectMeta = objMeta
		owners = abstractObject.GetOwnerReferences()
	}

	if lastKey != "" {
		bytes, err := json.Marshal(objectCache[lastKey])
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	} else {
		bytes, err := json.Marshal(podResource)
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	}
	return workload, nil
}
func cacheAllObjectsOfKind(ctx context.Context, apiVersion, kind string, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := (*restMapper).RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		logrus.Warnf("Error retrieving mapping of API %s and Kind %s because of error: %v ", apiVersion, kind, err)
		return err
	}

	objects, err := (*dynamicClient).Resource(mapping.Resource).Namespace("").List(ctx, kubeAPIMetaV1.ListOptions{})
	if err != nil {
		logrus.Warnf("Error retrieving parent object API %s and Kind %s because of error: %v ", mapping.Resource.Version, mapping.Resource.Resource, err)
		return err
	}
	for idx, object := range objects.Items {
		key := fmt.Sprintf("%s/%s/%s", object.GetKind(), object.GetNamespace(), object.GetName())
		objectCache[key] = objects.Items[idx]
	}
	return nil
}
