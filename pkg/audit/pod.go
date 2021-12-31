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

type validateFunc func(ctx context.Context, regoRulesList []string) kube.ValidateResults

func RegoRulesValidate(ctx context.Context, queryRule string, Resources []unstructured.Unstructured) validateFunc {
	
	return func(ctx context.Context, regoRulesList []string) kube.ValidateResults {
		var validateRolesResults kube.ValidateResults
		for _, resource := range Resources {
			if validateResults, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
				validateRolesResults.ValidateResults = append(validateRolesResults.ValidateResults, validateResults)
			}
		}
		return validateRolesResults
	}
}

// ValidateResources Validate kubernetes cluster Resources, put the results into channels.
func MergeRegoRulesValidate(ctx context.Context, regoRulesChan <-chan string, vfuncs ...validateFunc) <-chan kube.ValidateResults {

	resultChan := make(chan kube.ValidateResults)
	var wg sync.WaitGroup
	wg.Add(len(vfuncs))

	regoRulesList := make([]string, 0)

	for rule := range regoRulesChan {
		regoRulesList = append(regoRulesList, rule)
	}

	mergeResult := func(ctx context.Context, vf validateFunc) {
		defer wg.Done()
		resultChan <- vf(ctx, regoRulesList)
	}
    for _, vf := range vfuncs {
		go mergeResult(ctx, vf)
	}

	go func() {
		defer close(resultChan)
		wg.Wait()
	}()
	return resultChan
}

// ValidateK8SResource validate kubernetes resource by rego, return the validate results.
func validateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList []string, queryRule string) (kube.ResultReceiver, bool) {
	var resultReceiver kube.ResultReceiver
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
