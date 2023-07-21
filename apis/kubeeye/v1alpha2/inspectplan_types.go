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

type Auditor string

const (
	AuditorKubeeye   Auditor = "kubeeye"
	Auditorkubebench Auditor = "kubebench"
)

// InspectPlanSpec defines the desired state of InspectPlan
type InspectPlanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Schedule    *string   `json:"schedule,omitempty"`
	Suspend     bool      `json:"suspend,omitempty"`
	Timeout     string    `json:"timeout,omitempty"`
	Tag         string    `json:"tag,omitempty"`
	RuleNames   []string  `json:"ruleNames,omitempty"`
	MaxTasks    int       `json:"maxTasks,omitempty"`
	ClusterName []*string `json:"clusterName,omitempty"`
	KubeConfig  string    `json:"kubeConfig,omitempty"`
}

// InspectPlanStatus defines the observed state of InspectPlan
type InspectPlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LastScheduleTime metav1.Time `json:"lastScheduleTime,omitempty"`
	LastTaskName     string      `json:"lastTaskName,omitempty"`
	TaskNames        []string    `json:"TaskNames,omitempty"`
	LastTaskStatus   Phase       `json:"lastTaskStatus,omitempty"`
	NextScheduleTime metav1.Time `json:"nextScheduleTime,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// InspectPlan is the Schema for the InspectPlans API
type InspectPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InspectPlanSpec   `json:"spec,omitempty"`
	Status InspectPlanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InspectPlanList contains a list of InspectPlan
type InspectPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InspectPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InspectPlan{}, &InspectPlanList{})
}
