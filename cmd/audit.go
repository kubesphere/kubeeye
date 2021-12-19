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
	"context"
	"flag"
	"fmt"

	"github.com/leonharetd/kubeeye/pkg/audit"
	cobra "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var KubeConfig string
var regorulepath string
var regoruleconfigmapName string
var output string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "audit resources from the cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if regoruleconfigmapName != "" && regorulepath != "" {
			fmt.Println("regorulepath or regoruleconfigmapName only one")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := audit.NewCluster(KubeConfig).Run(context.TODO(), regorulepath, output)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	rootCmd.PersistentFlags().StringVarP(&KubeConfig, "config", "f", "", "Specify the path of kubeconfig.")
	auditCmd.PersistentFlags().StringVarP(&regorulepath, "rego-rule-path", "p", "", "Specify the path of additional rego rule files director.")
	auditCmd.PersistentFlags().StringVarP(&regoruleconfigmapName, "rego-rule-configmap-name", "c", "", "Specify the name of additional rego rule configMap")
	auditCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "The format of result output, support JSON and CSV")
}
