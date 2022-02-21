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

	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/regorules"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
	certexp   = "data.kubeeye_certexpiration"
)

func Cluster(ctx context.Context, kubeConfigPath string, additionalregoruleputh string, output string) error {
	kubeConfig, err := kube.GetKubeConfig(kubeConfigPath)
	if err != nil {
		return errors.Wrap(err, "Failed to load config file")
	}

	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		return err
	}

	_, validationResultsChan := ValidationResults(ctx, clients, additionalregoruleputh)

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

func ValidationResults(ctx context.Context, kubernetesClient *kube.KubernetesClient, additionalregoruleputh string) (kube.K8SResource, <-chan v1alpha1.AuditResult) {
	log := log.FromContext(ctx)

	// get kubernetes resources and put into the channel.
	log.Info("starting get kubernetes resources")
	go func(ctx context.Context, kubernetesClient *kube.KubernetesClient) {
		err := kube.GetK8SResourcesProvider(ctx, kubernetesClient)
		if err != nil {
			log.Error(err, "failed to get kubernetes resources")
		}
	}(ctx, kubernetesClient)

	k8sResources := <-kube.K8sResourcesChan

	log.Info("getting and merging the Rego rules")
	regoRulesChan := regorules.MergeRegoRules(ctx, regorules.GetDefaultRegofile("rules"), regorules.GetAdditionalRegoRulesfiles(additionalregoruleputh))

	log.Info("starting audit kubernetes resources")
	RegoRulesValidateChan := MergeRegoRulesValidate(ctx, regoRulesChan,
		RegoRulesValidate(workloads, k8sResources),
		RegoRulesValidate(rbac, k8sResources),
		RegoRulesValidate(events, k8sResources),
		RegoRulesValidate(nodes, k8sResources),
		RegoRulesValidate(certexp, k8sResources),
	)

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	log.Info("return audit results")
	return MergeValidationResults(ctx, k8sResources, RegoRulesValidateChan)
}

// MergeValidationResults merge all validate result from
func MergeValidationResults(ctx context.Context, k8sResources kube.K8SResource, channels ...<-chan v1alpha1.AuditResult) (kube.K8SResource, <-chan v1alpha1.AuditResult) {
	result := make(chan v1alpha1.AuditResult)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	mergeResult := func(ctx context.Context, ch <-chan v1alpha1.AuditResult) {
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

	return k8sResources, result
}
