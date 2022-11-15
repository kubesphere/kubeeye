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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Auditor string

const (
	AuditorKubeeye  Auditor = "kubeeye"
	Auditorkubebench Auditor= "kubebench"
)

// AuditPlanSpec defines the desired state of AuditPlan
type AuditPlanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Schedule string `json:"schedule"`
	Suspend  bool   `json:"suspend,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Auditors []Auditor `json:"auditors"` // like "kubeeye,kubebench"

	Timeout string `json:"timeout,omitempty"`
}

// AuditPlanStatus defines the observed state of AuditPlan
type AuditPlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastScheduleTime metav1.Time `json:"lastScheduleTime"`
	LastTaskName     string       `json:"lastTaskName"`
}


// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuditPlan is the Schema for the auditplans API
type AuditPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditPlanSpec   `json:"spec,omitempty"`
	Status AuditPlanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuditPlanList contains a list of AuditPlan
type AuditPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuditPlan{}, &AuditPlanList{})
}
