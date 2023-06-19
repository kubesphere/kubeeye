package create

import (
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/spf13/cobra"
)

func NewCmdCreate(client *kube.KubernetesClient) *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "create inspect job on Kubernetes cluster.",
	}

	createCmd.AddCommand(NewJobCmd(client))
	createCmd.AddCommand(NewConfigCmd())
	return createCmd
}
