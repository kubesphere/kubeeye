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
	regorules2 "github.com/kubesphere/kubeeye/pkg/regorules"
)

var (
	workloads = "data.kubeeye_workloads_rego"
	rbac      = "data.kubeeye_RBAC_rego"
	nodes     = "data.kubeeye_nodes_rego"
	events    = "data.kubeeye_events_rego"
)

func Cluster(ctx context.Context, kubeconfig string, additionalregoruleputh string, output string) error {

	// get kubernetes resources and put into the channel.
	go func(ctx context.Context, kubeconfig string) {
		kube.GetK8SResourcesProvider(ctx, kubeconfig)
	}(ctx, kubeconfig)

	k8sResources := <-kube.K8sResourcesChan
	regoRulesChan := regorules2.MergeRegoRules(ctx, regorules2.GetDefaultRegofile("rules"), regorules2.GetRegoRulesfiles(additionalregoruleputh))
	// Get kube-apiserver certificate expiration, it will be recode.
	// var certExpires []kube.Certificate
	// cmd := fmt.Sprintf("cat /etc/kubernetes/pki/%s", "apiserver.crt")
	// combinedoutput, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	// if combinedoutput != nil {
	// 	certs, _ := certutil.ParseCertsPEM([]byte(combinedoutput))
	// 	if len(certs) != 0 {
	// 		certExpire := kube.Certificate{
	// 			Name:     "kube-apiserver",
	// 			Expires:  certs[0].NotAfter.Format("Jan 02, 2006 15:04 MST"),
	// 			Residual: ResidualTime(certs[0].NotAfter),
	// 		}
	// 		if strings.Contains(certExpire.Residual, "d") {
	// 			tmpTime, _ := strconv.Atoi(strings.TrimRight(certExpire.Residual, "d"))
	// 			if tmpTime < 30 {
	// 				certExpires = append(certExpires, certExpire)
	// 			}
	// 		} else {
	// 			certExpires = append(certExpires, certExpire)
	// 		}
	// 	}
	// }

	RegoRulesValidateChan := MergeRegoRulesValidate(ctx, regoRulesChan,
		RegoRulesValidate(ctx, workloads, k8sResources.Deployments),
		RegoRulesValidate(ctx, workloads, k8sResources.DaemonSets),
		RegoRulesValidate(ctx, workloads, k8sResources.Jobs),
		RegoRulesValidate(ctx, workloads, k8sResources.CronJobs),
		RegoRulesValidate(ctx, rbac, k8sResources.ClusterRoles),
		RegoRulesValidate(ctx, events, k8sResources.Events),
		RegoRulesValidate(ctx, nodes, k8sResources.Nodes),
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

// func ResidualTime(t time.Time) string {
// 	d := time.Until(t)
// 	if seconds := int(d.Seconds()); seconds < -1 {
// 		return fmt.Sprintf("<invalid>")
// 	} else if seconds < 0 {
// 		return fmt.Sprintf("0s")
// 	} else if seconds < 60 {
// 		return fmt.Sprintf("%ds", seconds)
// 	} else if minutes := int(d.Minutes()); minutes < 60 {
// 		return fmt.Sprintf("%dm", minutes)
// 	} else if hours := int(d.Hours()); hours < 24 {
// 		return fmt.Sprintf("%dh", hours)
// 	} else if hours < 24*365 {
// 		return fmt.Sprintf("%dd", hours/24)
// 	}
// 	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
// }
