/*
Copyright © 2020 KubeSphere Authors

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
package cmd

import (
	"embed"

	"github.com/kubesphere/kubeeye/pkg/funcrules"
	register "github.com/kubesphere/kubeeye/pkg/register"
	"github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "ke",
	Short: "KubeEye finds various problems on Kubernetes cluster.",
}

func NewKubeEyeCommand() *KubeEyeCommand {
	return &KubeEyeCommand{
		commands: make([]*cobra.Command, 0),
	}
}

type KubeEyeCommand struct {
	commands []*cobra.Command
}

// add an additional command
func (ke *KubeEyeCommand) WithCommand(command *cobra.Command) *KubeEyeCommand {
	ke.commands = append(ke.commands, command)
	return ke
}

func (ke *KubeEyeCommand) WithRegoRule(r embed.FS) *KubeEyeCommand {
	register.RegoRuleRegistry(r)
	return ke
}

func (ke *KubeEyeCommand) WithFuncRule(e funcrules.FuncRule) *KubeEyeCommand {
	register.FuncRuleRegistry(e)
	return ke
}

// new an kubeeye command
func (ke KubeEyeCommand) DO() *cobra.Command {
	for _, command := range ke.commands {
		rootCmd.AddCommand(command)
	}
	return rootCmd
}
