package regorules

var hostNetworkRule = Rule{
	Rule: `
		package hostNetworkRule
		
		default allow = false
		
		allow {
			type_name(input.spec.hostNetwork)
			input.spec.hostNetwork == true
		}
	`,
	Pkg:           "hostNetworkRule",
	PromptMessage: "hostNetwork should not be set.",
	Target:        "pod",
}
