/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package ctl

import (
	"github.com/golang/glog"
	"github.com/kubesphere/kubeeye/pkg/expend"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var uninstallNPD = &cobra.Command{
	Use:   "npd",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := expend.UninstallNPD(cmd.Context(), KubeConfig); err != nil {
			glog.Fatal("Uninstall npd failed with error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.AddCommand(uninstallNPD)

}
