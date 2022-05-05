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
	"time"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var K8sResourcesChan = make(chan K8SResource)
var PluginsResultsChan = make(chan []kubeeyev1alpha1.PluginsResult)
var RegoRulesListChan = make(chan RegoRulesList)
var ResultChan = make(chan ValidateResult)

type K8SResource struct {
	ServerVersion    string
	CreationTime     time.Time
	APIServerAddress string
	Nodes            *unstructured.UnstructuredList
	NodesCount       int
	Namespaces       *unstructured.UnstructuredList
	NameSpacesCount  int
	NameSpacesList   []string
	Deployments      *unstructured.UnstructuredList
	DaemonSets       *unstructured.UnstructuredList
	StatefulSets     *unstructured.UnstructuredList
	Jobs             *unstructured.UnstructuredList
	CronJobs         *unstructured.UnstructuredList
	WorkloadsCount   int
	Roles            *unstructured.UnstructuredList
	ClusterRoles     *unstructured.UnstructuredList
	Events           *unstructured.UnstructuredList
}

type RegoRulesList struct {
	RegoRules []string
}

type Workload struct {
	Kind       string
	Pod        corev1.Pod
	PodSpec    corev1.PodSpec
	ObjectMeta metav1.Object
}

type ValidateResult struct {
	Name      string
	Namespace string
	Type      string
	Level     string
	Message   string
	Reason    string
}

type ResourceProvider struct {
	ServerVersion   string
	CreationTime    time.Time
	AuditAddress    string
	Nodes           []corev1.Node
	Namespaces      []corev1.Namespace
	Pods            *corev1.PodList
	ConfigMap       []corev1.ConfigMap
	ProblemDetector []corev1.Event
	Workloads       []Workload
}

type ReturnMsg struct {
	what string
}

type Certificate struct {
	Name     string `yaml:"name" json:"name,omitempty"`
	Expires  string `yaml:"expires" json:"expires,omitempty"`
	Residual string `yaml:"residual" json:"residual,omitempty"`
}
