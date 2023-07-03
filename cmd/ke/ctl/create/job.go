package create

import (
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/spf13/cobra"
)

var taskName string
var taskNamespace string
var resultName string

func NewJobCmd(client *kube.KubernetesClient) *cobra.Command {
	jobCmd := &cobra.Command{
		Use:   "job",
		Short: "create inspect job task",
	}

	jobCmd.AddCommand(NewFileChangeCmd(client))
	jobCmd.AddCommand(NewOpaCmd(client))
	jobCmd.AddCommand(NewPrometheusCmd(client))
	jobCmd.AddCommand(NewSysctlCmd(client))
	jobCmd.AddCommand(NewSystemdCmd(client))
	jobCmd.AddCommand(NewFileFilterCmd(client))

	jobCmd.PersistentFlags().StringVar(&resultName, "result-name", "", " result config name")
	jobCmd.PersistentFlags().StringVar(&taskName, "task-name", "", "task name")
	jobCmd.PersistentFlags().StringVar(&taskNamespace, "task-namespace", "", "task-namespace")

	return jobCmd
}
