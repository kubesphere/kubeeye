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

// InspectRuleSpec defines the desired state of InspectRule
type InspectRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PrometheusEndpoint string              `json:"prometheusEndpoint,omitempty"`
	Opas               []OpaRule           `json:"opas,omitempty"`
	Prometheus         []PrometheusRule    `json:"prometheus,omitempty"`
	FileChange         []FileChangeRule    `json:"fileChange,omitempty" `
	Sysctl             []SysRule           `json:"sysctl,omitempty"`
	Systemd            []SysRule           `json:"systemd,omitempty"`
	FileFilter         []FileFilterRule    `json:"fileFilter,omitempty"`
	CustomCommand      []CustomCommandRule `json:"customCommand,omitempty"`
	NodeInfo           []NodeInfo          `json:"nodeInfo,omitempty"`
	Component          *string             `json:"component,omitempty"`
}
type RuleItemBases struct {
	Name  string  `json:"name,omitempty"`
	Rule  *string `json:"rule,omitempty"`
	Desc  string  `json:"desc,omitempty"`
	Level Level   `json:"level,omitempty"`
}

type Node struct {
	NodeName     *string           `json:"nodeName,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type NodeInfo struct {
	RuleItemBases `json:",inline"`
	Node          `json:",inline"`
}

type FileFilterRule struct {
	RuleItemBases `json:",inline"`
	Path          string `json:"path,omitempty"`
	Node          `json:",inline"`
}

type SysRule struct {
	RuleItemBases `json:",inline"`
	Node          `json:",inline"`
}

type OpaRule struct {
	RuleItemBases `json:",inline"`
	Module        string `json:"module,omitempty"`
}
type PrometheusRule struct {
	RuleItemBases `json:",inline"`
	Endpoint      string `json:"endpoint,omitempty"`
}

type FileChangeRule struct {
	RuleItemBases `json:",inline"`
	Path          string `json:"path,omitempty"`
	Node          `json:",inline"`
}
type CustomCommandRule struct {
	RuleItemBases `json:",inline"`
	Command       string `json:"command,omitempty"`
	Node          `json:",inline"`
}

type State string

const (
	StartImport    State = "StartImport"
	ImportComplete State = "ImportComplete"
)

// InspectRuleStatus defines the observed state of InspectRule
type InspectRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	StartImportTime metav1.Time    `json:"startImportTime,omitempty"`
	EndImportTime   metav1.Time    `json:"endImportTime,omitempty"`
	State           State          `json:"state,omitempty"`
	LevelCount      map[Level]*int `json:"levelCount,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// InspectRule is the Schema for the InspectRule API
type InspectRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectRuleSpec   `json:"spec,omitempty"`
	Status InspectRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InspectRuleList contains a list of InspectRule
type InspectRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectRule{}, &InspectRuleList{})
}
