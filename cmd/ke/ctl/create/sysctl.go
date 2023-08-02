package create

import (
	"fmt"
	"github.com/kubesphere/kubeeye-v1alpha2/constant"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/inspect"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

func NewSysctlCmd(client *kube.KubernetesClient) *cobra.Command {
	sysctlCmd := &cobra.Command{
		Use:   constant.Sysctl,
		Short: "inspect on sysctl rule on Kubernetes cluster.",
		Run: func(cmd *cobra.Command, args []string) {

			if len(taskName) == 0 || len(resultName) == 0 {
				klog.Errorf("taskName  or resultName Incomplete parameters")
				os.Exit(1)
			}

			err := inspect.JobInspect(cmd.Context(), taskName, resultName, client, constant.Sysctl)
			if err != nil {
				klog.Errorf("kubeeye inspect failed with error: %s,%v", err, err)
				os.Exit(1)
			}
			fmt.Println(args, taskName, "inspect success")
		},
	}
	return sysctlCmd
}
