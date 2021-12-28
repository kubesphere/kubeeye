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
func ValidateResources(ctx context.Context) <-chan kube.ValidateResults{
	defer close(kube.RegoRulesListChan)
	// get the rego rules from channel RegoRulesListChan.
	regoRulesList := <-kube.RegoRulesListChan

	defer close(kube.K8sResourcesChan)
	// get the kubernetes resources from the channel K8sResourcesChan.
	k8sResources := <-kube.K8sResourcesChan
    
	resultChan := make(chan kube.ValidateResults)
	var wg sync.WaitGroup
	wg.Add(9)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateDeployments(ctx, kubeResources.Deployments, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateDaemonSets(ctx, kubeResources.DaemonSets, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateStatefulSets(ctx, kubeResources.StatefulSets, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateJobs(ctx, kubeResources.Jobs, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateCronJobs(ctx, kubeResources.CronJobs, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateRoles(ctx, kubeResources.Roles, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateClusterRoles(ctx, kubeResources.ClusterRoles, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateNodes(ctx, kubeResources.Nodes, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)

	go func(ctx context.Context, kubeResources kube.K8SResource, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
		defer wg.Done()
		ValidateEvents(ctx, kubeResources.Events, regoRulesList, resultChan)
	}(ctx, k8sResources, regoRulesList, resultChan)
	go func(){
		defer close(resultChan)
		wg.Wait()
	}()
	return resultChan
}

// ValidateDeployments validate deployments of kubernetes by ValidateK8SResource, put the results into the channel DeploymentsResultsChan.
func ValidateDeployments(ctx context.Context, deployments []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateDeploymentsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, deployment := range deployments {
		validateResults := ValidateK8SResource(ctx, deployment, regoRulesList, queryRule)
		validateDeploymentsResults.ValidateResults = append(validateDeploymentsResults.ValidateResults, validateResults)
	}
	resultChan <- validateDeploymentsResults

}

// ValidateDaemonSets validate daemonSets of kubernetes by ValidateK8SResource, put the results into the channel validateDaemonSetsResults.
func ValidateDaemonSets(ctx context.Context, daemonSets []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateDaemonSetsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, daemonSet := range daemonSets {
		validateResults := ValidateK8SResource(ctx, daemonSet, regoRulesList, queryRule)
		validateDaemonSetsResults.ValidateResults = append(validateDaemonSetsResults.ValidateResults, validateResults)
	}
	resultChan <- validateDaemonSetsResults

}

// ValidateStatefulSets validate StatefulSets of kubernetes by ValidateK8SResource, put the results into the channel StatefulSetsResultsChan.
func ValidateStatefulSets(ctx context.Context, statefulSets []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateStatefulSetsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, statefulSet := range statefulSets {
		validateResults := ValidateK8SResource(ctx, statefulSet, regoRulesList, queryRule)
		validateStatefulSetsResults.ValidateResults = append(validateStatefulSetsResults.ValidateResults, validateResults)
	}
	resultChan <- validateStatefulSetsResults
}

// ValidateJobs validate Jobs of kubernetes by ValidateK8SResource, put the results into the channel JobsResultsChan.
func ValidateJobs(ctx context.Context, jobs []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateReplicaJobs kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, job := range jobs {
		validateResults := ValidateK8SResource(ctx, job, regoRulesList, queryRule)
		validateReplicaJobs.ValidateResults = append(validateReplicaJobs.ValidateResults, validateResults)
	}
	resultChan <- validateReplicaJobs
}

// ValidateCronJobs validate cronjobs of kubernetes by ValidateK8SResource, put the results into the channel CronjobsResultsChan.
func ValidateCronJobs(ctx context.Context, cronjobs []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateCronjobsResults kube.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, cronjob := range cronjobs {
		validateResults := ValidateK8SResource(ctx, cronjob, regoRulesList, queryRule)
		validateCronjobsResults.ValidateResults = append(validateCronjobsResults.ValidateResults, validateResults)
	}
	resultChan <- validateCronjobsResults
}

// ValidateRoles validate roles of kubernetes by ValidateK8SResource, put the results into the channel RolesResultsChan.
func ValidateRoles(ctx context.Context, roles []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateRolesResults kube.ValidateResults
	queryRule := "data.kubeeye_RBAC_rego"
	for _, role := range roles {
		validateResults := ValidateK8SResource(ctx, role, regoRulesList, queryRule)
		validateRolesResults.ValidateResults = append(validateRolesResults.ValidateResults, validateResults)
	}
	resultChan <- validateRolesResults
}

// ValidateClusterRoles validate clusterroles of kubernetes by ValidateK8SResource, put the results into the channel ClusterRolesResultsChan.
func ValidateClusterRoles(ctx context.Context, clusterroles []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateClusterRolesResults kube.ValidateResults
	queryRule := "data.kubeeye_RBAC_rego"
	for _, clusterrole := range clusterroles {
		validateResults := ValidateK8SResource(ctx, clusterrole, regoRulesList, queryRule)
		validateClusterRolesResults.ValidateResults = append(validateClusterRolesResults.ValidateResults, validateResults)
	}
	resultChan <- validateClusterRolesResults
}

func ValidateNodes(ctx context.Context, nodes []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateNodesResults kube.ValidateResults
	queryRule := "data.kubeeye_nodes_rego"
	for _, node := range nodes {
		validateResults := ValidateK8SResource(ctx, node, regoRulesList, queryRule)
		validateNodesResults.ValidateResults = append(validateNodesResults.ValidateResults, validateResults)
	}
	resultChan <- validateNodesResults
}

func ValidateEvents(ctx context.Context, events []unstructured.Unstructured, regoRulesList kube.RegoRulesList, resultChan chan<- kube.ValidateResults) {
	var validateEventsResults kube.ValidateResults
	queryRule := "data.kubeeye_events_rego"
	for _, clusterrole := range events {
		validateResults := ValidateK8SResource(ctx, clusterrole, regoRulesList, queryRule)
		validateEventsResults.ValidateResults = append(validateEventsResults.ValidateResults, validateResults)
	}
	resultChan <- validateEventsResults
}

// ValidateK8SResource validate kubernetes resource by rego, return the validate results.
func ValidateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList kube.RegoRulesList, queryRule string) kube.ResultReceiver {
	var resultReceiver kube.ResultReceiver
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
	return resultReceiver
}
