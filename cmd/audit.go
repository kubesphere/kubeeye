package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubeye/pkg/validator"
)

func init() {
	rootCmd.AddCommand(auditCmd)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "audit the result",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := validator.Cluster(cmd.Context())
		if err != nil {
			fmt.Println(err)
		}
	},
}
