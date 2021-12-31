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
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// GetK8SResourcesProvider get kubeconfig by KubernetesAPI, get kubernetes resources by GetK8SResources.
func GetK8SResourcesProvider(ctx context.Context, kubeconfig string) {
	// get kubernetes client
	kubernetesClient := KubernetesAPI(kubeconfig)
	err := GetK8SResources(ctx, kubernetesClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GetK8SResources get kubernetes resources by GroupVersionResource, put the resources into the channel K8sResourcesChan, return error.
func GetK8SResources(ctx context.Context, kubernetesClient *KubernetesClient) error {
	kubeconfig := kubernetesClient.kubeconfig
	clientSet := kubernetesClient.ClientSet
	dynamicClient := kubernetesClient.DynamicClient
	listOpts := metav1.ListOptions{}
	excludedNamespaces := []string{"kube-system", "kubesphere-system"}
	fieldSelectorString := listOpts.FieldSelector
	for _, excludedNamespace := range excludedNamespaces {
		fieldSelectorString += ",metadata.namespace!=" + excludedNamespace
	}
	fieldSelector, _ := fields.ParseSelector(fieldSelectorString)
	listOptsExcludedNamespace := metav1.ListOptions{
		FieldSelector: fieldSelectorString,
		LabelSelector: fieldSelector.String(),
	}

	serverVersion, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		err := fmt.Errorf("failed to fetch serverVersion: %s", err.Error())
		return err
	}

	nodesGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	nodes, err := dynamicClient.Resource(nodesGVR).List(ctx, listOpts)
	if err != nil {
		err := fmt.Errorf("failed to fetch nodes: %s", err.Error())
		return err
	}

	namespacesGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	namespaces, err := dynamicClient.Resource(namespacesGVR).List(ctx, listOpts)
	if err != nil {
		err := fmt.Errorf("failed to fetch namespaces: %s", err.Error())
		return err
	}

	deploymentsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := dynamicClient.Resource(deploymentsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch deployments: %s", err.Error())
		return err
	}

	daemonSetsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	daemonSets, err := dynamicClient.Resource(daemonSetsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch daemonSets: %s", err.Error())
		return err
	}

	statefulSetsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
	statefulSets, err := dynamicClient.Resource(statefulSetsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch statefulSets: %s", err.Error())
		return err
	}

	jobsGVR := schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	jobs, err := dynamicClient.Resource(jobsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch jobs: %s", err.Error())
		return err
	}

	cronjobsGVR := schema.GroupVersionResource{Group: "batch", Version: "v1beta1", Resource: "cronjobs"}
	cronjobs, err := dynamicClient.Resource(cronjobsGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch cronjobs: %s", err.Error())
		return err
	}

	eventsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "events"}
	events, err := dynamicClient.Resource(eventsGVR).List(ctx, listOpts)
	if err != nil {
		err := fmt.Errorf("failed to fetch events: %s", err.Error())
		return err
	}

	rolesGVR := schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Resource: "roles"}
	roles, err := dynamicClient.Resource(rolesGVR).List(ctx, listOptsExcludedNamespace)
	if err != nil {
		err := fmt.Errorf("failed to fetch clusterRoles: %s", err.Error())
		return err
	}

	clusterRolesGVR := schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Resource: "clusterroles"}
	clusterRoles, err := dynamicClient.Resource(clusterRolesGVR).List(ctx, listOpts)
	if err != nil {
		err := fmt.Errorf("failed to fetch clusterRoles: %s", err.Error())
		return err
	}

	K8sResourcesChan <- K8SResource{
		ServerVersion: serverVersion.Major + "." + serverVersion.Minor,
		CreationTime:  time.Now(),
		AuditAddress:  kubeconfig.Host,
		Nodes:         nodes.Items,
		Namespaces:    namespaces.Items,
		Deployments:   deployments.Items,
		DaemonSets:    daemonSets.Items,
		StatefulSets:  statefulSets.Items,
		Jobs:          jobs.Items,
		CronJobs:      cronjobs.Items,
		Roles:         roles.Items,
		ClusterRoles:  clusterRoles.Items,
		Events:        events.Items,
	}
	return nil
}
