package create

import (
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

func NewFileFilterCmd(client *kube.KubernetesClient) *cobra.Command {
	systemdCmd := &cobra.Command{
		Use:   constant.FileFilter,
		Short: "inspect on file filter rule on Kubernetes cluster.",
		Run: func(cmd *cobra.Command, args []string) {

			if len(taskName) == 0 || len(taskNamespace) == 0 || len(resultName) == 0 {
				klog.Errorf("taskName or taskNamespace or resultName Incomplete parameters")
				os.Exit(1)
			}

			err := inspect.JobInspect(cmd.Context(), taskName, taskNamespace, resultName, client, constant.FileFilter)
			if err != nil {
				klog.Errorf("kubeeye inspect failed with error: %s,%v", err, err)
				os.Exit(1)
			}
			fmt.Println(args, taskName, taskNamespace)
		},
	}
	return systemdCmd
}
