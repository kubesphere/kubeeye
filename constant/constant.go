package constant

import "time"

const AuditorServiceAddrConfigMap = "auditor-service-addr"

const DefaultTimeout = 10 * time.Minute

const DefaultNamespace = "kubeeye-system"

const BaseFilePrefix = "kubeeye-base-file"
const (
	Opa        = "opa"
	FileChange = "filechange"
	Prometheus = "prometheus"
	BaseFile   = "basefile"
	Result     = "result"
	Sysctl     = "sysctl"
	Systemd    = "systemd"
)

const (
	LabelName       = "kubeeye.kubesphere.io/name"
	LabelResultName = "kubeeye.kubesphere.io/result"
	LabelConfigType = "kubeeye.kubesphere.io/configType"
	LabelRuleTag    = "kubeeye.kubesphere.io/rule-tag"
)
