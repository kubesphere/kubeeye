package regorules

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"github.com/ghodss/yaml"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/constant"
	"io/ioutil"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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

func GetDefaultRegofile(path string) []map[string][]byte {
	var regoRules []map[string][]byte
	files, err := defaultRegoRules.ReadDir(path)
	if err != nil {
		fmt.Printf("Failed to get Default Rego rule files.\n")
	}
	for _, file := range files {
		rule, _ := defaultRegoRules.ReadFile(path + "/" + file.Name())
		regoRule := map[string][]byte{"name": []byte(file.Name()), "rule": rule}
		regoRules = append(regoRules, regoRule)
	}
	return regoRules
}

func RegoToRuleYaml(path string) {
	regofile := GetDefaultRegofile(path)
	var inspectRules []kubeeyev1alpha2.InspectRules

	for _, m := range regofile {
		var ruleItems []kubeeyev1alpha2.OpaRule
		var inspectRule kubeeyev1alpha2.InspectRules
		opaRule := kubeeyev1alpha2.OpaRule{}
		var space string
		opaRule.Name = strings.Replace(string(m["name"]), ".rego", "", -1)
		opaRule.Rule = string(m["rule"])
		scanner := bufio.NewScanner(bytes.NewReader(m["rule"]))
		if scanner.Scan() {
			space = strings.TrimSpace(strings.Replace(scanner.Text(), "package", "", -1))
		}
		opaRule.Module = space
		for i := range inspectRules {
			if space == inspectRules[i].Labels[constant.LabelRuleTag] {
				inspectRule = inspectRules[i]
				inspectRules = append(inspectRules[:i], inspectRules[i+1:]...)
				break
			}
		}

		ruleItems = append(ruleItems, opaRule)

		inspectRule.Labels = map[string]string{
			"app.kubernetes.io/name":       "inspectrules",
			"app.kubernetes.io/instance":   "inspectrules-sample",
			"app.kubernetes.io/part-of":    "kubeeye",
			"app.kubernetes.io/managed-by": "kustomize",
			"app.kubernetes.io/created-by": "kubeeye",
			constant.LabelRuleTag:          space,
		}
		if inspectRule.Spec.Opas != nil {
			ruleItems = append(ruleItems, *inspectRule.Spec.Opas...)
		}

		inspectRule.Spec.Opas = &ruleItems
		inspectRule.Name = fmt.Sprintf("%s-%s", "kubeeye-inspectrules", strconv.Itoa(int(time.Now().Unix())))
		inspectRule.Namespace = "kubeeye-system"
		inspectRule.APIVersion = "kubeeye.kubesphere.io/v1alpha2"
		inspectRule.Kind = "InspectRules"
		inspectRules = append(inspectRules, inspectRule)
	}

	for i := range inspectRules {

		data, err := yaml.Marshal(&inspectRules[i])
		if err != nil {
			panic(err)
		}
		filename := fmt.Sprintf("./rules/kubeeye_v1alpha2_inspectrules%d_%d.yaml", i, time.Now().Unix())
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("YAML file written successfully")
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
