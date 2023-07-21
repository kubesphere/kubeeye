package main

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	cluster, _ := kube.GetKubeConfigInCluster()
	var kc kube.KubernetesClient
	clients, _ := kc.K8SClients(cluster)

	clusterName := "member1"
	c, _ := kube.GetMultiClusterClient(context.TODO(), clients, &clusterName)

	list, err := c.ClientSet.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(list)
	serviceAccount, err := clients.ClientSet.CoreV1().ServiceAccounts(constant.DefaultNamespace).Get(context.TODO(), "kubeeye-controller-manager", metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(serviceAccount)
	secret, err := clients.ClientSet.CoreV1().Secrets(constant.DefaultNamespace).Get(context.TODO(), serviceAccount.Secrets[0].Name, metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(secret)
}
