package main

import (
	"context"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
)

func main() {
	var kc kube.KubernetesClient
	cluster, _ := kube.GetKubeConfigInCluster()

	clients, _ := kc.K8SClients(cluster)
	var task kubeeyev1alpha2.InspectResult
	err := clients.VersionClientSet.KubeeyeV1alpha2().RESTClient().Get().
		Resource("inspectresults").Name("inspectplan-20230705-10-00-filefilter").Do(context.TODO()).Into(&task)

	fmt.Println(err)
}
