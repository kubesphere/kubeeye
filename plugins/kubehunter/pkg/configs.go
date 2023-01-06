package pkg

type KubeHunterResults struct {
	Nodes           []Node           `json:"nodes"`
	Services        []Service        `json:"service"`
	Vulnerabilities []Vulnerabilitie `json:"vulnerabilities"`
}

type Node struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type Service struct {
	Service  string `json:"service"`
	Location string `json:"location"`
}

type Vulnerabilitie struct {
	Location      string `json:"location"`
	Vid           string `json:"vid"`
	Category      string `json:"category"`
	Severity      string `json:"severity"`
	Vulnerability string `json:"vulnerability"`
	Description   string `json:"description"`
	Evidence      string `json:"evidence"`
	Avd_reference string `json:"avd_reference"`
	Hunter        string `json:"hunter"`
}
