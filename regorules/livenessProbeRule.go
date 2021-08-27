package regorules

var livenessProbeRule = Rule{
	Rule: `
		package livenessProbeRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[_].livenessProbe)
		}
	`,
	Pkg:           "livenessProbeRule",
	PromptMessage: "livenessProbe should be set.",
	Target:        "pod",
}
