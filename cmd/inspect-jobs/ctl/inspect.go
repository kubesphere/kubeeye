package ctl

import (
	"fmt"
	"github.com/kubesphere/kubeeye/cmd/inspect-jobs/ctl/create"
	"github.com/kubesphere/kubeeye/cmd/inspect-jobs/ctl/get"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

var kubeConfig string

func Execute() {
	k8sConfig, err := kube.GetKubeConfig(kubeConfig)
	if err != nil {
		klog.Error(fmt.Sprintf("Failed to load cluster clients. err:%s", err))
		os.Exit(1)
	}
	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(k8sConfig)
	if err != nil {
		klog.Error(err, ",Failed to load cluster clients")
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "ke",
		Short: "inspect finds various problems on Kubernetes cluster.",
	}

	rootCmd.AddCommand(create.NewCmdCreate(clients))

	rootCmd.AddCommand(get.NewGetCmd(clients))

	addFlags(rootCmd)
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func addFlags(cmd *cobra.Command) {

}
