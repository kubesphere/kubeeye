package regorules

var memoryRequestsRule = Rule{
	Rule: `
		package memoryRequestsRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[_].resources.requests.memory)
		}
	`,
	Pkg:           "memoryRequestsRule",
	PromptMessage: "memory requests should be set.",
	Target:        "pod",
}
