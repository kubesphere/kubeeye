package validator

import (
	corev1 "k8s.io/api/core/v1"
	"kubeye/pkg/config"
	"time"
)

type AuditData struct {
	AuditTime        string
	AuditAddress     string
	ClusterInfo      ClusterInfo
	ComponentStatus  interface{}
	ProblemDetector  []ProblemDetector
	GoodPractice     []PodResult
	NodeStatus       []NodeStatus
}

type ProblemDetector struct {
	Namespace        string
	Name             string
	EventTime        time.Time
	Reason           string
	Message          string
}

type NodeStatus struct {
	Name             string
	Status           corev1.ConditionStatus
	HeartbeatTime    time.Time
	Reason           string
	Message          string
}

type ClusterInfo struct {
	K8sVersion       string
	NodeNum          int
	PodNum           int
	NamespaceNum     int
}

type PodResult struct {
	CreatedTime      string
	Namespace        string
	Kind             string
	Name             string
	ContainerResults []ContainerResult
}

type ContainerResult struct {
	Results          ResultSet
}

type ResultSet map[string]ResultMessage


type ResultMessage struct {
	ID               string
	Message          string
	Success          bool
	Severity         config.Severity
	Category         string
}
