package regorules

var imageTagRule = Rule{
	Rule: `
		package imageTagRule
		
		default allow = false
		
		allow {
			regex.match("^.+:latest$", input.spec.containers[_].image)
		}
	`,
	Pkg:           "imageTagRule",
	PromptMessage: "Do not ues \"latest\" in image tag, it should be specified.",
	Target:        "pod",
}
