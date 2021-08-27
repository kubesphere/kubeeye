package regorules

var priorityClassRule = Rule{
	Rule: `
		package priorityClassRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.priorityClassName)
		}
	`,
	Pkg:           "priorityClassRule",
	PromptMessage: "priorityClassName should be set.",
	Target:        "pod",
}
