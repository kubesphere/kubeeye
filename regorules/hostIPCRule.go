package regorules

var hostIPCRule = Rule{
	Rule: `
		package hostIPCRule
		
		default allow = false
		
		allow {
			type_name(input.spec.hostIPC)
			input.spec.hostIPC == true
		}
	`,
	Pkg:           "hostIPCRule",
	PromptMessage: "hostIPC should not be set.",
	Target:        "pod",
}
