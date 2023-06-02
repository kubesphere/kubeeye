package ctl

import (
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

var sysctlCmd = &cobra.Command{
	Use:   constant.Sysctl,
	Short: "inspect finds various problems on Kubernetes cluster.",
	Run: func(cmd *cobra.Command, args []string) {

		if len(taskName) == 0 || len(taskNamespace) == 0 || len(resultName) == 0 {
			klog.Errorf("taskName or taskNamespace or resultName Incomplete parameters")
			os.Exit(1)
		}

		err := inspect.JobInspect(cmd.Context(), taskName, taskNamespace, resultName, clients, constant.Sysctl)
		if err != nil {
			klog.Errorf("kubeeye inspect failed with error: %s,%v", err, err)
			os.Exit(1)
		}
		fmt.Println(args, taskName, taskNamespace)
	},
}

func init() {
	rootCmd.AddCommand(sysctlCmd)
}
