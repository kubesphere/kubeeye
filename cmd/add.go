package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubeye/pkg/validator"
)

var addCmd = &cobra.Command{
	Use:   "install npd",
	Short: "install the npd",
	Run: func(cmd *cobra.Command, args []string) {
		err := validator.Add(cmd.Context())
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}
