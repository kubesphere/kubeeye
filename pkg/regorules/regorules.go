package regorules

import (
	"context"
	"embed"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed rules
var defaultRegoRules embed.FS

// GetRegoRules get rego rules , put it into the channel RegoRulesListChan.
// func GetRegoRules(additionalRegoRulePath string) {
// 	var regoRulesList kube.RegoRulesList
// 	if additionalRegoRulePath != "" {
// 		GetRegoRulesfiles(additionalRegoRulePath, &regoRulesList)
// 	}
// 	GetDefaultRegofile("rules", &regoRulesList)

// 	kube.RegoRulesListChan <- regoRulesList
// }

// GetAdditionalRegoRulesfiles get Additional rego rules , put it into pointer of RegoRulesList
func GetAdditionalRegoRulesfiles(path string) []string {
	var regoRules []string
	if path == "" {
		return nil
	}
	pathabs, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get the files of additional rego rule.\033[0m\n")
	}
	if strings.HasSuffix(pathabs, "/") == false {
		pathabs += "/"
	}
	files, err := ioutil.ReadDir(pathabs)
	if err != nil {
		fmt.Printf("\033[1;33;49mFailed to get the dir of additional rego rule files.\033[0m\n")
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".rego") == false {
			continue
		}

		getregoRule, err := ioutil.ReadFile(pathabs + file.Name())
		if err != nil {
			fmt.Printf("\033[1;33;49mFailed to read the files of additional rego rules.\033[0m\n")
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
		fmt.Printf("\033[1;33;49mFailed to get Default Rego rule files.\033[0m\n")
	}
	for _, file := range files {
		rule, _ := defaultRegoRules.ReadFile(path + "/" + file.Name())
		regoRule := string(rule)
		regoRules = append(regoRules, regoRule)
	}
	return regoRules
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
