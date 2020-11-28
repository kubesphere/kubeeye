package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubeye/pkg/validator"
)

var config string

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

func init() {
	rootCmd.AddCommand(auditCmd)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	auditCmd.Flags().StringVarP(&config, "filename", "f", "", "Customize best practice configuration")
}
