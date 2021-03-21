package cmd

import "github.com/spf13/cobra"

var nsenterCmd = &cobra.Command{
	Use:   "nsenter",
	Short: "Debug host in the container",
}

func init() {
	rootCmd.AddCommand(nsenterCmd)
}
