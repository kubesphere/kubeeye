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
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var KubeConfig *rest.Config

type KubernetesClient struct {
	kubeconfig      *rest.Config
	ClientSet 		kubernetes.Interface
	DynamicClient   dynamic.Interface
}

func GetKubeConfig(kubeconfigPath string) *rest.Config  {
	if kubeconfigPath != "" {
		kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			fmt.Errorf("failed to load kubernetes config: %s\n", strings.ReplaceAll(err.Error(), "KUBERNETES_MASTER", "KUBECONFIG"))
			os.Exit(1)
		}
		KubeConfig = kubeconfig
	} else {
		kubeconfig, err :=  config.GetConfig()
		if err != nil {
			fmt.Errorf("failed to load kubernetes config: %s\n", strings.ReplaceAll(err.Error(), "KUBERNETES_MASTER", "KUBECONFIG"))
			os.Exit(1)
		}
		KubeConfig = kubeconfig
	}
	return KubeConfig
}

func ClientSet(path string) *kubernetes.Clientset {
	k8sconfig := GetKubeConfig(path)

	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		fmt.Printf("Failed to load config file, reason: %s", err.Error())
		os.Exit(1)
	}
	return clientset
}

func DynamicClient(path string) dynamic.Interface {
	k8sconfig := GetKubeConfig(path)

	dynamicClient, err := dynamic.NewForConfig(k8sconfig)
	if err != nil {
		fmt.Printf("Failed to load config file, reason: %s", err.Error())
		os.Exit(1)
	}
	return dynamicClient
}

func KubernetesAPI(kubeconfigPath string) *KubernetesClient {
	k8sconfig := GetKubeConfig(kubeconfigPath)

	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		fmt.Printf("Failed to load config file, reason: %s", err.Error())
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(k8sconfig)
	if err != nil {
		fmt.Printf("Failed to load config file, reason: %s", err.Error())
		os.Exit(1)
	}
	return &KubernetesClient{
		kubeconfig: k8sconfig,
		ClientSet: clientset,
		DynamicClient: dynamicClient,
	}
}