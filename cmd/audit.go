// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kubeye/pkg/validator"
)

var config string
var allInformation bool

var auditCmd = &cobra.Command{
	Use:   "diag",
	Short: "diagnostic information from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		err := validator.Cluster(config, cmd.Context(), allInformation)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	auditCmd.Flags().StringVarP(&config, "filename", "f", "", "Customize best practice configuration")
	auditCmd.Flags().BoolVarP(&allInformation, "all", "A", false, "Show more specific information")
}
