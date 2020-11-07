package validator

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	conf "kubeye/pkg/config"
	"kubeye/pkg/kube"

	"os"
	"sigs.k8s.io/yaml"
	"time"
)

func Cluster(ctx context.Context) (int, error) {
	k, err := kube.CreateResourceProvider(ctx)
	if err != nil {
		fmt.Println("do not get cluster information")
	}

	componentStatusResult, err := ComponentStatusResult(k.ComponentStatus)
	if err != nil {
		fmt.Println("do not get componentStatus")
	}

	problemDetector, err := ProblemDetectorResult(k.ProblemDetector)
	if err != nil {
		fmt.Println("do not get problemDetector")
	}

	nodeStatus, err := NodeStatusResult(k.Nodes)
	if err != nil {
		fmt.Println("do not get nodeStatus")
	}

	var config conf.Configuration
	config, err = conf.ParseFile()
	goodPractice, err := ValidatePods(ctx, &config, k)
	if err != nil {
		fmt.Println("1")
	}

	auditData := AuditData{
		AuditTime:       k.CreationTime.Format(time.RFC3339),
		AuditAddress:      k.AuditAddress,
		ComponentStatus: componentStatusResult,
		ClusterInfo: ClusterInfo{
			K8sVersion:   k.ServerVersion,
			PodNum:       len(k.Pods),
			NodeNum:      len(k.Nodes),
			NamespaceNum: len(k.Namespaces),
		},

		GoodPractice: goodPractice,
		NodeStatus: nodeStatus,
		ProblemDetector: problemDetector,
	}

	jsonBytes, err := json.Marshal(auditData)
	outputBytes, err := yaml.JSONToYAML(jsonBytes)
	return os.Stdout.Write(outputBytes)

}

func ComponentStatusResult(cs []v1.ComponentStatus) (interface{}, error) {
	cr := make(map[string]string)
	for i := 0; i < len(cs); i++ {
		cr[cs[i].ObjectMeta.Name] = cs[i].Conditions[0].Message
	}
	return cr, nil
}
func ProblemDetectorResult(event []v1.Event) ([]ProblemDetector, error) {
	var pdrs []ProblemDetector
	for j := 0; j < len(event); j++{
		if  event[j].Type == "Warning" {
			pdr := ProblemDetector{
				Namespace: event[j].ObjectMeta.Namespace,
				Name: event[j].ObjectMeta.Name,
				EventTime: event[j].LastTimestamp.Time,
				Reason: event[j].Reason,
				Message: event[j].Message,
			}
			pdrs = append(pdrs, pdr)
		}
	}
	return pdrs, nil
}
func NodeStatusResult(nodes []v1.Node) ([]NodeStatus, error) {
	var nodestatus []NodeStatus
	for k :=0; k < len(nodes); k++{
		nodestate := NodeStatus{
			Name: nodes[k].ObjectMeta.Name,
			HeartbeatTime: nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].LastHeartbeatTime.Time,
			Status: nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Status,
			Reason: nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Reason,
			Message: nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Message,
		}
		nodestatus = append(nodestatus, nodestate)
	}
	return nodestatus, nil
}