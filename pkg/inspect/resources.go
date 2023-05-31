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

package inspect

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/kubesphere/kubeeye/visitor/parser"
	"github.com/open-policy-agent/opa/rego"
	"github.com/prometheus/client_golang/api"
	apiprometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex

type validateFunc func(ctx context.Context, regoRulesList []string) []v1alpha2.ResourceResult

func RegoRulesValidate(queryRule string, Resources kube.K8SResource, auditPercent *PercentOutput) validateFunc {

	return func(ctx context.Context, regoRulesList []string) []v1alpha2.ResourceResult {
		var auditResults []v1alpha2.ResourceResult

		if queryRule == workloads && Resources.Deployments != nil {
			for _, resource := range Resources.Deployments.Items {
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.StatefulSets != nil {
			for _, resource := range Resources.StatefulSets.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.DaemonSets != nil {
			for _, resource := range Resources.DaemonSets.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.Jobs != nil {
			for _, resource := range Resources.Jobs.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == workloads && Resources.CronJobs != nil {
			for _, resource := range Resources.CronJobs.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == rbac && Resources.Roles != nil {
			for _, resource := range Resources.Roles.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == rbac && Resources.ClusterRoles != nil {
			for _, resource := range Resources.ClusterRoles.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == nodes && Resources.Nodes != nil {
			for _, resource := range Resources.Nodes.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == events && Resources.Events != nil {
			for _, resource := range Resources.Events.Items {
				lock.Lock()
				auditPercent.CurrentAuditCount--
				auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
				lock.Unlock()
				if auditResult, found := validateK8SResource(ctx, resource, regoRulesList, queryRule); found {
					auditResults = append(auditResults, auditResult)
				}
			}
		}
		if queryRule == certexp && Resources.APIServerAddress != "" {
			lock.Lock()
			auditPercent.CurrentAuditCount--
			auditPercent.AuditPercent = (auditPercent.TotalAuditCount - auditPercent.CurrentAuditCount) * 100 / auditPercent.TotalAuditCount
			lock.Unlock()
			resource := Resources.APIServerAddress
			if auditResult, found := validateCertExp(resource); found {
				auditResults = append(auditResults, auditResult)

			}
		}

		return auditResults
	}
}

// MergeRegoRulesValidate Validate kubernetes cluster Resources, put the results into channels.
func MergeRegoRulesValidate(ctx context.Context, regoRulesChan []string, vfuncs ...validateFunc) <-chan []v1alpha2.ResourceResult {

	resultChan := make(chan []v1alpha2.ResourceResult)
	var wg sync.WaitGroup
	wg.Add(len(vfuncs))

	//regoRulesList := make([]string, 0)
	//
	//for rule := range regoRulesChan {
	//	regoRulesList = append(regoRulesList, rule)
	//}

	mergeResult := func(ctx context.Context, vf validateFunc) {
		defer wg.Done()
		resultChan <- vf(ctx, regoRulesChan)
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

func VailOpaRulesResult(ctx context.Context, k8sResources kube.K8SResource, RegoRules []string) v1alpha2.KubeeyeOpaResult {
	klog.Info("start Opa rule inspect")
	auditPercent := &PercentOutput{}

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

	//regoRulesChan := rules.MergeRegoRules(ctx, RegoRules, rules.GetAdditionalRegoRulesfiles(additionalregoruleputh))
	RulesValidateChan := MergeRegoRulesValidate(ctx, RegoRules,
		RegoRulesValidate(workloads, k8sResources, auditPercent),
		RegoRulesValidate(rbac, k8sResources, auditPercent),
		RegoRulesValidate(events, k8sResources, auditPercent),
		RegoRulesValidate(nodes, k8sResources, auditPercent),
		RegoRulesValidate(certexp, k8sResources, auditPercent),
	)
	klog.Info("get inspect results")

	RuleResult := v1alpha2.KubeeyeOpaResult{}
	var results []v1alpha2.ResourceResult
	ctxCancel, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				RuleResult.Percent = auditPercent.AuditPercent // update kubeeye inspect percent
				ext := runtime.RawExtension{}
				marshal, err := json.Marshal(RuleResult)
				if err != nil {
					klog.Error(err, " failed marshal kubeeye result")
					return
				}
				ext.Raw = marshal
				//auditResult.Result = ext
			case <-ctx.Done():
				return
			}
		}
	}(ctxCancel)

	for r := range RulesValidateChan {
		for _, result := range r {
			results = append(results, result)
		}
	}

	cancel()
	scoreInfo := CalculateScore(results, k8sResources)
	RuleResult.Percent = 100
	RuleResult.ScoreInfo = scoreInfo
	RuleResult.ExtraInfo = v1alpha2.ExtraInfo{
		WorkloadsCount: k8sResources.WorkloadsCount,
		NamespacesList: k8sResources.NameSpacesList,
	}

	RuleResult.ResourceResults = results
	return RuleResult
}

func PrometheusRulesResult(ctx context.Context, rule []byte) ([]byte, error) {
	var proRules []v1alpha2.PrometheusRule
	err := json.Unmarshal(rule, &proRules)
	if err != nil {
		klog.Error(err, " Failed to marshal kubeeye result")
		return nil, err
	}

	var proRuleResult [][]map[string]string
	for _, proRule := range proRules {
		client, err := api.NewClient(api.Config{
			Address: proRule.Endpoint,
		})
		if err != nil {
			klog.Error("create prometheus client failed", err)
			continue
		}
		queryApi := apiprometheusv1.NewAPI(client)
		query, _, err := queryApi.Query(ctx, *proRule.Rule, time.Now())
		if err != nil {
			klog.Errorf("failed to query rule:%s", *proRule.Rule)
			return nil, err
		}
		marshal, err := json.Marshal(query)

		var queryResults model.Samples
		err = json.Unmarshal(marshal, &queryResults)
		if err != nil {
			klog.Error("unmarshal modal Samples failed", err)
			continue
		}
		var queryResultsMap []map[string]string
		for i, result := range queryResults {
			temp := map[string]string{"value": result.Value.String(), "time": result.Timestamp.String()}
			klog.Info(i, result)
			for name, value := range result.Metric {
				klog.Info(name, value)
				temp[string(name)] = string(value)
			}
			queryResultsMap = append(queryResultsMap, temp)
		}

		proRuleResult = append(proRuleResult, queryResultsMap)
	}
	if proRules == nil && len(proRules) == 0 {
		return nil, nil
	}
	marshal, err := json.Marshal(proRuleResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func FileChangeRuleResult(ctx context.Context, task *v1alpha2.InspectTask, clients *kube.KubernetesClient, ownerRef ...v1.OwnerReference) ([]byte, error) {
	var nodeInfoResult v1alpha2.NodeInfoResult

	//fs, err := procfs.NewFS(constant.DefaultProcPath)
	//if err != nil {
	//	return nil, err
	//}
	//
	//meminfo, err := fs.Meminfo()
	//if err != nil {
	//	return nil, err
	//}
	//totalMemory := *meminfo.MemTotal
	//freeMemory := *meminfo.MemFree + *meminfo.Buffers + *meminfo.Cached
	//usedMemory := totalMemory - freeMemory
	//memoryUsage := float64(usedMemory) / float64(totalMemory)
	//memoryFree := float64(freeMemory) / float64(totalMemory)
	//nodeInfoResult.NodeInfo = map[string]string{"memoryUsage": fmt.Sprintf("%.2f", memoryUsage*100), "memoryIdle": fmt.Sprintf("%.2f", memoryFree*100)}
	//avg, err := fs.LoadAvg()
	//if err != nil {
	//	klog.Errorf(" failed to get loadavg,err:%s", err)
	//} else {
	//	nodeInfoResult.NodeInfo["load1"] = fmt.Sprintf("%.2f", avg.Load1)
	//	nodeInfoResult.NodeInfo["load5"] = fmt.Sprintf("%.2f", avg.Load5)
	//	nodeInfoResult.NodeInfo["load15"] = fmt.Sprintf("%.2f", avg.Load15)
	//}
	//
	//stat, err := fs.Stat()
	//if err != nil {
	//	klog.Error(err)
	//} else {
	//	totalUsage := 0.0
	//	totalIdle := 0.0
	//	for _, cpuStat := range stat.CPU {
	//		totalUsage += cpuStat.System + cpuStat.User + cpuStat.Nice
	//		totalIdle += cpuStat.Idle
	//	}
	//
	//	usage := totalUsage / (totalUsage + totalIdle)
	//	idle := totalIdle / (totalUsage + totalIdle)
	//	nodeInfoResult.NodeInfo["cpuUsage"] = fmt.Sprintf("%.2f", usage*100)
	//	nodeInfoResult.NodeInfo["cpuIdle"] = fmt.Sprintf("%.2f", idle*100)
	//}
	fileBytes, ok := task.Spec.Rules[constant.FileChange]
	if ok {
		var fileRule []v1alpha2.FileChangeRule
		err := json.Unmarshal(fileBytes, &fileRule)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		for _, file := range fileRule {
			var resultItem v1alpha2.FileChangeResultItem

			resultItem.FileName = file.Name
			resultItem.Path = file.Path
			baseFile, fileErr := os.ReadFile(path.Join(file.Path))
			if fileErr != nil {
				klog.Errorf("Failed to open base file path:%s,error:%s", baseFile, fileErr)
				resultItem.Issues = []string{fmt.Sprintf("%s:The file does not exist", file.Name)}
				nodeInfoResult.FileChangeResult = append(nodeInfoResult.FileChangeResult, resultItem)

				continue
			}
			baseFileName := fmt.Sprintf("%s-%s", constant.BaseFilePrefix, file.Name)
			baseConfig, configErr := clients.ClientSet.CoreV1().ConfigMaps(task.Namespace).Get(ctx, baseFileName, v1.GetOptions{})
			if configErr != nil {
				klog.Errorf("Failed to open file. cause：file Do not exist,err:%s", err)
				if kubeErr.IsNotFound(configErr) {
					var Immutable = true
					baseConfigMap := &corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Name:            baseFileName,
							Namespace:       task.Namespace,
							OwnerReferences: ownerRef,
							Labels:          map[string]string{constant.LabelConfigType: constant.BaseFile},
						},
						Immutable:  &Immutable,
						BinaryData: map[string][]byte{constant.FileChange: baseFile},
					}
					_, createErr := clients.ClientSet.CoreV1().ConfigMaps(task.Namespace).Create(ctx, baseConfigMap, v1.CreateOptions{})
					if createErr != nil {
						resultItem.Issues = []string{fmt.Sprintf("%s:create configMap failed", file.Name)}
						nodeInfoResult.FileChangeResult = append(nodeInfoResult.FileChangeResult, resultItem)
					}
					continue
				}
			}
			baseContent := baseConfig.BinaryData[constant.FileChange]
			//baseContent configmap读取的基准内容  baseFile文件读取需要对比的内容
			diffString := utils.DiffString(string(baseContent), string(baseFile))

			for i := range diffString {
				diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[32m", "")
				diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[31m", "")
				diffString[i] = strings.ReplaceAll(diffString[i], "\x1b[0m", "")
			}
			resultItem.Issues = diffString
			nodeInfoResult.FileChangeResult = append(nodeInfoResult.FileChangeResult, resultItem)
		}
	}

	//sysctlBytes, ok := task.Spec.Rules[constant.Sysctl]
	//if ok {
	//	var sysctl []v1alpha2.SysRule
	//	err := json.Unmarshal(sysctlBytes, &sysctl)
	//	if err != nil {
	//		klog.Error(err, " Failed to marshal kubeeye result")
	//		return nil, err
	//	}
	//
	//	for _, sysRule := range sysctl {
	//		ctlRule, err := fs.SysctlStrings(sysRule.Name)
	//		klog.Infof("name:%s,value:%s", sysRule.Name, ctlRule)
	//		var ctl v1alpha2.NodeResultItem
	//		ctl.Name = sysRule.Name
	//		if err != nil {
	//			errVal := fmt.Sprintf("name:%s to does not exist", sysRule.Name)
	//			ctl.Value = &errVal
	//		} else {
	//			val := strings.Join(ctlRule, ",")
	//			ctl.Value = &val
	//
	//			if sysRule.Rule != nil {
	//				if _, err := parser.CheckRule(*sysRule.Rule); err != nil {
	//					sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
	//					klog.Error(sprintf)
	//					ctl.Value = &sprintf
	//				} else {
	//					err, res := parser.EventRuleEvaluate(map[string]interface{}{sysRule.Name: ctlRule[0]}, *sysRule.Rule)
	//					if err != nil {
	//						sprintf := fmt.Sprintf("err:%s", err.Error())
	//						klog.Error(sprintf)
	//						ctl.Value = &sprintf
	//					} else {
	//						ctl.Assert = &res
	//					}
	//
	//				}
	//
	//			}
	//		}
	//		nodeInfoResult.SysctlResult = append(nodeInfoResult.SysctlResult, ctl)
	//	}
	//
	//}
	systemdBytes, ok := task.Spec.Rules[constant.Systemd]

	if ok {
		var systemd []v1alpha2.SysRule
		err := json.Unmarshal(systemdBytes, &systemd)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		conn, err := dbus.NewWithContext(ctx)
		if err != nil {
			return nil, err
		}
		unitsContext, err := conn.ListUnitsContext(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range systemd {
			var ctl v1alpha2.NodeResultItem
			ctl.Name = r.Name
			for _, status := range unitsContext {
				if status.Name == fmt.Sprintf("%s.service", r.Name) {
					ctl.Value = &status.ActiveState

					if r.Rule != nil {
						if _, err := parser.CheckRule(*r.Rule); err != nil {
							sprintf := fmt.Sprintf("rule condition is not correct, %s", err.Error())
							klog.Error(sprintf)
							ctl.Value = &sprintf
						} else {
							err, res := parser.EventRuleEvaluate(map[string]interface{}{r.Name: status.ActiveState}, *r.Rule)
							if err != nil {
								sprintf := fmt.Sprintf("err:%s", err.Error())
								klog.Error(sprintf)
								ctl.Value = &sprintf
							} else {
								ctl.Assert = &res
							}

						}

					}
					break
				}
			}
			if ctl.Value == nil {
				errVal := fmt.Sprintf("name:%s to does not exist", r.Name)
				ctl.Value = &errVal
			}
			nodeInfoResult.SystemdResult = append(nodeInfoResult.SystemdResult, ctl)
		}
	}

	marshal, err := json.Marshal(nodeInfoResult)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func OpaRuleResult(ctx context.Context, rule []byte, clients *kube.KubernetesClient) ([]byte, error) {
	k8sResources := kube.GetK8SResources(ctx, clients)
	klog.Info("getting  Rego rules")

	var opaRules []v1alpha2.OpaRule
	err := json.Unmarshal(rule, &opaRules)
	if err != nil {
		fmt.Printf("unmarshal opaRule failed,err:%s\n", err)
		return nil, err
	}
	var RegoRules []string
	for i := range opaRules {
		RegoRules = append(RegoRules, *opaRules[i].Rule)
	}

	result := VailOpaRulesResult(ctx, k8sResources, RegoRules)
	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

// ValidateK8SResource validate kubernetes resource by rego, return the validate results.
func validateK8SResource(ctx context.Context, resource unstructured.Unstructured, regoRulesList []string, queryRule string) (v1alpha2.ResourceResult, bool) {
	var auditResult v1alpha2.ResourceResult
	var resultItems v1alpha2.ResultItem
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
							auditResult.Name = result.Name
							auditResult.ResourceType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							auditResult.ResultItems = append(auditResult.ResultItems, resultItems)
						} else if result.Type == "Event" {
							auditResult.Name = result.Name
							auditResult.NameSpace = result.Namespace
							auditResult.ResourceType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							auditResult.ResultItems = append(auditResult.ResultItems, resultItems)
						} else {
							auditResult.Name = result.Name
							auditResult.NameSpace = result.Namespace
							auditResult.ResourceType = result.Type
							resultItems.Level = result.Level
							resultItems.Message = result.Message
							resultItems.Reason = result.Reason

							auditResult.ResultItems = append(auditResult.ResultItems, resultItems)
						}
					}
				}
			}
		}
	}
	return auditResult, find
}

// validateCertExp validate kube-apiserver certificate expiration
func validateCertExp(ApiAddress string) (v1alpha2.ResourceResult, bool) {
	var auditResult v1alpha2.ResourceResult
	var resultItems v1alpha2.ResultItem
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
			fmt.Printf("Failed to get Kubernetes kube-apiserver certificate expiration.\n")
			return auditResult, find
		}
		defer func() { _ = resp.Body.Close() }()

		for _, cert := range resp.TLS.PeerCertificates {
			expDate := int(cert.NotAfter.Sub(time.Now()).Hours() / 24)
			if expDate <= 30 {
				find = true
				auditResult.ResourceType = resourceType
				auditResult.Name = "certificateExpire"
				resultItems.Message = "CertificateExpiredPeriod"
				resultItems.Level = "dangerous"
				resultItems.Reason = "Certificate expiration time <= 30 days"
			}
		}
	}
	auditResult.ResultItems = append(auditResult.ResultItems, resultItems)
	return auditResult, find
}
