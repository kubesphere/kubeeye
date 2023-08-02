package get

import (
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/kube"
	"github.com/spf13/cobra"
)

func NewGetCmd(client *kube.KubernetesClient) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get Inspect  the Config or Result for Cluster  ",
	}

	getCmd.AddCommand(NewResultCmd(client))
	return getCmd
}
