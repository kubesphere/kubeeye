package regorules

var memoryLimitsRule = Rule{
	Rule: `
		package memoryLimitsRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[0].resources.limits.memory)
		}
	`,
	Pkg:           "memoryLimitsRule",
	PromptMessage: "memory limits should be set.",
	Target:        "pod",
}
