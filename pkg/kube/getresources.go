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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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

// GetK8SResources get kubernetes resources by GroupVersionResource, put the resources into the channel K8sResourcesChan, return error.
func GetK8SResources(ctx context.Context, kubernetesClient *KubernetesClient) {
	kubeconfig := kubernetesClient.KubeConfig
	clientSet := kubernetesClient.ClientSet
	dynamicClient := kubernetesClient.DynamicClient
	listOpts := metav1.ListOptions{}

	var serverVersion string
	var nodesCount int
	var namespacesCount int
	var namespacesList []string
	var deploymentsCount int
	var statefulsetsCount int
	var daemonsetsCount int
	var workloadsCount int

	// TODO
	// Implement method to excluded namespace.
	//excludedNamespaces := []string{"kube-system", "kubesphere-system"}
	fieldSelectorString := listOpts.FieldSelector
	//for _, excludedNamespace := range excludedNamespaces {
	//	fieldSelectorString += ",metadata.namespace!=" + excludedNamespace
	//}
	fieldSelector, _ := fields.ParseSelector(fieldSelectorString)
	listOptsExcludedNamespace := metav1.ListOptions{
		FieldSelector: fieldSelectorString,
		LabelSelector: fieldSelector.String(),
	}

	versionInfo, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes serverVersion.\033[0m\n")
		//fmt.Errorf("failed to fetch serverVersion: %s", err.Error())
	}
	if versionInfo != nil {
		serverVersion = versionInfo.Major + "." + versionInfo.Minor
	}

	nodesGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	nodes, err := dynamicClient.Resource(nodesGVR).List(ctx, listOpts)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes nodes.\033[0m\n")
	}
	if nodes != nil {
		nodesCount = len(nodes.Items)
	}

	namespacesGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	namespaces, err := dynamicClient.Resource(namespacesGVR).List(ctx, listOpts)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes namespaces.\033[0m\n")
	}
	if namespaces != nil {
		namespacesCount = len(namespaces.Items)
		for _, namespacesItem := range namespaces.Items {
			namespacesList = append(namespacesList, namespacesItem.GetName())
		}
	}

	deploymentsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := dynamicClient.Resource(deploymentsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes deployments.\033[0m\n")
	}
	if deployments != nil {
		deploymentsCount = len(deployments.Items)
	}

	daemonSetsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	daemonSets, err := dynamicClient.Resource(daemonSetsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes daemonSets.\033[0m\n")
	}
	if daemonSets != nil {
		daemonsetsCount = len(daemonSets.Items)
	}

	statefulSetsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	statefulSets, err := dynamicClient.Resource(statefulSetsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes statefulSets.\033[0m\n")
	}
	if statefulSets != nil {
		statefulsetsCount = len(statefulSets.Items)
	}

	workloadsCount = deploymentsCount + daemonsetsCount + statefulsetsCount

	jobsGVR := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	jobs, err := dynamicClient.Resource(jobsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes jobs.\033[0m\n")
	}

	cronjobsGVR := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}
	cronjobs, err := dynamicClient.Resource(cronjobsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes cronjobs.\033[0m\n")
	}

	eventsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "events"}
	events, err := dynamicClient.Resource(eventsGVR).List(ctx, listOpts)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes events.\033[0m\n")
	}

	rolesGVR := schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}
	roles, err := dynamicClient.Resource(rolesGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes roles.\033[0m\n")
	}

	clusterRolesGVR := schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}
	clusterRoles, err := dynamicClient.Resource(clusterRolesGVR).List(ctx, listOpts)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get Kubernetes clusterroles.\033[0m\n")
	}

	K8sResourcesChan <- K8SResource{
		ServerVersion:    serverVersion,
		CreationTime:     time.Now(),
		APIServerAddress: kubeconfig.Host,
		Nodes:            nodes,
		NodesCount:       nodesCount,
		Namespaces:       namespaces,
		NameSpacesCount:  namespacesCount,
		NameSpacesList:   namespacesList,
		Deployments:      deployments,
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