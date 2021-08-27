package regorules

var imagePullPolicyRule = Rule{
	Rule: `
		package imagePullPolicyRule
		
		default allow = true
		
		allow = false {
			type_name(input.spec.containers[_].imagePullPolicy)
			input.spec.containers[_].imagePullPolicy == "Always"
		}
	`,
	Pkg:           "imagePullPolicyRule",
	PromptMessage: "imagePullPolicy should be \"Always\".",
	Target:        "pod",
}
