package ctl

import (
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

var taskName string
var taskNamespace string
var resultName string
var KubeConfig string
var rootCmd = &cobra.Command{
	Use:   "ke",
	Short: "inspect finds various problems on Kubernetes cluster.",
}

var clients *kube.KubernetesClient

func Execute() {
	k8sConfig, err := kube.GetKubeConfig(KubeConfig)
	if err != nil {
		klog.Error(fmt.Sprintf("Failed to load cluster clients. err:%s", err))
		os.Exit(1)
	}
	var kc kube.KubernetesClient
	clients, err = kc.K8SClients(k8sConfig)
	if err != nil {
		klog.Error(err, ",Failed to load cluster clients")
		os.Exit(1)
	}

	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func init() {
	rootCmd.PersistentFlags().StringVar(&resultName, "result-name", "", "configmap name")
	rootCmd.PersistentFlags().StringVar(&taskName, "task-name", "", "task name")
	rootCmd.PersistentFlags().StringVar(&taskNamespace, "task-namespace", "", "task-namespace")
	rootCmd.PersistentFlags().StringVar(&KubeConfig, "kube-config", "", "kube-config")
}
