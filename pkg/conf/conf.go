package conf

import (
	"github.com/aquasecurity/kube-bench/check"
)

const (
	AppsGroup = "apps"
	NoGroup = ""
	BatchGroup = "batch"
	RoleGroup = "rbac.authorization.k8s.io"
	APIVersionV1 = "v1"
	Nodes = "nodes"
	Deployments = "deployments"
	Daemonsets = "daemonsets"
	Statefulsets = "statefulsets"
	Jobs = "jobs"
	Cronjobs = "cronjobs"
	Namespaces = "namespaces"
	Events = "events"
	Roles = "roles"
	Clusterroles = "clusterroles"
)

type PluginsResult struct {
	// +kubebuilder:validation:one-of=[]reporthandling.FrameworkReport;kubehunterpkg.KubeHunterResults,kubebenchpkg.KubeBenchResults
	Name    string `json:"name,omitempty"`
	Results Results `json:"results,omitempty"`
	Ready   bool   `json:"ready,omitempty"`
}


type Results struct {
	KubeBenchResults KubeBenchResults `json:"kubebenchResults,omitempty"`
	KubeHunterResults KubeHunterResults `json:"kubehunterResults,omitempty"`
	KubescapeResults []FrameworkReport `json:"kubescapeResults,omitempty"`
	StringResults string `json:"stringResults,omitempty"`
}

type KubeBenchResults struct {
	Controls []check.Controls
}

// State is the state of a control check.
type State string
// Check contains information about a recommendation in the
// CIS Kubernetes document.
type Check struct {
	ID                string   `yaml:"id" json:"test_number"`
	Text              string   `json:"test_desc"`
	Audit             string   `json:"audit"`
	Type              string   `json:"type"`
	Set               bool     `json:"-"`
	Remediation       string   `json:"remediation"`
	TestInfo          []string `json:"test_info"`
	State             `json:"status"`
	ActualValue       string `json:"actual_value"`
	Scored            bool   `json:"scored"`
	ExpectedResult    string `json:"expected_result"`
	Reason            string `json:"reason,omitempty"`
	AuditOutput       string `json:"-"`
	AuditEnvOutput    string `json:"-"`
	AuditConfigOutput string `json:"-"`
	DisableEnvTesting bool   `json:"-"`
}

type KubeHunterResults struct {
	Nodes           []Node           `json:"nodes,omitempty"`
	Services        []Service        `json:"service,omitempty"`
	Vulnerabilities []Vulnerabilitie `json:"vulnerabilities,omitempty"`
}

type Node struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type Service struct {
	Service  string `json:"service"`
	Location string `json:"location"`
}

type Vulnerabilitie struct {
	Location      string `json:"location"`
	Vid           string `json:"vid"`
	Category      string `json:"category"`
	Severity      string `json:"severity"`
	Vulnerability string `json:"vulnerability"`
	Description   string `json:"description"`
	Evidence      string `json:"evidence"`
	Avd_reference string `json:"avd_reference"`
	Hunter        string `json:"hunter"`
}

type FrameworkReport struct {
	Name                  string          `json:"name"`
	ControlReports        []ControlReport `json:"controlReports"`
	Score                 string         `json:"score,omitempty"`
	ARMOImprovement       string         `json:"ARMOImprovement,omitempty"`
	WCSScore              string         `json:"wcsScore,omitempty"`
	ResourceUniqueCounter `json:",inline"`
}
type ResourceUniqueCounter struct {
	TotalResources   int `json:"totalResources"`
	FailedResources  int `json:"failedResources"`
	WarningResources int `json:"warningResources"`
}

type ControlReport struct {
	Control_ID            string       `json:"id,omitempty"` // to be Deprecated
	ControlID             string       `json:"controlID"`
	Name                  string       `json:"name"`
	RuleReports           []RuleReport `json:"ruleReports"`
	Remediation           string       `json:"remediation"`
	Description           string       `json:"description"`
	Score                 string      `json:"score"`
	BaseScore             string      `json:"baseScore,omitempty"`
	ARMOImprovement       string      `json:"ARMOImprovement,omitempty"`
	ResourceUniqueCounter `json:",inline"`
}

type RuleReport struct {
	Name                  string         `json:"name"`
	Remediation           string         `json:"remediation"`
	RuleStatus            RuleStatus     `json:"ruleStatus"` // did we run the rule or not (if there where compile errors, the value will be failed)
	RuleResponses         []RuleResponse `json:"ruleResponses"`
	ListInputKinds        []string       `json:"listInputIDs"`
	ResourceUniqueCounter `json:",inline"`
}
type RuleStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RuleResponse struct {
	FixCommand   string                            `json:"fixCommand,omitempty"`
	AlertMessage string                            `json:"alertMessage"`
	FailedPaths  []string                          `json:"failedPaths"`
	RuleStatus   string                            `json:"ruleStatus"`
	PackageName  string                            `json:"packagename"`
}