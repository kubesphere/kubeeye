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

// ClusterInsightSpec defines the desired state of ClusterInsight
type ClusterInsightSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AuditPeriod string `json:"auditPeriod"`
}

// ClusterInsightStatus defines the observed state of ClusterInsight
type ClusterInsightStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AfterTime    metav1.Time `json:"afterTime,omitempty"`
	ClusterInfo  ClusterInfo `json:"clusterInfo,omitempty"`
	AuditResults AuditResult `json:"auditResults,omitempty"`
}

type ClusterInfo struct {
	ClusterVersion  string   `json:"version,omitempty"`
	NodesCount      int      `json:"nodesCount,omitempty"`
	NamespacesCount int      `json:"namespacesCount,omitempty"`
	WorkloadsCount  int      `json:"workloadsCount,omitempty"`
	NamespacesList  []string `json:"namespacesList,omitempty"`
}

type AuditResult struct {
	ClusterScore ScoreReceiver     `json:"namespace,omitempty"`
	Results      []ValidateResults `json:"auditResults,omitempty"`
}

type ResultInfos struct {
	Namespace     string          `json:"namespace"`
	ResourceInfos []ResourceInfos `json:"resourceInfos"`
}

type ResourceInfos struct {
	Name        string        `json:"name,omitempty"`
	ResultItems []ResultItems `json:"items"`
}

type ResultItems struct {
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type ScoreReceiver struct {
	Score     int32 `json:"score,omitempty"`
	Passing   int32 `json:"passing"`
	Warning   int32 `json:"warning"`
	Dangerous int32 `json:"dangerous"`
	Total     int32 `json:"total"`
}

type ValidateResults struct {
	ResourcesType string        `json:"resourcesType,omitempty"`
	ResultInfos   []ResultInfos `json:"resultInfos,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ClusterInsight is the Schema for the clusterinsights API
type ClusterInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterInsightSpec   `json:"spec,omitempty"`
	Status ClusterInsightStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterInsightList contains a list of ClusterInsight
type ClusterInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterInsight{}, &ClusterInsightList{})
}
