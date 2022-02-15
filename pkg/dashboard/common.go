package dashboard

type OverviewResponse struct {
	ApiVersion string `json:"apiVersion"`
	Kind string	`json:"kind"`
	Metadata Metadata `json:"metadata"`
	Spec OverviewSpec `json:"spec"`
}

type Metadata struct {
	Name string `json:"name"`
}

type OverviewSpec struct {
	ClusterVersion string `json:"clusterVersion"`
	NodesCount int `json:"nodesCount"`
	NamespacesCount int `json:"namespacesCount"`
	WorkloadsCount int `json:"workloadsCount"`
	NamespacesList []string `json:"namespacesList"`
}

const (
	overviewApi = "cluster/v1"
	overviewKind = "Overview"
	name = "clusterOverview"
)

