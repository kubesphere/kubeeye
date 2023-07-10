package main

import (
	"fmt"
	"math"
)

func main() {
	//cluster, _ := kube.GetKubeConfigInCluster()
	//kubeClient, _ := kubernetes.NewForConfig(cluster)
	//nodes := controllers.GetNodes(context.TODO(), kubeClient)

	fmt.Println(math.Round(float64(32) * 0.1))

}
