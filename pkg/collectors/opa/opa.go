package opa

import (
	"fmt"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"k8s.io/klog/v2"
	"regexp"
	"strings"
)

type RulesManager struct {
	Rules map[string][]*v1alpha2.OpaRule
}

func NewRulesManager() *RulesManager {
	return &RulesManager{
		Rules: make(map[string][]*v1alpha2.OpaRule),
	}
}

func (rm *RulesManager) AddRule(rule *v1alpha2.OpaRule) error {
	// parse resourceKind and apiVersion from rule
	resourceKind, apiVersion, err := parseResourceInfo(rule.Rule)
	if err != nil {
		return fmt.Errorf("failed to parse resource info from rule %s: %v", rule.Name, err)
	}

	key := fmt.Sprintf("%s.%s", resourceKind, apiVersion)

	rm.Rules[key] = append(rm.Rules[key], rule)

	klog.Infof("resourceKind: %s, apiVersion: %s, count: %d, ruleName: %s", resourceKind, apiVersion, len(rm.Rules[key]), rule.Name)

	return nil
}

// parseResourceInfo parses resourceKind and apiVersion from rego rule
func parseResourceInfo(regoContent string) (kind, apiVersion string, err error) {

	// check if regoContent contains "package inspect.kubeeye.nodeStatsSummary"
	if strings.Contains(regoContent, "package inspect.kubeeye.nodeStatsSummary") {
		return constant.NodeStatsSummary, "v1", nil
	}

	kindPattern := `input\.kind\s*==\s*"([^"]+)"`
	apiVersionPattern := `input\.apiVersion\s*==\s*"([^"]+)"`

	kindRegex := regexp.MustCompile(kindPattern)
	apiVersionRegex := regexp.MustCompile(apiVersionPattern)

	kindMatches := kindRegex.FindStringSubmatch(regoContent)
	apiVersionMatches := apiVersionRegex.FindStringSubmatch(regoContent)

	if len(kindMatches) < 2 || len(apiVersionMatches) < 2 {
		return "", "", fmt.Errorf("invalid rego rule format")
	}

	return kindMatches[1], apiVersionMatches[1], nil
}
