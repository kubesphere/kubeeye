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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"time"

	"github.com/kubesphere/kubeeye-v1alpha2/pkg/conf"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// GetK8SResourcesProvider get kubeconfig by KubernetesAPI, get kubernetes resources by GetK8SResources.
func GetK8SResourcesProvider(ctx context.Context, kubernetesClient *KubernetesClient) error {

	GetK8SResources(ctx, kubernetesClient)
	return nil
}

// TODO
//Add method to excluded namespaces in GetK8SResources.

// GetObjectCounts get kubernetes resources by GroupVersion
func GetObjectCounts(ctx context.Context, kubernetesClient *KubernetesClient, resource string, group string) (*unstructured.UnstructuredList, int, error) {

	var rsourceCount int

	dynamicClient := kubernetesClient.DynamicClient
	listOpts := metav1.ListOptions{}

	resourceGVR := schema.GroupVersionResource{Group: group, Resource: resource, Version: conf.APIVersionV1}
	rsource, err := dynamicClient.Resource(resourceGVR).List(ctx, listOpts)
	if err != nil {
		fmt.Printf("Failed to get Kubernetes %v.\n,error:%s", resource, err)
	}
	if rsource != nil {
		rsourceCount = len(rsource.Items)
	}

	return rsource, rsourceCount, err
}

// GetK8SResources get kubernetes resources by GroupVersionResource, return K8SResource.
func GetK8SResources(ctx context.Context, kubernetesClient *KubernetesClient) K8SResource {
	kubeconfig := kubernetesClient.KubeConfig
	clientSet := kubernetesClient.ClientSet

	var serverVersion string
	var namespacesList []string

	versionInfo, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		fmt.Printf("Failed to get Kubernetes serverVersion.\n")
	}
	if versionInfo != nil {
		serverVersion = versionInfo.Major + "." + versionInfo.Minor
	}

	nodes, nodesCount, err := GetObjectCounts(ctx, kubernetesClient, conf.Nodes, conf.NoGroup)
	if err != nil {
		klog.Error("failed to get nodes and nodesCount")
	}

	namespaces, namespacesCount, _ := GetObjectCounts(ctx, kubernetesClient, conf.Namespaces, conf.NoGroup)
	for _, namespacesItem := range namespaces.Items {
		namespacesList = append(namespacesList, namespacesItem.GetName())
	}

	deployments, deploymentsCount, _ := GetObjectCounts(ctx, kubernetesClient, conf.Deployments, conf.AppsGroup)
	pods, podsCount, _ := GetObjectCounts(ctx, kubernetesClient, conf.Pods, conf.NoGroup)

	daemonSets, daemonSetsCount, _ := GetObjectCounts(ctx, kubernetesClient, conf.Daemonsets, conf.AppsGroup)

	statefulSets, statefulSetsCount, _ := GetObjectCounts(ctx, kubernetesClient, conf.Statefulsets, conf.AppsGroup)

	workloadsCount := deploymentsCount + daemonSetsCount + statefulSetsCount + podsCount

	jobs, _, _ := GetObjectCounts(ctx, kubernetesClient, conf.Jobs, conf.BatchGroup)

	cronjobs, _, _ := GetObjectCounts(ctx, kubernetesClient, conf.Cronjobs, conf.BatchGroup)

	events, _, _ := GetObjectCounts(ctx, kubernetesClient, conf.Events, conf.NoGroup)

	roles, _, _ := GetObjectCounts(ctx, kubernetesClient, conf.Roles, conf.RoleGroup)

	clusterRoles, _, _ := GetObjectCounts(ctx, kubernetesClient, conf.Clusterroles, conf.RoleGroup)

	return K8SResource{
		ServerVersion:    serverVersion,
		CreationTime:     time.Now(),
		APIServerAddress: kubeconfig.Host,
		Nodes:            nodes,
		NodesCount:       nodesCount,
		Namespaces:       namespaces,
		NameSpacesCount:  namespacesCount,
		NameSpacesList:   namespacesList,
		Deployments:      deployments,
		Pods:             pods,
		DaemonSets:       daemonSets,
		StatefulSets:     statefulSets,
		Jobs:             jobs,
		CronJobs:         cronjobs,
		WorkloadsCount:   workloadsCount,
		Roles:            roles,
		ClusterRoles:     clusterRoles,
		Events:           events,
	}
}

func GetNodes(ctx context.Context, clients kubernetes.Interface) []corev1.Node {
	nodeAll, err := clients.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("failed to get nodes", err)
		return nil
	}
	return nodeAll.Items
}
