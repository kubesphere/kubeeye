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

func NewComponentCmd(client *kube.KubernetesClient) *cobra.Command {
	opaCmd := &cobra.Command{
		Use:   constant.Component,
		Short: "inspect on opa rule on Kubernetes cluster.",
		Run: func(cmd *cobra.Command, args []string) {

			if len(taskName) == 0 || len(resultName) == 0 {
				klog.Errorf("taskName or taskNamespace or resultName Incomplete parameters")
				os.Exit(1)
			}

			err := inspect.JobInspect(cmd.Context(), taskName, resultName, client, constant.Component)
			if err != nil {
				klog.Errorf("kubeeye inspect failed with error: %s,%v", err, err)
				os.Exit(1)
			}
			fmt.Println(args, taskName, "inspect success")
		},
	}
	return opaCmd
}
