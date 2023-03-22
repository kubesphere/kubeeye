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

// InspectRulesSpec defines the desired state of InspectRules
type InspectRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +operator-sdk:validation:Optional
	PrometheusEndpoint string `json:"prometheusEndpoint,omitempty" yaml:"prometheusEndpoint"`
	// +kubebuilder:validation:MinItems=1
	Rules []RuleItems `yaml:"rules,omitempty" json:"rules,omitempty"`
}

type State string

const (
	StartImport   State = "StartImport"
	ImportSuccess State = "importSuccess"
)

type RuleItems struct {
	RuleName           string   `yaml:"ruleName,omitempty" json:"ruleName,omitempty" `
	Desc               string   `yaml:"desc,omitempty" json:"desc,omitempty"`
	Opa                string   `yaml:"opa,omitempty" json:"opa,omitempty"`
	Prometheus         string   `yaml:"prometheus,omitempty" json:"prometheus,omitempty"`
	PrometheusEndpoint string   `yaml:"prometheusEndpoint,omitempty" json:"prometheusEndpoint,omitempty"`
	Priority           string   `yaml:"priority,omitempty" json:"priority,omitempty"`
	Tags               []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// InspectRulesStatus defines the observed state of InspectRules
type InspectRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ImportTime metav1.Time `yaml:"importTime,omitempty" json:"importTime,omitempty"`

	State State `yaml:"state,omitempty" json:"state,omitempty"`

	RuleCount map[string]int `yaml:"ruleCount,omitempty" json:"ruleCount,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InspectRules is the Schema for the InspectRules API
type InspectRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectRulesSpec   `json:"spec,omitempty"`
	Status InspectRulesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InspectRulesList contains a list of InspectRules
type InspectRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectRules{}, &InspectRulesList{})
}
