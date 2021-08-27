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

package validator

import (
	"context"
	"sync"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/regorules"
	"github.com/open-policy-agent/opa/rego"
	corev1 "k8s.io/api/core/v1"
)

func ValidatePods(ctx context.Context, kubeResources *kube.ResourceProvider, wg *sync.WaitGroup) {
	//controllers value includ kind(pod, daemonset, deployment), podSpec, ObjectMeta and OriginalObjectJSON
	workloadsToAudit := kubeResources.Workloads

	for _, workload := range workloadsToAudit {
		// set resource name, kind, namespace
		result := regorules.Result{
			Name:      workload.ObjectMeta.GetName(),
			Namespace: workload.ObjectMeta.GetNamespace(),
			Kind:      workload.Kind,
		}

		// go through the list of rules
		// work by goroutine, get OPA check results.
		wg.Add(1)
		go func(ctx context.Context, rules regorules.RulesList, pod corev1.Pod, result regorules.Result) {

			for _, rule := range rules.Rules {
				regoRule := rule.Rule                                                                                   //get the rule
				queryRule := "data." + rule.Pkg + ".allow"                                                              //set the query rule
				query, _ := rego.New(rego.Query(queryRule), rego.Module("examples.rego", regoRule)).PrepareForEval(ctx) //creat a rego and execute rego query
				results, err := query.Eval(ctx, rego.EvalInput(pod))                                                    // execute rego evaluation
				if err != nil {
					panic(err)
				}

				if results[0].Expressions[0].Value == true { // if the rego evaluation result is true, that means something should be set or not should be set, put result into the struck of Result.
					//rule.PromptMessage = rule.PromptMessage + "\t"
					result.PromptMessage = append(result.PromptMessage, rule.PromptMessage)
				}
			}
			resultChan <- result
		}(ctx, regorules.PodRulelist, workload.Pod, result)
	}
}
