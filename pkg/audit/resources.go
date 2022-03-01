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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/open-policy-agent/opa/rego"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type validateFunc func(ctx context.Context, regoRulesList []string) kubeeyev1alpha1.AuditResult

func RegoRulesValidate(queryRule string, Resources kube.K8SResource) validateFunc {

	return func(ctx context.Context, regoRulesList []string) kubeeyev1alpha1.AuditResult {
		var auditResults kubeeyev1alpha1.AuditResult

		if queryRule == workloads && Resources.Deployments != nil {
			for _, resource := range Resources.Deployments.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.StatefulSets != nil {
			for _, resource := range Resources.StatefulSets.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.DaemonSets != nil {
			for _, resource := range Resources.DaemonSets.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.Jobs != nil {
			for _, resource := range Resources.Jobs.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.CronJobs != nil {
			for _, resource := range Resources.CronJobs.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == rbac && Resources.Roles != nil {
			for _, resource := range Resources.Roles.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == rbac && Resources.ClusterRoles != nil {
			for _, resource := range Resources.ClusterRoles.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == nodes && Resources.Nodes != nil {
			for _, resource := range Resources.Nodes.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == events && Resources.Events != nil {
			for _, resource := range Resources.Events.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults.Results = append(auditResults.Results, auditResult)
				}
			}
		}
		if queryRule == certexp && Resources.APIServerAddress != "" {
			resource := Resources.APIServerAddress
			if auditResult, found := validateCertExp(resource); found {
				auditResults.Results = append(auditResults.Results, auditResult)

			}
		}

		return auditResults
	}
}

// MergeRegoRulesValidate Validate kubernetes cluster Resources, put the results into channels.
func MergeRegoRulesValidate(ctx context.Context, regoRulesChan <-chan string, vfuncs ...validateFunc) <-chan kubeeyev1alpha1.AuditResult {

	resultChan := make(chan kubeeyev1alpha1.AuditResult)
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
func validateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList []string, queryRule string) (kubeeyev1alpha1.ValidateResults, bool) {
	var auditResult kubeeyev1alpha1.ValidateResults
	var resultReceiver kubeeyev1alpha1.ResultInfos
	var resourceInfos kubeeyev1alpha1.ResourceInfos
	var resultItems kubeeyev1alpha1.ResultItems
	find := false
	for _, regoRule := range regoRulesList {
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
			for key := range regoResult.Expressions {
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
						if result.Type == "ClusterRole" || result.Type == "Node" {
							resourceInfos.Name = result.Name
							auditResult.ResourcesType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
						} else if result.Type == "Event" {
							resourceInfos.Name = result.Name
							resultReceiver.Namespace = result.Namespace
							auditResult.ResourcesType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
						} else {
							resourceInfos.Name = result.Name
							resultReceiver.Namespace = result.Namespace
							auditResult.ResourcesType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
						}
					}
				}
			}
		}
	}
	resultReceiver.ResourceInfos = append(resultReceiver.ResourceInfos, resourceInfos)
	auditResult.ResultInfos = append(auditResult.ResultInfos, resultReceiver)
	return auditResult, find
}

// validateCertExp validate kube-apiserver certificate expiration
func validateCertExp(ApiAddress string) (kubeeyev1alpha1.ValidateResults, bool) {
	var auditResult kubeeyev1alpha1.ValidateResults
	var resultReceiver kubeeyev1alpha1.ResultInfos
	var resourceInfos kubeeyev1alpha1.ResourceInfos
	var resultItems kubeeyev1alpha1.ResultItems
	var find bool
	resourceType := "Cert"

	if ApiAddress != "" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(ApiAddress)
		if err != nil {
			find = false
			fmt.Printf("\033[1;33;49mFailed to get Kubernetes kube-apiserver certificate expiration.\033[0m\n")
			return auditResult, find
		}
		defer func() { _ = resp.Body.Close() }()

		for _, cert := range resp.TLS.PeerCertificates {
			expDate := int(cert.NotAfter.Sub(time.Now()).Hours() / 24)
			if expDate <= 90 {
				find = true
				auditResult.ResourcesType = resourceType
				resourceInfos.Name = "certificateExpire"
				resultItems.Message = "CertificateExpiredPeriod"
				resultItems.Reason = "Certificate expiration time <= 90 days"
			}
		}
	}

	resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
	resultReceiver.ResourceInfos = append(resultReceiver.ResourceInfos, resourceInfos)
	auditResult.ResultInfos = append(auditResult.ResultInfos, resultReceiver)
	return auditResult, find
}
