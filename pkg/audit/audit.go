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
	"k8s.io/klog/v2"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
	certexp   = "data.kubeeye_certexpiration"
)

var AuditPercent = sync.Map{}

type PercentOutput struct {
	TotalAuditCount   int
	CurrentAuditCount int
	AuditPercent      int
}

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

	_, validationResultsChan := ValidationResults(ctx, clients, additionalregoruleputh, "")

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		if err := JSONOutput(validationResultsChan); err != nil {
			return err
		}
	case "CSV", "csv", "Csv":
		if err := CSVOutput(validationResultsChan); err != nil {
			return err
		}
	default:
		if err := defaultOutput(validationResultsChan); err != nil {
			return err
		}
	}
	return nil
}

func ValidationResults(ctx context.Context, kubernetesClient *kube.KubernetesClient, additionalregoruleputh string, clusterInsigntName string) (kube.K8SResource, <-chan []v1alpha1.AuditResults) {
	// get kubernetes resources and put into the channel.
	klog.Info("starting get kubernetes resources")
	go func(ctx context.Context, kubernetesClient *kube.KubernetesClient) {
		err := kube.GetK8SResourcesProvider(ctx, kubernetesClient)
		if err != nil {
			klog.Error("failed to get kubernetes resources", err)
		}
	}(ctx, kubernetesClient)

	k8sResources := <-kube.K8sResourcesChan

	percent, ok := AuditPercent.Load(clusterInsigntName)
	var auditPercent *PercentOutput
	if !ok {
		auditPercent = &PercentOutput{}
		AuditPercent.Store(clusterInsigntName, auditPercent)
	} else {
		auditPercent = percent.(*PercentOutput)
	}

	if k8sResources.Deployments != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Deployments.Items)
	}
	if k8sResources.StatefulSets != nil {
		auditPercent.TotalAuditCount += len(k8sResources.StatefulSets.Items)
	}
	if k8sResources.DaemonSets != nil {
		auditPercent.TotalAuditCount += len(k8sResources.DaemonSets.Items)
	}
	if k8sResources.Jobs != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Jobs.Items)
	}
	if k8sResources.CronJobs != nil {
		auditPercent.TotalAuditCount += len(k8sResources.CronJobs.Items)
	}
	if k8sResources.Roles != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Roles.Items)
	}
	if k8sResources.ClusterRoles != nil {
		auditPercent.TotalAuditCount += len(k8sResources.ClusterRoles.Items)
	}
	if k8sResources.Nodes != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Nodes.Items)
	}
	if k8sResources.Events != nil {
		auditPercent.TotalAuditCount += len(k8sResources.Events.Items)
	}
	auditPercent.TotalAuditCount++
	auditPercent.CurrentAuditCount = auditPercent.TotalAuditCount

	klog.Info("getting and merging the Rego rules")
	regoRulesChan := regorules.MergeRegoRules(ctx, regorules.GetDefaultRegofile("rules"), regorules.GetAdditionalRegoRulesfiles(additionalregoruleputh))

	klog.Info("starting audit kubernetes resources")
	RegoRulesValidateChan := MergeRegoRulesValidate(ctx, regoRulesChan,
		RegoRulesValidate(workloads, k8sResources, auditPercent),
		RegoRulesValidate(rbac, k8sResources, auditPercent),
		RegoRulesValidate(events, k8sResources, auditPercent),
		RegoRulesValidate(nodes, k8sResources, auditPercent),
		RegoRulesValidate(certexp, k8sResources, auditPercent),
	)

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	klog.Info("get audit results")
	return k8sResources, RegoRulesValidateChan
}
