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

// InspectResultSpec defines the desired state of InspectResult
type InspectResultSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	InspectRuleTotal map[string]int            `json:"inspectRuleTotal,omitempty"`
	PrometheusResult [][]map[string]string     `json:"prometheusResult,omitempty"`
	OpaResult        KubeeyeOpaResult          `json:"opaResult,omitempty"`
	NodeInfoResult   map[string]NodeInfoResult `json:"nodeInfoResult,omitempty"`
	ComponentResult  []ComponentResultItem     `json:"componentResult,omitempty"`
}

// InspectResultStatus defines the observed state of InspectResult
type InspectResultStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Complete      bool           `json:"complete,omitempty"`
	Policy        Policy         `json:"policy,omitempty"`
	Duration      string         `json:"duration,omitempty"`
	TaskStartTime string         `json:"taskStartTime,omitempty"`
	TaskEndTime   string         `json:"taskEndTime,omitempty"`
	Level         map[Level]*int `json:"level,omitempty"`
}

type Level string

const (
	DangerLevel  Level = "danger"
	WarningLevel Level = "warning"
	IgnoreLevel  Level = "ignore"
)

type NodeInfoResult struct {
	NodeInfo         map[string]string      `json:"nodeInfo,omitempty"`
	FileChangeResult []FileChangeResultItem `json:"fileChangeResult,omitempty"`
	FileFilterResult []FileChangeResultItem `json:"fileFilterResult,omitempty"`
	SysctlResult     []NodeResultItem       `json:"sysctlResult,omitempty"`
	SystemdResult    []NodeResultItem       `json:"systemdResult,omitempty"`
	CommandResult    []CommandResultItem    `json:"commandResult,omitempty"`
}

type FileChangeResultItem struct {
	FileName string   `json:"fileName,omitempty"`
	Issues   []string `json:"issues,omitempty"`
	Path     string   `json:"path,omitempty"`
	Level    Level    `json:"level,omitempty"`
}
type NodeResultItem struct {
	Name   string  `json:"name,omitempty"`
	Assert bool    `json:"assert,omitempty"`
	Value  *string `json:"value,omitempty"`
	Level  Level   `json:"level,omitempty"`
}
type ComponentResultItem struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	Level     Level  `json:"level,omitempty"`
}
type CommandResultItem struct {
	Name    string `json:"name,omitempty"`
	Command string `json:"command,omitempty"`
	Level   Level  `json:"level,omitempty"`
	Assert  bool   `json:"assert,omitempty"`
	Value   string `json:"value,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// InspectResult is the Schema for the inspectresults API
type InspectResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectResultSpec   `json:"spec,omitempty"`
	Status InspectResultStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InspectResultList contains a list of InspectResult
type InspectResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectResult `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectResult{}, &InspectResultList{})
}
