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
	"github.com/leonharetd/kubeeye/pkg/kube"
	util "github.com/leonharetd/kubeeye/pkg/util"
	_ "github.com/leonharetd/kubeeye/pkg/execrules"
	_ "github.com/leonharetd/kubeeye/pkg/regorules"
	register "github.com/leonharetd/kubeeye/pkg/register"
)

func Cluster(ctx context.Context, kubeconfig string, additionalregoruleputh string, output string) error {

	// get kubernetes resources and put into the channel.
	go func(ctx context.Context, kubeconfig string) {
		kube.GetK8SResourcesProvider(ctx, kubeconfig)
	}(ctx, kubeconfig)

	// get rego rules and put into the channel.
	go func(additionalregoruleputh string) {
		// embed file
		for _, emb := range *register.RegoRuleList() {
			outOfTreeEmbFiles := util.ListRegoRuleFileName(emb)
			embedRegoRules := kube.RegoRulesList{RegoRules: util.GetRegoRules(outOfTreeEmbFiles, emb)}
			kube.RegoRulesListChan <- embedRegoRules
		}
		if additionalregoruleputh == "" {
			return
		}
		// additation embed file
		// addlFS := os.DirFS(additionalregoruleputh)
		// ADDLEmbedFiles := regorules2.ListRegoRuleFileName(addlFS)
		// ADDLEmbedRegoRules := kube.RegoRulesList{RegoRules: regorules2.GetRegoRules(ADDLEmbedFiles, addlFS)}
		// fmt.Println("addl", ADDLEmbedRegoRules)

	}(additionalregoruleputh)

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	go ValidateResources(ctx)
	// ValidateOther
	// go other(ctx)

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		JsonOutput(kube.ValidateResultsChan)
	case "CSV", "csv", "Csv":
		CSVOutput(kube.ValidateResultsChan)
	default:
		defaultOutput(kube.ValidateResultsChan)
	}
	return nil
}
