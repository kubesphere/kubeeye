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
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var KubeConfig *rest.Config

type KubernetesClient struct {
	kubeconfig    *rest.Config
	ClientSet     kubernetes.Interface
	DynamicClient dynamic.Interface
}

// GetKubeConfig get the kubeconfig from path or by GetConfig
func GetKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	execEnv := os.Getenv("EXEC_ENV")
	if execEnv == "K8SENV" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		KubeConfig = config
	} else if kubeconfigPath != "" && execEnv != "K8SENV" {
		kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			err = fmt.Errorf("failed to load kubernetes config: %s\n", strings.ReplaceAll(err.Error(), "KUBERNETES_MASTER", "KUBECONFIG"))
			return nil, err
		}
		KubeConfig = kubeconfig
	} else if kubeconfigPath == "" && execEnv != "K8SENV" {
		kubeconfig, err := config.GetConfig()
		if err != nil {
			err = fmt.Errorf("failed to load kubernetes config: %s\n", strings.ReplaceAll(err.Error(), "KUBERNETES_MASTER", "KUBECONFIG"))
			return nil, err
		}
		KubeConfig = kubeconfig
	}
	return KubeConfig, nil
}

// ClientSet return clientset
func ClientSet(path string) (*kubernetes.Clientset, error) {
	k8sconfig, err := GetKubeConfig(path)

	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		err := fmt.Errorf("Failed to load config file, reason: %s", err.Error())
		return nil, err
	}
	return clientset, nil
}

// DynamicClient return dynamicClient
func DynamicClient(path string) (dynamic.Interface, error) {
	k8sconfig, err := GetKubeConfig(path)

	dynamicClient, err := dynamic.NewForConfig(k8sconfig)
	if err != nil {
		err := fmt.Errorf("Failed to load config file, reason: %s", err.Error())
		return nil, err
	}
	return dynamicClient, nil
}

// KubernetesAPI return kubeconfig clientset and dynamicClient.
func KubernetesAPI(kubeconfigPath string) (*KubernetesClient, error) {
	k8sconfig, err := GetKubeConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		err := fmt.Errorf("Failed to load config file, reason: %s", err.Error())
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(k8sconfig)
	if err != nil {
		err := fmt.Errorf("Failed to load config file, reason: %s", err.Error())
		return nil, err
	}
	return &KubernetesClient{
		kubeconfig:    k8sconfig,
		ClientSet:     clientset,
		DynamicClient: dynamicClient,
	}, nil
}
