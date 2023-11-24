package constant

import (
	"time"
)

const AuditorServiceAddrConfigMap = "auditor-service-addr"

const DefaultTimeout = 10 * time.Minute

const (
	DefaultNamespace = "kubeeye-system"
)

var SystemNamespaces = []string{"kubesphere-system", "kubesphere-logging-system", "kubesphere-monitoring-system", "openpitrix-system", "kube-system", "istio-system", "kubesphere-devops-system", "porter-system"}

const BaseFilePrefix = "kubeeye-base-file"
const (
	Opa            = "opa"
	FileChange     = "filechange"
	Prometheus     = "prometheus"
	BaseFile       = "basefile"
	Data           = "data"
	Sysctl         = "sysctl"
	Systemd        = "systemd"
	FileFilter     = "filefilter"
	ServiceConnect = "serviceconnect"
	Component      = "component"
	CustomCommand  = "customcommand"
	NodeInfo       = "nodeinfo"
)

const (
	Cpu        = "cpu"
	Memory     = "memory"
	Filesystem = "filesystem"
	LoadAvg    = "loadavg"
	Inode      = "inode"
)

const (
	LabelName             = "kubeeye.kubesphere.io/name"
	LabelPlanName         = "kubeeye.kubesphere.io/plan-name"
	LabelRuleType         = "kubeeye.kubesphere.io/rule-type"
	LabelTaskName         = "kubeeye.kubesphere.io/task-name"
	LabelNodeName         = "kubeeye.kubesphere.io/node-name"
	LabelConfigType       = "kubeeye.kubesphere.io/config-type"
	LabelRuleGroup        = "kubeeye.kubesphere.io/rule-group"
	LabelInspectRuleGroup = "kubeeye.kubesphere.io/inspect-rule-group"
	LabelSystemWorkspace  = "kubesphere.io/workspace"
)

const (
	AnnotationStartTime     = "kubeeye.kubesphere.io/task-start-time"
	AnnotationEndTime       = "kubeeye.kubesphere.io/task-end-time"
	AnnotationInspectPolicy = "kubeeye.kubesphere.io/task-inspect-policy"
	AnnotationJoinPlanNum   = "kubeeye.kubesphere.io/join-plan-num"
	AnnotationJoinRuleNum   = "kubeeye.kubesphere.io/join-rule-num"
	AnnotationDescription   = "kubeeye.kubesphere.io/description"
	AnnotationInspectType   = "kubeeye.kubesphere.io/inspect-type"
)

const (
	DefaultProcPath = "/hosts/proc"
	RootPathPrefix  = "/hosts/root"
	ResultPath      = "/hosts/result"
)
