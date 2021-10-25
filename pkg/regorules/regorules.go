package regorules

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubesphere/kubeeye/pkg/kube"
)

//go:embed rules
var defaultRegoRules embed.FS

// GetRegoRules get rego rules , put it into the channel RegoRulesListChan.
func GetRegoRules(additionalRegoRulePath string) {
	var regoRulesList kube.RegoRulesList
	if additionalRegoRulePath != "" {
		GetRegoRulesfiles(additionalRegoRulePath, &regoRulesList)
	}
	GetDefaultRegofile("rules", &regoRulesList)

	kube.RegoRulesListChan <- regoRulesList
}

// GetRegoRulesfiles get rego rules , put it into pointer of RegoRulesList
func GetRegoRulesfiles(path string, regoRulesList *kube.RegoRulesList) {
	var regoRules []string
	pathabs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	if strings.HasSuffix(pathabs, "/") == false {
		pathabs += "/"
	}
	files, err := ioutil.ReadDir(pathabs)
	if err != nil {
		fmt.Println("Failed to read the dir of rego rule files.")
		os.Exit(1)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".rego") == false {
			continue
		}

		getregoRule, _ := ioutil.ReadFile(pathabs + file.Name())
		regoRule := string(getregoRule)
		regoRules = append(regoRules, regoRule)
	}
	regoRulesList.RegoRules = append(regoRulesList.RegoRules, regoRules...)
}

func GetDefaultRegofile(path string, regoRulesList *kube.RegoRulesList) {
	var regoRules []string
	files, err := defaultRegoRules.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		rule, _ := defaultRegoRules.ReadFile(path + "/" + file.Name())
		regoRule := string(rule)
		regoRules = append(regoRules, regoRule)
	}
	regoRulesList.RegoRules = append(regoRulesList.RegoRules, regoRules...)
}
