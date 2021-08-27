package regorules

type Rule struct {
	Rule          string
	Pkg           string
	PromptMessage string
	Target        string
}

type RulesList struct {
	Rules []Rule
}

var PodRulelist = RulesList{
	[]Rule{
		allowPrivilegeEscalationRule,
		CPURequestsRule,
		CPULimitsRule,
		highRiskCapabilitiesRule,
		hostIPCRule,
		hostNetworkRule,
		hostPIDRule,
		hostPortRule,
		imagePullPolicyRule,
		imageTagRule,
		insecureCapabilitiesRule,
		livenessProbeRule,
		memoryRequestsRule,
		memoryLimitsRule,
		priorityClassRule,
		privilegedRule,
		readinessProbeRule,
		readOnlyRootFilesystemRule,
		runAsNonRootRule,
	},
}

type Result struct {
	Name          string
	Namespace     string
	Kind          string
	PromptMessage []string
}

type ResultsList struct {
	Results []Result
}
