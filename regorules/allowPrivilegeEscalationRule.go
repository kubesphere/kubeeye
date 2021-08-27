package regorules

var allowPrivilegeEscalationRule = Rule{
	Rule: `
		package allowPrivilegeEscalationRule
		
		default allow = false
		
		allow {
			type_name(input.spec.securityContext.allowPrivilegeEscalation)
			input.spec.securityContext.allowPrivilegeEscalation == true
		} else {
			type_name(input.spec.containers[_].securityContext.allowPrivilegeEscalation)
			input.spec.containers[_].securityContext.allowPrivilegeEscalation == true
		}
	`,
	Pkg:           "allowPrivilegeEscalationRule",
	PromptMessage: "allowPrivilegeEscalation should be set false.",
	Target:        "pod",
}
