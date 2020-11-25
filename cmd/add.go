package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubeye/pkg/validator"
)

func init() {
	rootCmd.AddCommand(addCmd)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

var addCmd = &cobra.Command{
	Use:   "add ntp",
	Short: "add the ntp",
	Run: func(cmd *cobra.Command, args []string) {
		err := validator.Add(cmd.Context())
		if err != nil {
			fmt.Println(err)
		}
	},
}
