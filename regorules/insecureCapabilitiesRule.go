package regorules

var insecureCapabilitiesRule = Rule{
	Rule: `
		package insecureCapabilitiesRule
		capabilities_set = [{"key": [
			"CHOWN",
			"DAC_OVERRIDE",
			"FSETID",
			"FOWNER",
			"MKNOD",
			"NET_RAW",
			"SETGID",
			"SETUID",
			"SETFCAP",
			"SETPCAP",
			"NET_BIND_SERVICE",
			"SYS_CHROOT",
			"KILL",
			"AUDIT_WRITE"
		]}]

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
	Pkg:           "insecureCapabilitiesRule",
	PromptMessage: "insecure Capabilities Rule should not be set.",
	Target:        "pod",
}
