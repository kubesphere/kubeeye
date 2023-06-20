package main

import (
	"context"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {

	var kc kube.KubernetesClient
	cluster, _ := kube.GetKubeConfigInCluster()
	clients, _ := kc.K8SClients(cluster)
	//get, _ := clients.ClientSet.CoreV1().ConfigMaps("kubeeye-system").Get(context.TODO(), "inspectplan-1687247535-prometheus", metav1.GetOptions{})

	task, _ := clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks("kubeeye-system").Get(context.TODO(), "inspectplan-1687247535", metav1.GetOptions{})

	c, _ := client.New(cluster, client.Options{})

	inspectInterface := inspect.RuleOperatorMap[constant.Prometheus]

	runInspect, _ := inspectInterface.RunInspect(context.TODO(), task, clients, "inspectplan-1687247535-prometheus")

	configMap := &corev1.ConfigMap{BinaryData: map[string][]byte{constant.Result: runInspect}}

	inspectInterface.GetResult(context.TODO(), c, nil, configMap, task)

}
