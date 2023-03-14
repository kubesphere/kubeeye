package regorules

import (
	"context"
	"embed"
	"fmt"
	"github.com/kubesphere/kubeeye/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/constant"
	"io/ioutil"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed rules
var defaultRegoRules embed.FS

// GetAdditionalRegoRulesfiles get Additional rego rules , put it into pointer of RegoRulesList
func GetAdditionalRegoRulesfiles(path string) []string {
	var regoRules []string
	if path == "" {
		return nil
	}
	pathabs, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Failed to get the files of additional rego rule.\n")
	}
	if strings.HasSuffix(pathabs, "/") == false {
		pathabs += "/"
	}
	files, err := ioutil.ReadDir(pathabs)
	if err != nil {
		fmt.Printf("Failed to get the dir of additional rego rule files.\n")
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".rego") == false {
			continue
		}

		getregoRule, err := ioutil.ReadFile(pathabs + file.Name())
		if err != nil {
			fmt.Printf("Failed to read the files of additional rego rules.\n")
		}
		regoRule := string(getregoRule)
		regoRules = append(regoRules, regoRule)
	}
	return regoRules
}

func GetDefaultRegofile(path string) []string {
	var regoRules []string
	files, err := defaultRegoRules.ReadDir(path)
	if err != nil {
		fmt.Printf("Failed to get Default Rego rule files.\n")
	}
	for _, file := range files {
		rule, _ := defaultRegoRules.ReadFile(path + "/" + file.Name())
		regoRule := string(rule)
		regoRules = append(regoRules, regoRule)
	}
	return regoRules
}

func GetRegoRules(ctx context.Context, task types.NamespacedName, client versioned.Interface) []string {
	var rules []string

	inspectTask, err := client.KubeeyeV1alpha2().InspectTasks(task.Namespace).Get(ctx, task.Name, metav1.GetOptions{})
	if err != nil {
		if kubeErr.IsNotFound(err) {
			fmt.Printf("rego rules not found .\n")
			return nil
		}
		fmt.Printf("Failed to Get rego rules.\n")
		return nil
	}
	for _, rule := range inspectTask.Spec.Rules {
		if rule[constant.RuleType] == constant.Opa {
			rules = append(rules, rule[constant.Rules])
		}
	}
	return rules
}

// MergeRegoRules fun-out merge rego rules
func MergeRegoRules(ctx context.Context, channels ...[]string) <-chan string {
	res := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	mergeRegoRuls := func(ctx context.Context, ch []string) {
		defer wg.Done()
		for _, c := range ch {
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
