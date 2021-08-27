package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kubeeye/pkg/validator"
)

var ntpImage string

var nsenterNtpCmd = &cobra.Command{
	Use:   "ntp",
	Short: "Check that the cluster NTP service is working",
	Run: func(cmd *cobra.Command, args []string) {
		err := validator.CheckNtp(cmd.Context(), ntpImage)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	nsenterCmd.AddCommand(nsenterNtpCmd)
	nsenterNtpCmd.Flags().StringVarP(&ntpImage, "image", "i", "kubespheredev/alpine:3.12", "Customize ntp image")

}
