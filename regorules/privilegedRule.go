package regorules

var privilegedRule = Rule{
	Rule: `
		package privilegedRule
		
		default allow = false
		
		allow {
			type_name(input.spec.securityContext.privileged)
			input.spec.securityContext.privileged == true
		} else {
			type_name(input.spec.containers[_].securityContext.privileged)
			input.spec.containers[_].securityContext.privileged == true
		}
	`,
	Pkg:           "privilegedRule",
	PromptMessage: "privileged should be set  false.",
	Target:        "pod",
}
