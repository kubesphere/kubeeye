package opa

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	statsApi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func (rc *ResourceCollector) CollectNodeStatsSummary() ([]statsApi.Summary, error) {
	nodes, err := rc.client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nodeStatsSummarys []statsApi.Summary

	for _, node := range nodes.Items {
		if !isNodeReady(node) {
			continue
		}

		result := &statsApi.Summary{}
		// Get node stats summary (/api/v1/nodes/{nodeName}/proxy/stats/summary)
		resultRaw, err := rc.client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("stats/summary").Do(context.Background()).Raw()
		if err != nil {
			klog.Error(fmt.Sprintf("Failed get node %s stats summary", node.Name), err)
			continue
		}

		// Unmarshal node stats summary
		err = json.Unmarshal(resultRaw, result)
		if err != nil {
			klog.Error(fmt.Sprintf("Failed to unmarshal node %s stats summary", node.Name), err)
			continue
		}

		nodeStatsSummarys = append(nodeStatsSummarys, *result)
	}

	return nodeStatsSummarys, nil
}

// isNodeReady checks if a node is ready
func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
