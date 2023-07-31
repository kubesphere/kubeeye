/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InspectTaskSpec defines the desired state of InspectTask
type InspectTaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName   []string `json:"clusterName,omitempty"`
	Timeout       string   `json:"timeout,omitempty"`
	InspectPolicy Policy   `json:"inspectPolicy,omitempty"`
}

// InspectTaskStatus defines the observed state of InspectTask
type InspectTaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClusterInfo    `json:"clusterInfo,omitempty"`
	JobPhase       []JobPhase  `json:"phase,omitempty"`
	StartTimestamp metav1.Time `json:"startTimestamp,omitempty"`
	EndTimestamp   metav1.Time `json:"endTimestamp,omitempty"`
}

type JobPhase struct {
	JobName string `json:"jobName,omitempty"`
	Phase   Phase  `json:"phase,omitempty"`
}

type JobRule struct {
	JobName  string `json:"jobName,omitempty"`
	RuleType string `json:"ruleType,omitempty"`
	RunRule  []byte `json:"runRule,omitempty"`
}

type Policy string

const (
	TimingPolicy  Policy = "timing"
	InstantPolicy Policy = "instant"
)

type Phase string

const (
	PhasePending   Phase = "Pending"
	PhaseRunning   Phase = "Running"
	PhaseSucceeded Phase = "Succeeded"
	PhaseFailed    Phase = "Failed"
	PhaseUnknown   Phase = "Unknown"
)

type ClusterInfo struct {
	ClusterVersion  string `json:"version,omitempty"`
	NodesCount      int    `json:"nodesCount,omitempty"`
	NamespacesCount int    `json:"namespacesCount,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// InspectTask is the Schema for the InspectTasks API
type InspectTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectTaskSpec   `json:"spec,omitempty"`
	Status InspectTaskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InspectTaskList contains a list of InspectTask
type InspectTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectTask{}, &InspectTaskList{})
}

// kubeeye inspect result
type KubeeyeOpaResult struct {
	ScoreInfo       `json:"scoreInfo,omitempty"`
	ResourceResults []ResourceResult `json:"resourceResults,omitempty"`
	Percent         int              `json:"percent,omitempty"`
	ExtraInfo       `json:"extraInfo,omitempty"`
}

type ResourceResult struct {
	NameSpace    string       `json:"namespace,omitempty"`
	ResourceType string       `json:"resourceType,omitempty"`
	Name         string       `json:"name,omitempty"`
	ResultItems  []ResultItem `json:"resultItems,omitempty"`
}

type ResultItem struct {
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}
type ScoreInfo struct {
	Score     int `json:"score,omitempty"`
	Total     int `json:"total,omitempty"`
	Dangerous int `json:"dangerous,omitempty"`
	Warning   int `json:"warning,omitempty"`
	Ignore    int `json:"ignore,omitempty"`
	Passing   int `json:"passing,omitempty"`
}
type ExtraInfo struct {
	WorkloadsCount int      `json:"workloadsCount,omitempty"`
	NamespacesList []string `json:"namespacesList,omitempty"`
}

type FileChangeResultItem struct {
	FileName string   `json:"fileName,omitempty"`
	Issues   []string `json:"issues,omitempty"`
	Path     string   `json:"path,omitempty"`
}
