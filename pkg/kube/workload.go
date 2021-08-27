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
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func LoadWorkloads(ctx context.Context, pods *corev1.PodList, dynamicREST *dynamic.Interface, restMapper *meta.RESTMapper) ([]Workload, error) {
	var workloads []Workload
	for _, pod := range pods.Items {
		workload := getWorkLoad(pod, dynamicREST, restMapper, ctx)
		workloads = append(workloads, workload)
	}
	return removeDuplicateWorkloads(workloads), nil
}

func getWorkLoad(pod corev1.Pod, dynamicREST *dynamic.Interface, restMapper *meta.RESTMapper, ctx context.Context) Workload {
	objectCache := map[string]unstructured.Unstructured{}
	// set default kind to Pod.
	workload := Workload{
		Kind:       "Pod",
		Pod:        pod,
		PodSpec:    pod.Spec,
		ObjectMeta: pod.ObjectMeta.GetObjectMeta(),
	}

	owners := workload.ObjectMeta.GetOwnerReferences()

	for len(owners) > 0 {
		if len(owners) > 1 {
			logrus.Warn("More than 1 owner found")
		}
		if owners[0].Kind == "Node" {
			break
		}
		workload.Kind = owners[0].Kind
		key := fmt.Sprintf("%s/%s/%s", owners[0].Kind, pod.ObjectMeta.GetObjectMeta().GetNamespace(), owners[0].Name)

		abstractObject, ok := objectCache[key]
		if !ok {
			err := cacheAllObjectsOfKind(ctx, owners[0].APIVersion, owners[0].Kind, dynamicREST, restMapper, objectCache)
			if err != nil {
				logrus.Warnf("Error caching objects of Kind %s %v", owners[0].Kind, err)
				break
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				logrus.Errorf("Cache missed %s again", key)
				break
			}
		}
		objMeta, _ := meta.Accessor(&abstractObject)
		workload.ObjectMeta = objMeta
		owners = abstractObject.GetOwnerReferences()
	}
	return workload
}

func cacheAllObjectsOfKind(ctx context.Context, apiVersion string, kind string, dynamicREST *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {

	fqkind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := (*restMapper).RESTMapping(fqkind.GroupKind(), fqkind.Version)
	if err != nil {
		logrus.Warnf("Error retrieving mapping of API %s and Kind %s because of error: %v ", apiVersion, kind, err)
		return err
	}
	objects, err := (*dynamicREST).Resource(mapping.Resource).Namespace("").List(ctx, metav1.ListOptions{})
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

func removeDuplicateWorkloads(workloads []Workload) []Workload {
	workloadsList := make(map[string]Workload)
	for _, workload := range workloads {
		key := workload.ObjectMeta.GetNamespace() + "/" + workload.Kind + workload.ObjectMeta.GetName()
		_, ok := workloadsList[key]
		if !ok {
			workloadsList[key] = workload
		}
	}

	var NewWorkloads []Workload
	for _, workload := range workloadsList {
		NewWorkloads = append(NewWorkloads, workload)
	}

	return NewWorkloads
}
