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
	Data       = "data"
	Sysctl     = "sysctl"
	Systemd    = "systemd"
	FileFilter = "filefilter"
	Component  = "component"
)

const (
	LabelName             = "kubeeye.kubesphere.io/name"
	LabelResultName       = "kubeeye.kubesphere.io/result"
	LabelConfigType       = "kubeeye.kubesphere.io/config-type"
	LabelRuleGroup        = "kubeeye.kubesphere.io/rule-group"
	LabelInspectRuleGroup = "kubeeye.kubesphere.io/inspect-rule-group"
)

const (
	DefaultProcPath = "/hosts/proc"
	RootPathPrefix  = "/hosts/root"
)
