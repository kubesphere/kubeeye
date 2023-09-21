package main

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
)

func main() {

	cluster, _ := kube.GetKubeConfigInCluster()

	var kc kube.KubernetesClient
	clients, _ := kc.K8SClients(cluster)

	opaInspect := inspect.OpaInspect{}
	runInspect, err := opaInspect.RunInspect(context.TODO(), nil, clients, "")
	if err != nil {

	}
	fmt.Println(runInspect)
}
