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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Workload struct {
	Kind       string
	Pod        corev1.Pod
	PodSpec    corev1.PodSpec
	ObjectMeta metav1.Object
}

type ResourceProvider struct {
	ServerVersion   string
	CreationTime    time.Time
	AuditAddress    string
	Nodes           []corev1.Node
	Namespaces      []corev1.Namespace
	Pods            *corev1.PodList
	ComponentStatus []corev1.ComponentStatus
	ConfigMap       []corev1.ConfigMap
	ProblemDetector []corev1.Event
	Workloads       []Workload
}

func CreateResourceProvider(ctx context.Context) (*ResourceProvider, error) {
	return SetClient(ctx)
}

//Get kubeConfig
func SetClient(ctx context.Context) (*ResourceProvider, error) {
	kubeConf, configError := config.GetConfig()
	if configError != nil {
		logrus.Errorf("Error fetching KubeConfig: %v", configError)
		return nil, configError
	}

	clientSet, err1 := kubernetes.NewForConfig(kubeConf)
	if err1 != nil {
		logrus.Errorf("Error fetching api: %v", err1)
		return nil, err1
	}

	dynamicREST, err := dynamic.NewForConfig(kubeConf)
	if err != nil {
		logrus.Errorf("Error fetching dynamicInterface: %v", err)
		return nil, err
	}
	return GetResources(ctx, clientSet, kubeConf.Host, &dynamicREST)
}

//Get serverVersion, nodes, namespaces, pods, problemDetectors, componentStatus, controllers
func GetResources(ctx context.Context, clientSet kubernetes.Interface, auditAddress string, dynamicREST *dynamic.Interface) (*ResourceProvider, error) {
	listOpts := metav1.ListOptions{}

	serverVersion, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("Error fetching serverVersion: %v", err)
		return nil, err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching nodes: %v", err)
		return nil, err
	}

	namespaces, err := clientSet.CoreV1().Namespaces().List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching namespaces: %v", err)
		return nil, err
	}

	pods, err := clientSet.CoreV1().Pods("").List(ctx, listOpts)
	if err != nil {
		logrus.Errorf("Error fetching pods: %v", err)
		return nil, err
	}

	problemDetectors, _ := clientSet.CoreV1().Events("").List(ctx, listOpts)

	//componentStatus, err := clientSet.CoreV1().ComponentStatuses().List(ctx, listOpts)
	APIGroupResources, err := restmapper.GetAPIGroupResources(clientSet.Discovery())
	if err != nil {
		logrus.Errorf("Error fetching resources: %v", err)
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(APIGroupResources)

	workloads, err := LoadWorkloads(ctx, pods, dynamicREST, &restMapper)
	if err != nil {
		logrus.Errorf("Error loading controllers from pods: %v", err)
		return nil, err
	}

	resources := ResourceProvider{
		ServerVersion:   serverVersion.Major + "." + serverVersion.Minor,
		AuditAddress:    auditAddress,
		CreationTime:    time.Now(),
		Nodes:           nodes.Items,
		Namespaces:      namespaces.Items,
		Pods:            pods,
		ProblemDetector: problemDetectors.Items,
		Workloads:       workloads,
	}
	return &resources, nil
}
