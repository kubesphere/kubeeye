package regorules

var readinessProbeRule = Rule{
	Rule: `
		package readinessProbeRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[_].readinessProbe)
		}
	`,
	Pkg:           "readinessProbeRule",
	PromptMessage: "readinessProbe should be set.",
	Target:        "pod",
}
