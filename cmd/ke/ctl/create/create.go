package create

import (
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "create inspect job on Kubernetes cluster.",
	}

	createCmd.AddCommand(NewJobCmd())
	createCmd.AddCommand(NewConfigCmd())
	return createCmd
}
