package ctl

import (
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
)

var outPath string

var outFile = &cobra.Command{
	Use:   "outFile",
	Short: "out inspect result to excel",
	Run: func(cmd *cobra.Command, args []string) {

		if len(taskName) == 0 {
			list, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks("").List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				klog.Errorf("task not found: %s", err)
				os.Exit(1)
			}
			for _, item := range list.Items {
				fmt.Println(item.Name)
			}
			fmt.Println("all task loading done")
			os.Exit(1)
		}

		err := inspect.CSVOutput(clients, &outPath, taskName, taskNamespace)
		if err != nil {
			klog.Errorf("outfile error: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(outFile)
	outFile.Flags().StringVar(&outPath, "outpath", "", "result file out to path")
}
