package regorules

var readOnlyRootFilesystemRule = Rule{
	Rule: `
		package readOnlyRootFilesystemRule
		
		default allow = false
		
		allow {
			type_name(input.spec.securityContext.readOnlyRootFilesystem)
			input.spec.securityContext.readOnlyRootFilesystem == true
		} else {
			type_name(input.spec.containers[_].securityContext.readOnlyRootFilesystem)
			input.spec.containers[_].securityContext.readOnlyRootFilesystem == true
		}
	`,
	Pkg:           "readOnlyRootFilesystemRule",
	PromptMessage: "Root file system should be read only.",
	Target:        "pod",
}
