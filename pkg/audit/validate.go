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

	"github.com/leonharetd/kubeeye/pkg/funcrules"
	"github.com/leonharetd/kubeeye/pkg/kube"
	"github.com/open-policy-agent/opa/rego"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func mergeValidateRegoRules(ctx context.Context, channels ...funcrules.ValidateResults) <-chan funcrules.ValidateResults {
	ch := make(chan funcrules.ValidateResults)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	validate := func(ctx context.Context, valid funcrules.ValidateResults) {
		defer wg.Done()
		ch <- valid
	}
	go func() {
		for _, c := range channels {
			go validate(ctx, c)
		}
	}()
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	return ch
}

type validateFunc func(ctx context.Context, workloads []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults

func validate(ctx context.Context, v validateFunc, Resources []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults {
	return v(ctx, Resources, regoRulesList)
}

// ValidateResources Validate kubernetes cluster Resources, put the results into channels.
func ValidateRegoRules(ctx context.Context, K8sResourcesChan chan kube.K8SResource, regoRulesChan <-chan string) <-chan funcrules.ValidateResults {
	// get the kubernetes resources from the channel K8sResourcesChan.
	k8sResources := <-K8sResourcesChan

	regoRulesList := make([]string, 0)
	for r := range regoRulesChan {
		regoRulesList = append(regoRulesList, string(r))
	}
	// validate workloads
	deployment := validate(ctx, workloads, k8sResources.Workloads.Deployments, regoRulesList)
	statefulSets := validate(ctx, workloads, k8sResources.Workloads.StatefulSets, regoRulesList)
	job := validate(ctx, workloads, k8sResources.Workloads.Jobs, regoRulesList)
	cronJobs := validate(ctx,  workloads, k8sResources.Workloads.CronJobs,regoRulesList)

	// validate roles
	roles := validate(ctx, RBAC, k8sResources.Roles, regoRulesList)
	clusterRoles := validate(ctx, RBAC, k8sResources.ClusterRoles, regoRulesList)
    // cluster
	nodes := validate(ctx, Nodes, k8sResources.Nodes,  regoRulesList)
	events := validate(ctx, Events,k8sResources.Events, regoRulesList)

	return mergeValidateRegoRules(ctx, deployment, statefulSets, job, cronJobs, roles, clusterRoles, nodes, events)
}

func ValidateFuncRules(ctx context.Context, funcRulesChan <-chan funcrules.FuncRule) <-chan funcrules.ValidateResults {
	ch := make(chan funcrules.ValidateResults)
	go func(ctx context.Context, funcs <-chan funcrules.FuncRule) {
		defer close(ch)
		for funcRule := range funcRulesChan {
			ch <- funcRule.Exec()
		}
	}(ctx, funcRulesChan)
	return ch
}

// ValidateDeployments validate deployments of kubernetes by ValidateK8SResource, put the results into the channel DeploymentsResultsChan.
func workloads(ctx context.Context, workloads []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults {
	var validateWorkloadsResults funcrules.ValidateResults
	queryRule := "data.kubeeye_workloads_rego"
	for _, w := range workloads {
		if validateResults, found := validateK8SResource(ctx, w, regoRulesList, queryRule); found {
			validateWorkloadsResults.ValidateResults = append(validateWorkloadsResults.ValidateResults, validateResults)
		}
	}
	return validateWorkloadsResults
}

// ValidateRoles validate roles of kubernetes by ValidateK8SResource, put the results into the channel RolesResultsChan.
func RBAC(ctx context.Context, roles []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults {
	var validateRolesResults funcrules.ValidateResults
	queryRule := "data.kubeeye_RBAC_rego"
	for _, role := range roles {
		if validateResults, found := validateK8SResource(ctx, role, regoRulesList, queryRule); found {
			validateRolesResults.ValidateResults = append(validateRolesResults.ValidateResults, validateResults)
		}
	}
	return validateRolesResults
}

func Nodes(ctx context.Context, nodes []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults {
	var validateNodesResults funcrules.ValidateResults
	queryRule := "data.kubeeye_nodes_rego"
	for _, node := range nodes {
		if validateResults, found := validateK8SResource(ctx, node, regoRulesList, queryRule); found {
			validateNodesResults.ValidateResults = append(validateNodesResults.ValidateResults, validateResults)
		}
	}
	return validateNodesResults
}

func Events(ctx context.Context, events []unstructured.Unstructured, regoRulesList []string) funcrules.ValidateResults {
	var validateEventsResults funcrules.ValidateResults
	queryRule := "data.kubeeye_events_rego"
	for _, clusterrole := range events {
		if validateResults, found := validateK8SResource(ctx, clusterrole, regoRulesList, queryRule); found {
			validateEventsResults.ValidateResults = append(validateEventsResults.ValidateResults, validateResults)
		}
	}
	return validateEventsResults
}

// ValidateK8SResource validate kubernetes resource by rego, return the validate results.
func validateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList []string, queryRule string) (funcrules.ResultReceiver, bool) {
	var resultReceiver funcrules.ResultReceiver
	find := false
	for _, regoRule := range regoRulesList {
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
