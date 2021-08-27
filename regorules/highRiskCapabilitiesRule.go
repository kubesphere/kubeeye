package regorules

var highRiskCapabilitiesRule = Rule{
	Rule: `
		package highRiskCapabilitiesRule
		
		capabilities_set = [
			{"key": ["NET_ADMIN", "SYS_ADMIN", "ALL"]},
		]

		default allow = false
		
		allow {
			caps := input.spec.securityContext.capabilities
			type_name(caps)
			capabilities_set[_].key[_] == caps.add[_]
		} else {
			caps := input.spec.containers[_].securityContext.capabilities
			type_name(caps)
			capabilities_set[_].key[_] == caps.add[_]
		}
	`,
	Pkg:           "highRiskCapabilitiesRule",
	PromptMessage: "high-risk Capabilities Rule should not be set.",
	Target:        "pod",
}
