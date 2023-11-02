package ctl

import (
	"github.com/kubesphere/kubeeye/cmd/ke/ctl/create"
	"github.com/spf13/cobra"
)

var kubeConfig string

func Execute() error {

	var rootCmd = &cobra.Command{
		Use:   "ke",
		Short: "inspect finds various problems on Kubernetes cluster.",
	}

	rootCmd.AddCommand(create.NewCmdCreate())

	addFlags(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func addFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&kubeConfig, "kube-config", "", "kube config")
}
