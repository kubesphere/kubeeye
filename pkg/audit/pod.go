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

package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/open-policy-agent/opa/rego"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ValidateResources Validate kubernetes cluster Resources, put the results into channels.
func ValidateResources(ctx context.Context) {

	defer close(kube.RegoRulesListChan)
	// get the rego rules from channel RegoRulesListChan.
	regoRulesList := <-kube.RegoRulesListChan

	defer close(kube.K8sResourcesChan)
	// get the kubernetes resources from the channel K8sResourcesChan.
	k8sResources := <-kube.K8sResourcesChan

	var wg sync.WaitGroup
	wg.Add(9)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateDeployments(ctx, kubeResources.Deployments, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateDaemonSets(ctx, kubeResources.DaemonSets, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateStatefulSets(ctx, kubeResources.StatefulSets, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateJobs(ctx, kubeResources.Jobs, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateCronJobs(ctx, kubeResources.CronJobs, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateRoles(ctx, kubeResources.Roles, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateClusterRoles(ctx, kubeResources.ClusterRoles, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateNodes(ctx, kubeResources.Nodes, regoRulesList)
	}(ctx, k8sResources, regoRulesList)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList) {
		defer wg.Done()
		ValidateEvents(ctx, kubeResources.Events, regoRulesList)
	}(ctx, k8sResources, regoRulesList)
	wg.Wait()
	defer close(kube.ValidateResultsChan)
}

// ValidateDeployments validate deployments of kubernetes by ValidateK8SResource, put the results into the channel DeploymentsResultsChan.
func ValidateDeployments(ctx context.Context, deployments []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateDeploymentsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, deployment := range deployments {
		if validateResults, found := ValidateK8SResource(ctx, deployment, regoRulesList, queryRule); found {
			validateDeploymentsResults.ValidateResults = append(validateDeploymentsResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateDeploymentsResults
}

// ValidateDaemonSets validate daemonSets of kubernetes by ValidateK8SResource, put the results into the channel validateDaemonSetsResults.
func ValidateDaemonSets(ctx context.Context, daemonSets []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateDaemonSetsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, daemonSet := range daemonSets {
		if validateResults, found := ValidateK8SResource(ctx, daemonSet, regoRulesList, queryRule); found {
			validateDaemonSetsResults.ValidateResults = append(validateDaemonSetsResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateDaemonSetsResults
}

// ValidateStatefulSets validate StatefulSets of kubernetes by ValidateK8SResource, put the results into the channel StatefulSetsResultsChan.
func ValidateStatefulSets(ctx context.Context, statefulSets []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateStatefulSetsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, statefulSet := range statefulSets {
		if validateResults, found := ValidateK8SResource(ctx, statefulSet, regoRulesList, queryRule); found {
			validateStatefulSetsResults.ValidateResults = append(validateStatefulSetsResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateStatefulSetsResults
}

// ValidateJobs validate Jobs of kubernetes by ValidateK8SResource, put the results into the channel JobsResultsChan.
func ValidateJobs(ctx context.Context, jobs []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateReplicaJobs kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, job := range jobs {
		if validateResults, found := ValidateK8SResource(ctx, job, regoRulesList, queryRule); found {
			validateReplicaJobs.ValidateResults = append(validateReplicaJobs.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateReplicaJobs
}

// ValidateCronJobs validate cronjobs of kubernetes by ValidateK8SResource, put the results into the channel CronjobsResultsChan.
func ValidateCronJobs(ctx context.Context, cronjobs []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateCronjobsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, cronjob := range cronjobs {
		if validateResults, found := ValidateK8SResource(ctx, cronjob, regoRulesList, queryRule); found {
			validateCronjobsResults.ValidateResults = append(validateCronjobsResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateCronjobsResults
}

// ValidateRoles validate roles of kubernetes by ValidateK8SResource, put the results into the channel RolesResultsChan.
func ValidateRoles(ctx context.Context, roles []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateRolesResults kube.ValidateResults
	queryRule := "data.kubeeye_RBAC_rego"
	for _, role := range roles {
		if validateResults, found := ValidateK8SResource(ctx, role, regoRulesList, queryRule); found {
			validateRolesResults.ValidateResults = append(validateRolesResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateRolesResults
}

// ValidateClusterRoles validate clusterroles of kubernetes by ValidateK8SResource, put the results into the channel ClusterRolesResultsChan.
func ValidateClusterRoles(ctx context.Context, clusterroles []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateClusterRolesResults kube.ValidateResults
	queryRule := "data.kubeeye_RBAC_rego"
	for _, clusterrole := range clusterroles {
		if validateResults, found := ValidateK8SResource(ctx, clusterrole, regoRulesList, queryRule); found {
			validateClusterRolesResults.ValidateResults = append(validateClusterRolesResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateClusterRolesResults
}

func ValidateNodes(ctx context.Context, nodes []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateNodesResults kube.ValidateResults
	queryRule := "data.kubeeye_nodes_rego"
	for _, node := range nodes {
		if validateResults, found := ValidateK8SResource(ctx, node, regoRulesList, queryRule); found {
			validateNodesResults.ValidateResults = append(validateNodesResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateNodesResults
}

func ValidateEvents(ctx context.Context, events []unstructured.Unstructured, regoRulesList kube.RegoRulesList) {
	var validateEventsResults kube.ValidateResults
	queryRule := "data.kubeeye_events_rego"
	for _, clusterrole := range events {
		if validateResults, found := ValidateK8SResource(ctx, clusterrole, regoRulesList, queryRule); found {
			validateEventsResults.ValidateResults = append(validateEventsResults.ValidateResults, validateResults)
		}
	}
	kube.ValidateResultsChan <- validateEventsResults
}

// ValidateK8SResource validate kubernetes resource by rego, return the validate results.
func ValidateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList kube.RegoRulesList, queryRule string) (kube.ResultReceiver, bool) {
	var resultReceiver kube.ResultReceiver
	find := false
	for _, regoRule := range regoRulesList.RegoRules {
		//queryRule := "data.kubeeye_workloads_rego"
		query, err := rego.New(rego.Query(queryRule), rego.Module("examples.rego", regoRule)).PrepareForEval(ctx)
		if err != nil {
			err := fmt.Errorf("failed to parse rego input: %s", err.Error())
			fmt.Println(err)
			os.Exit(1)
		}
		regoResults, err := query.Eval(ctx, rego.EvalInput(resource))
		if err != nil {
			err := fmt.Errorf("failed to validate resource: %s", err.Error())
			fmt.Println(err)
			os.Exit(1)
		}
		for _, regoResult := range regoResults {
			for key, _ := range regoResult.Expressions {

				for _, validateResult := range regoResult.Expressions[key].Value.(map[string]interface{}) {
					var results []kube.ValidateResult
					jsonresult, err := json.Marshal(validateResult)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					if err := json.Unmarshal(jsonresult, &results); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					for _, result := range results {
						find = true
						if result.Type == "ClusterRole" {
							resultReceiver.Name = result.Name
							resultReceiver.Type = result.Type
							resultReceiver.Message = append(resultReceiver.Message, result.Message)
						} else {
							resultReceiver.Name = result.Name
							resultReceiver.Namespace = result.Namespace
							resultReceiver.Type = result.Type
							resultReceiver.Message = append(resultReceiver.Message, result.Message)
						}

					}
				}
			}
		}
	}
	return resultReceiver, find
}
