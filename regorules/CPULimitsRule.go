package regorules

var CPULimitsRule = Rule{
	Rule: `
		package CPULimitsRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[0].resources.limits.cpu)
		}
	`,
	Pkg:           "CPULimitsRule",
	PromptMessage: "CPU limits should be set.",
	Target:        "pod",
}
