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
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//var KubeConfig *rest.Config

type KubernetesClient struct {
	KubeConfig    *rest.Config
	ClientSet     kubernetes.Interface
	DynamicClient dynamic.Interface
}

// GetKubeConfig get the kubeconfig from path or by GetConfig
func GetKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	var kubeConfig *rest.Config
	var err error
	if kubeconfigPath != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			kubeConfig, err = config.GetConfig()
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load kubeconfig file from %v", kubeconfigPath)
			}
		}
	} else if kubeconfigPath == "" {
		kubeConfig, err = config.GetConfig()
		if err != nil {
			kubeConfig, err = rest.InClusterConfig()
			if err != nil {
				return nil, errors.Wrap(err, "failed to load kubeconfig file from $HOME/.kube/")
			}
		}
	}
	return kubeConfig, err
}

func GetKubeConfigInCluster() (*rest.Config, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		kubeConfig, err = config.GetConfig()
		if err != nil {
			return nil, err
		}
	}
	return kubeConfig, nil
}

// K8SClients return kubeconfig clientset and dynamicClient.
func (k *KubernetesClient) K8SClients(kubeConfig *rest.Config) (*KubernetesClient, error) {
	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load clientSet")
	}

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load dynamicClient")
	}

	k.ClientSet = clientSet
	k.DynamicClient = dynamicClient
	k.KubeConfig = kubeConfig

	return k, nil
}
