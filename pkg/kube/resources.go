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
	"time"

	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ResourceProvider struct {
	ServerVersion   string
	CreationTime    time.Time
	AuditAddress    string
	Nodes           []corev1.Node
	Namespaces      []corev1.Namespace
	Pods            []corev1.Pod
	ComponentStatus []corev1.ComponentStatus
	ConfigMap       []corev1.ConfigMap
	ProblemDetector []corev1.Event
	Controllers     []GenericWorkload
}

func CreateResourceProvider(ctx context.Context) (*ResourceProvider, error) {
	return CreateResourceProviderFromCluster(ctx)
}

//Get kubeConfig
func CreateResourceProviderFromCluster(ctx context.Context) (*ResourceProvider, error) {
	kubeConf, configError := config.GetConfig()
	if configError != nil {
		logrus.Errorf("Error fetching KubeConfig: %v", configError)
		return nil, configError
	}

	api, err1 := kubernetes.NewForConfig(kubeConf)
	if err1 != nil {
		logrus.Errorf("Error fetching api: %v", err1)
		return nil, err1
	}

	dynamicInterface, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error fetching dynamicInterface: %v", err)
		return nil, err
	}
	return CreateResourceProviderFromAPI(ctx, api, kubeConf.Host, &dynamicInterface)
}

//Get serverVersion, nodes, namespaces, pods, problemDetectors, componentStatus, controllers
func CreateResourceProviderFromAPI(ctx context.Context, kube kubernetes.Interface, auditAddress string, dynamic *dynamic.Interface) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}
	//var configmap = []corev1.ConfigMap{}
	//configmap.Data =
	serverVersion, err := kube.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching serverVersion: %v", err)
		return nil, err
	}

	//kube.CoreV1().ConfigMaps("").Create(ctx,configmap,listOpts)

	nodes, err := kube.CoreV1().Nodes().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching nodes: %v", err)
		return nil, err
	}
	namespaces, err := kube.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching namespaces: %v", err)
		return nil, err
	}
	pods, err := kube.CoreV1().Pods("").List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching pods: %v", err)
		return nil, err
	}

	problemDetectors, _ := kube.CoreV1().Events("").List(ctx, listOpts)

	componentStatus, err := kube.CoreV1().ComponentStatuses().List(ctx, listOpts)
	resources, err := restmapper.GetAPIGroupResources(kube.Discovery())
	if err != nil {
		logrus.Errorf("Error fetching resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(resources)

	objectCache := map[string]unstructured.Unstructured{}

	controllers, err := LoadControllers(ctx, pods.Items, dynamic, &restMapper, objectCache)
	if err != nil {
		logrus.Errorf("Error loading controllers from pods: %v", err)
		return nil, err
	}

	api := ResourceProvider{
		ServerVersion:   serverVersion.Major + "." + serverVersion.Minor,
		AuditAddress:    auditAddress,
		CreationTime:    time.Now(),
		Nodes:           nodes.Items,
		Namespaces:      namespaces.Items,
		Pods:            pods.Items,
		ComponentStatus: componentStatus.Items,
		ProblemDetector: problemDetectors.Items,
		Controllers:     controllers,
	}
	return &api, nil
}

func LoadControllers(ctx context.Context, pods []corev1.Pod, d *dynamic.Interface, m *meta.RESTMapper, cache map[string]unstructured.Unstructured) ([]GenericWorkload, error) {
	interfaces := []GenericWorkload{}
	deduped := map[string]corev1.Pod{}
	for _, pod := range pods {
		owners := pod.ObjectMeta.OwnerReferences
		if len(owners) == 0 {
			deduped[pod.ObjectMeta.Namespace+"/Pod/"+pod.ObjectMeta.Name] = pod
			continue
		}
		deduped[pod.ObjectMeta.Namespace+"/"+owners[0].Kind+"/"+pod.ObjectMeta.Name] = pod
	}
	for _, pod := range deduped {
		workload, err := NewGenericWorkload(ctx, pod, d, m, cache)
		if err != nil {
			return nil, err
		}
		interfaces = append(interfaces, workload)
	}
	return deduplicateControllers(interfaces), nil
}
func deduplicateControllers(inputControllers []GenericWorkload) []GenericWorkload {
	controllerMap := make(map[string]GenericWorkload)
	for _, controller := range inputControllers {
		key := controller.ObjectMeta.GetNamespace() + "/" + controller.Kind + "/" + controller.ObjectMeta.GetName()
		oldController, ok := controllerMap[key]
		if !ok || controller.ObjectMeta.GetCreationTimestamp().Time.After(oldController.ObjectMeta.GetCreationTimestamp().Time) {
			controllerMap[key] = controller
		}
	}
	results := make([]GenericWorkload, 0)
	for _, controller := range controllerMap {
		results = append(results, controller)
	}
	return results
}
