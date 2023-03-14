package constant

import "time"

const AuditorServiceAddrConfigMap = "auditor-service-addr"

const DefaultTimeout = 10 * time.Minute

const DefaultNamespace = "kubeeye-system"

const (
	Rules      = "rules"
	RuleType   = "ruleType"
	Opa        = "opa"
	Prometheus = "prometheus"
)
