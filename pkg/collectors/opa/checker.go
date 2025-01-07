package opa

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/open-policy-agent/opa/rego"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	statsApi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
	"sync"
)

type OPAChecker struct {
	batchSize int
	workers   int
}

func NewOPAChecker(batchSize int, workers int) *OPAChecker {
	return &OPAChecker{
		batchSize: batchSize,
		workers:   workers,
	}
}

func (oc *OPAChecker) VailOpaRulesResult(rulesManager *RulesManager, resourcesManager *ResourcesManager) (v1alpha2.KubeeyeOpaResult, error) {
	var result v1alpha2.KubeeyeOpaResult

	// create a channel to send tasks
	taskChan := make(chan task)
	resultChan := make(chan v1alpha2.ResourceResult)
	errChan := make(chan error)

	// launch workers
	var wg sync.WaitGroup
	for i := 0; i < oc.workers; i++ {
		klog.Infof("start worker %d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			oc.worker(taskChan, resultChan, errChan)
		}()
	}

	// collect results
	var results []v1alpha2.ResourceResult
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for r := range resultChan {
			results = append(results, r)
		}
	}()

	// send tasks
	go func() {
		for resourceType, resources := range resourcesManager.Resources {
			rules := rulesManager.Rules[resourceType]
			if rules == nil {
				continue
			}

			// handle resources in batch
			for i := 0; i < len(resources); i += oc.batchSize {
				end := i + oc.batchSize
				if end > len(resources) {
					end = len(resources)
				}

				taskChan <- task{
					resources:    resources[i:end],
					rules:        rules,
					resourceType: resourceType,
				}
			}
		}
		for resourceType, resources := range resourcesManager.StatsSummary {
			rules := rulesManager.Rules[resourceType]
			if rules == nil {
				continue
			}

			// handle resources in batch
			for i := 0; i < len(resources); i += oc.batchSize {
				end := i + oc.batchSize
				if end > len(resources) {
					end = len(resources)
				}

				taskChan <- task{
					statsSummary: resources[i:end],
					rules:        rules,
					resourceType: resourceType,
				}
			}
		}
		close(taskChan)
	}()

	// wait for all workers to finish
	wg.Wait()
	close(resultChan)

	// wait for result collector to finish
	resultWg.Wait()

	result.ResourceResults = results
	return result, nil
}

type task struct {
	resources    []unstructured.Unstructured
	statsSummary []statsApi.Summary
	rules        []*v1alpha2.OpaRule
	resourceType string
}

func (oc *OPAChecker) worker(tasks <-chan task, results chan<- v1alpha2.ResourceResult, errChan chan<- error) {
	klog.Infof("worker started, waiting for tasks, batchSize: %d, task count: %d", oc.batchSize, len(tasks))
	for task := range tasks {
		if task.resourceType == fmt.Sprintf("%s.v1", constant.NodeStatsSummary) {
			checkNodeStatsSummary(task, results)
		} else {
			checkResource(task, results)
		}
	}
}

// checkResource checks resource
func checkResource(task task, results chan<- v1alpha2.ResourceResult) {
	for _, resource := range task.resources {
		result := v1alpha2.ResourceResult{
			NameSpace:    resource.GetNamespace(),
			ResourceType: task.resourceType,
			Name:         resource.GetName(),
		}

		for _, rule := range task.rules {
			ctx := context.Background()

			klog.Infof("rule name: %v, resource name: %s", rule.Name, resource.GetName())
			query, err := rego.New(
				rego.Query("data.inspect.kubeeye.deny"),
				rego.Module("kubeeye.rego", rule.Rule),
			).PrepareForEval(ctx)

			if err != nil {
				klog.Errorf("prepare rule %s failed: %v", rule.Name, err)
				result.ResultItems = append(result.ResultItems, v1alpha2.ResultItem{
					Level:   string(rule.Level),
					Message: fmt.Sprintf("prepare rule %s failed: %v", rule.Name, err),
					Reason:  "PrepareRuleFailed",
				})
				continue
			}

			rs, err := query.Eval(ctx, rego.EvalInput(resource.Object))
			if err != nil {
				klog.Errorf("eval rule %s failed: %v", rule.Name, err)
				result.ResultItems = append(result.ResultItems, v1alpha2.ResultItem{
					Level:   string(rule.Level),
					Message: fmt.Sprintf("eval rule %s failed: %v", rule.Name, err),
					Reason:  "EvalRuleFailed",
				})
				continue
			}

			// parse result
			if len(rs) > 0 && len(rs[0].Expressions) > 0 {
				value := rs[0].Expressions[0].Value
				if violations, ok := value.([]interface{}); ok {
					for _, v := range violations {
						if violation, ok := v.(map[string]interface{}); ok {
							item := v1alpha2.ResultItem{
								Level:   fmt.Sprint(violation["Level"]),
								Message: fmt.Sprint(violation["Message"]),
								Reason:  fmt.Sprint(violation["Reason"]),
							}
							result.ResultItems = append(result.ResultItems, item)
						}
					}
				}
			}
		}

		if len(result.ResultItems) > 0 {
			results <- result
		}
	}
}

// checkNodeStatsSummary checks node stats summary
func checkNodeStatsSummary(task task, results chan<- v1alpha2.ResourceResult) {
	for _, resource := range task.statsSummary {
		result := v1alpha2.ResourceResult{
			NameSpace:    "",
			ResourceType: task.resourceType,
			Name:         resource.Node.NodeName,
		}

		for _, rule := range task.rules {
			ctx := context.Background()

			klog.Infof("rule name: %v, node stats summary for %s", rule.Name, resource.Node.NodeName)
			query, err := rego.New(
				rego.Query("data.inspect.kubeeye.nodeStatsSummary.deny"),
				rego.Module("nodeStatsSummary.rego", rule.Rule),
			).PrepareForEval(ctx)

			if err != nil {
				klog.Errorf("prepare rule %s failed: %v", rule.Name, err)
				result.ResultItems = append(result.ResultItems, v1alpha2.ResultItem{
					Level:   string(rule.Level),
					Message: fmt.Sprintf("prepare rule %s failed: %v", rule.Name, err),
					Reason:  "PrepareRuleFailed",
				})
				continue
			}

			rs, err := query.Eval(ctx, rego.EvalInput(resource))
			if err != nil {
				klog.Errorf("eval rule %s failed: %v", rule.Name, err)
				result.ResultItems = append(result.ResultItems, v1alpha2.ResultItem{
					Level:   string(rule.Level),
					Message: fmt.Sprintf("eval rule %s failed: %v", rule.Name, err),
					Reason:  "EvalRuleFailed",
				})
				continue
			}

			// parse result
			if len(rs) > 0 && len(rs[0].Expressions) > 0 {
				value := rs[0].Expressions[0].Value
				if violations, ok := value.([]interface{}); ok {
					for _, v := range violations {
						if violation, ok := v.(map[string]interface{}); ok {
							item := v1alpha2.ResultItem{
								Level:   fmt.Sprint(violation["Level"]),
								Message: fmt.Sprint(violation["Message"]),
								Reason:  fmt.Sprint(violation["Reason"]),
							}
							result.ResultItems = append(result.ResultItems, item)
						}
					}
				}
			}
		}

		if len(result.ResultItems) > 0 {
			results <- result
		}
	}
}
