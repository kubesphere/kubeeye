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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	funcrules "github.com/leonharetd/kubeeye/pkg/funcrules"
	"github.com/leonharetd/kubeeye/pkg/kube"
	register "github.com/leonharetd/kubeeye/pkg/register"
	_ "github.com/leonharetd/kubeeye/pkg/regorules"
	util "github.com/leonharetd/kubeeye/pkg/util"
)

type Interface interface {
	K8sResourcesProvider(ctx context.Context)
	Output(ctx context.Context)
}

func NewCluster(KubeConfig string) Cluster {
	return Cluster{
		k8sClient:        *kube.KubernetesAPI(KubeConfig),
		K8sResourcesChan: make(chan kube.K8SResource),
	}
}

type Cluster struct {
	k8sClient        kube.KubernetesClient
	K8sResourcesChan chan kube.K8SResource
}

func (c Cluster) K8sResourcesProvider(ctx context.Context) {
	// 全局队列，因为要共享
	err := kube.GetK8SResources(ctx, c.K8sResourcesChan, &c.k8sClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (c Cluster) getEmbedRegoRules(ctx context.Context) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, emb := range *register.RegoRuleList() {
			outOfTreeEmbFiles := util.ListRegoRuleFileName(emb)
			for _, rego := range util.GetRegoRules(outOfTreeEmbFiles, emb) {
				ch <- rego
			}
		}
	}()
	return ch
}

// additation file
func (c Cluster) getAddRegoRules(ctx context.Context, regorulePath string) <-chan string {
	ch := make(chan string)
	if regorulePath == "" {
		defer close(ch)
		return ch
	}
	go func() {
		defer close(ch)
		absPath, err := filepath.Abs(regorulePath)
		if err != nil {
			panic(err)
		}
		ADDLFS := os.DirFS(absPath)
		ADDLEmbedFiles := util.ListRegoRuleFileName(ADDLFS)
		for _, rego := range util.GetRegoRules(ADDLEmbedFiles, ADDLFS) {
			ch <- rego
		}
	}()
	return ch
}

func (c Cluster) getFuncRules(ctx context.Context) <-chan funcrules.FuncRule {
	ch := make(chan funcrules.FuncRule)
	go func() {
		defer close(ch)
		for _, f := range *register.FuncRuleList() {
			ch <- f
		}
	}()
	return ch
}

func (c Cluster) RegoRuleFanIn(ctx context.Context, channels ...<-chan string) <-chan string {
	res := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	mergeRegoRuls := func(ctx context.Context, ch <-chan string) {
		defer wg.Done()
		for c := range ch {
			res <- c
		}
	}

	for _, c := range channels {
		go mergeRegoRuls(ctx, c)
	}

	go func() {
		wg.Wait()
		defer close(res)
	}()
	return res
}

func (c Cluster) FuncRuleFanIn(ctx context.Context, channels ...<-chan funcrules.FuncRule) <-chan funcrules.FuncRule {
	res := make(chan funcrules.FuncRule)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	mergeRegoRuls := func(ctx context.Context, ch <-chan funcrules.FuncRule) {
		defer wg.Done()
		for c := range ch {
			res <- c
		}
	}

	for _, c := range channels {
		go mergeRegoRuls(ctx, c)
	}

	go func() {
		defer close(res)
		wg.Wait()
	}()
	return res
}

func (c Cluster) ValidateResultFanIn(ctx context.Context, channels ...<-chan funcrules.ValidateResults) <-chan funcrules.ValidateResults {
	fanIn := make(chan funcrules.ValidateResults)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	mergeResult := func(ctx context.Context, ch <-chan funcrules.ValidateResults) {
		defer wg.Done()
		for c := range ch {
			fanIn <- c
		}
	}

	for _, c := range channels {
		go mergeResult(ctx, c)
	}

	go func() {
		defer close(fanIn)
		wg.Wait()
	}()
	return fanIn
}

func (c Cluster) Run(ctx context.Context, regoruleputh string, output string) error {

	// get kubernetes resources and put into the channel.
	go c.K8sResourcesProvider(ctx)
	// get rego rules and put into the channel.
	regoRuleChan := c.RegoRuleFanIn(ctx, c.getEmbedRegoRules(ctx), c.getAddRegoRules(ctx, regoruleputh))
	funcRuleChan := c.FuncRuleFanIn(ctx, c.getFuncRules(ctx))
	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	validateResultChan := c.ValidateResultFanIn(ctx, ValidateRegoRules(ctx, c.K8sResourcesChan, regoRuleChan), ValidateFuncRules(ctx, funcRuleChan))

	switch output {
	case "JSON", "json", "Json":
		JsonOutput(validateResultChan)
	case "CSV", "csv", "Csv":
		CSVOutput(validateResultChan)
	default:
		defaultOutput(validateResultChan)
	}
	return nil
}
