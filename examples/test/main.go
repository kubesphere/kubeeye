package main

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/open-policy-agent/opa/rego"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
)

func main() {
	var kc kube.KubernetesClient
	cluster, _ := kube.GetKubeConfigInCluster()

	clients, _ := kc.K8SClients(cluster)
	dynamicClient := clients.DynamicClient
	listOpts := metav1.ListOptions{}

	resourceGVR := schema.GroupVersionResource{Group: conf.NoGroup, Resource: "pods", Version: conf.APIVersionV1}
	rsource, err := dynamicClient.Resource(resourceGVR).List(context.TODO(), listOpts)
	if err != nil {
		fmt.Printf("Failed to get Kubernetes %v.\n,error:%s", rsource, err)
	}
	regoRule := `
package kubeeye_workloads_rego

deny[msg] {
	resource := input
	type := resource.Object.kind
	resourcename := resource.Object.metadata.name
	resourcenamespace := resource.Object.metadata.namespace
	type == "Pod"

	CheckPodPhase(resource)
	CheckPodReady(resource)
	msg := {
		"Name": sprintf("%v", [resourcename]),
		"Namespace": sprintf("%v", [resourcenamespace]),
		"Type": sprintf("%v", [type]),
		"Message": sprintf("Pod Status is :%v", [resource.Object.status.phase])
	}
}

CheckPodPhase(resource) {
    resource.Object.status.phase != "Running"; resource.Object.status.phase != "Succeeded"
}
CheckPodReady(resource) {
	resource.Object.status.containerStatuses[_].ready != true
}

`

	for _, item := range rsource.Items {
		query, err := rego.New(rego.Query("data.kubeeye_workloads_rego"), rego.Module("examples.rego", regoRule)).PrepareForEval(context.TODO())
		if err != nil {
			err := fmt.Errorf("failed to parse rego input: %s", err.Error())
			fmt.Println(err)
			os.Exit(1)
		}
		regoResults, err := query.Eval(context.TODO(), rego.EvalInput(item))
		if err != nil {
			err := fmt.Errorf("failed to validate resource: %s", err.Error())
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(regoResults)
	}
}
