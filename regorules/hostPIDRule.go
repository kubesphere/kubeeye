package regorules

var hostPIDRule = Rule{
	Rule: `
		package hostPIDRule
		
		default allow = false
		
		allow {
			type_name(input.spec.hostPID)
			input.spec.hostPID == true
		}
	`,
	Pkg:           "hostPIDRule",
	PromptMessage: "hostPID should not be set.",
	Target:        "pod",
}
