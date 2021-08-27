package regorules

var CPURequestsRule = Rule{
	Rule: `
		package CPURequestsRule
		
		default allow = true
		
		allow = false{
			type_name(input.spec.containers[_].resources.requests.cpu)
		}
	`,
	Pkg:           "CPURequestsRule",
	PromptMessage: "CPU requests should be set.",
	Target:        "pod",
}
