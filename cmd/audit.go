package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kubeye/pkg/validator"
)

var config string

func init() {
	rootCmd.AddCommand(auditCmd)
	//flag.Parse()
	//pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	auditCmd.Flags().StringVarP(&config, "filename", "f", "", "Customize best practice configuration")
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "audit the result",
	Run: func(cmd *cobra.Command, args []string) {
		err := validator.Cluster(config, cmd.Context())
		if err != nil {
			fmt.Println(err)
		}
	},
}
