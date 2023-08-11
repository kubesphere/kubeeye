package constant

import (
	"time"
)

const AuditorServiceAddrConfigMap = "auditor-service-addr"

const DefaultTimeout = 10 * time.Minute

const DefaultNamespace = "kubeeye-system"

const BaseFilePrefix = "kubeeye-base-file"
const (
	Opa           = "opa"
	FileChange    = "filechange"
	Prometheus    = "prometheus"
	BaseFile      = "basefile"
	Data          = "data"
	Sysctl        = "sysctl"
	Systemd       = "systemd"
	FileFilter    = "filefilter"
	Component     = "component"
	CustomCommand = "customcommand"
)

const (
	LabelName             = "kubeeye.kubesphere.io/name"
	LabelRuleType         = "kubeeye.kubesphere.io/rule-type"
	LabelTaskName         = "kubeeye.kubesphere.io/task-name"
	LabelNodeName         = "kubeeye.kubesphere.io/node-name"
	LabelConfigType       = "kubeeye.kubesphere.io/config-type"
	LabelRuleGroup        = "kubeeye.kubesphere.io/rule-group"
	LabelInspectRuleGroup = "kubeeye.kubesphere.io/inspect-rule-group"
)

const (
	AnnotationStartTime      = "kubeeye.kubesphere.io/task-start-time"
	AnnotationEndTime        = "kubeeye.kubesphere.io/task-end-time"
	AnnotationInspectPolicy  = "kubeeye.kubesphere.io/task-inspect-policy"
	AnnotationInspectCluster = "kubeeye.kubesphere.io/task-inspect-Cluster"
	AnnotationRuleJoinNum    = "kubeeye.kubesphere.io/rule-join-num"
)

const (
	DefaultProcPath = "/hosts/proc"
	RootPathPrefix  = "/hosts/root"
	//ResultPath      = "/hosts/result"
	ResultPath = "/Users/xiaoliu/Desktop/work/kubeeye"
)
