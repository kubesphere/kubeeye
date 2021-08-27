package regorules

var hostPortRule = Rule{
	Rule: `
		package hostPortRule
		
		default allow = false
		
		allow {
			type_name(input.spec.containers[_].ports.hostPort)
		}
	`,
	Pkg:           "hostPortRule",
	PromptMessage: "hostPort should not be set.",
	Target:        "pod",
}
