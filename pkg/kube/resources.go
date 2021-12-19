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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)
type ValidateResult struct {
	Name      string
	Namespace string
	Type      string
	Message   string
}

type workloads struct {
	Deployments  []unstructured.Unstructured
	DaemonSets   []unstructured.Unstructured
	StatefulSets []unstructured.Unstructured
	Jobs         []unstructured.Unstructured
	CronJobs     []unstructured.Unstructured
}

type K8SResource struct {
	ServerVersion string
	CreationTime  time.Time
	AuditAddress  string
	Workloads     workloads
	Nodes         []unstructured.Unstructured
	Namespaces    []unstructured.Unstructured
	Roles         []unstructured.Unstructured
	ClusterRoles  []unstructured.Unstructured
	Events        []unstructured.Unstructured
}

type RegoRulesList struct {
	RegoRules []string
}


