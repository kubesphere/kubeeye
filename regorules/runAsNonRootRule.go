package regorules

var runAsNonRootRule = Rule{
	Rule: `
		package runAsNonRootRule
		
		default allow = false
		
		allow {
			type_name(input.spec.securityContext.runAsNonRoot)
			input.spec.securityContext.runAsNonRoot == true
		} else {
			type_name(input.spec.containers[_].securityContext.runAsNonRoot)
			input.spec.containers[_].securityContext.runAsNonRoot == true
		}
	`,
	Pkg:           "runAsNonRootRule",
	PromptMessage: "runAsNonRoot should be set.",
	Target:        "pod",
}
