package main

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
)

func main() {
	var kc kube.KubernetesClient
	cluster, _ := kube.GetKubeConfigInCluster()
	clients, _ := kc.K8SClients(cluster)
	config := template.GetJobConfig(context.TODO(), clients)
	fmt.Println(config)
}
