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
	"sync"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/regorules"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
	certexp   = "data.kubeeye_certexpiration"
)

func Cluster(ctx context.Context, kubeconfig string, additionalregoruleputh string, output string) error {

	// get kubernetes resources and put into the channel.
	go func(ctx context.Context, kubeconfig string) {
		err := kube.GetK8SResourcesProvider(ctx, kubeconfig)
		if err != nil {
			panic(err)
		}
	}(ctx, kubeconfig)

	k8sResources := <-kube.K8sResourcesChan
	regoRulesChan := regorules.MergeRegoRules(ctx, regorules.GetDefaultRegofile("rules"), regorules.GetAdditionalRegoRulesfiles(additionalregoruleputh))

	RegoRulesValidateChan := MergeRegoRulesValidate(ctx, regoRulesChan,
		RegoRulesValidate(workloads, k8sResources),
		RegoRulesValidate(rbac, k8sResources),
		RegoRulesValidate(events, k8sResources),
		RegoRulesValidate(nodes, k8sResources),
		RegoRulesValidate(certexp, k8sResources),
	)

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	validationResultsChan := MergeValidationResults(ctx, RegoRulesValidateChan)

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		JSONOutput(validationResultsChan)
	case "CSV", "csv", "Csv":
		CSVOutput(validationResultsChan)
	default:
		defaultOutput(validationResultsChan)
	}
	return nil
}

// MergeValidationResults merge all validate result from
func MergeValidationResults(ctx context.Context, channels ...<-chan kube.ValidateResults) <-chan kube.ValidateResults {
	result := make(chan kube.ValidateResults)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	mergeResult := func(ctx context.Context, ch <-chan kube.ValidateResults) {
		defer wg.Done()
		for c := range ch {
			result <- c
		}
	}

	for _, c := range channels {
		go mergeResult(ctx, c)
	}

	go func() {
		defer close(result)
		wg.Wait()
	}()

	return result
}
