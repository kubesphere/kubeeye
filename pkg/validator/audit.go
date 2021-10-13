// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"context"
	"fmt"

	"github.com/kubesphere/kubeeye/regorules"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	certutil "k8s.io/client-go/util/cert"

	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kubesphere/kubeeye/pkg/kube"
)

func Cluster(ctx context.Context, kubeconfig string, additionalregoruleputh string, output string) error {

	// get kubernetes resources and put into the channel.
	go func(ctx context.Context, kubeconfig string) {
		kube.GetK8SResourcesProvider(ctx, kubeconfig)
	}(ctx, kubeconfig)

	// get rego rules and put into the channel.
	go func(additionalregoruleputh string) {
		regorules.GetRegoRules(additionalregoruleputh)
	}(additionalregoruleputh)

	defer close(kube.K8sResourcesChan)
	// get the kubernetes resources from the channel K8sResourcesChan.
	k8sResources := <-kube.K8sResourcesChan

	// todo
	// audit the events of cluster, it will be recode.
	clusterCheckResults, err2 := ProblemDetectorResult(k8sResources.ProblemDetector)
	if err2 != nil {
		return errors.Wrap(err2, "Failed to get clusterCheckResults information")
	}

	// todo
	// audit the nodes of cluster, it will be recode.
	nodeStatus, err3 := NodeStatusResult(k8sResources.Nodes.Items)
	if err3 != nil {
		return errors.Wrap(err3, "Failed to get nodeStatus information")
	}

	// todo
	// Get kube-apiserver certificate expiration, it will be recode.
	var certExpires []Certificate
	cmd := fmt.Sprintf("cat /etc/kubernetes/pki/%s", "apiserver.crt")
	combinedoutput, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if combinedoutput != nil {
		certs, _ := certutil.ParseCertsPEM([]byte(combinedoutput))
		if len(certs) != 0 {
			certExpire := Certificate{
				Name:     "kube-apiserver",
				Expires:  certs[0].NotAfter.Format("Jan 02, 2006 15:04 MST"),
				Residual: ResidualTime(certs[0].NotAfter),
			}
			if strings.Contains(certExpire.Residual, "d") {
				tmpTime, _ := strconv.Atoi(strings.TrimRight(certExpire.Residual, "d"))
				if tmpTime < 30 {
					certExpires = append(certExpires, certExpire)
				}
			} else {
				certExpires = append(certExpires, certExpire)
			}
		}
	}

	// ValidateResources Validate Kubernetes Resource, put the results into the channels.
	ValidateResources(ctx, k8sResources)

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		JSONOutput(clusterCheckResults, nodeStatus, certExpires)
	case "CSV", "csv", "Csv":
		CSVOutput(clusterCheckResults, nodeStatus, certExpires)
	default:
		defaultOutput(clusterCheckResults, nodeStatus, certExpires)
	}
	return nil
}

// ProblemDetectorResult Get kubernetes pod result, it will be recode.
func ProblemDetectorResult(event []v1.Event) ([]ClusterCheckResults, error) {
	var pdrs []ClusterCheckResults
	for j := 0; j < len(event); j++ {
		if event[j].Type == "Warning" {
			pdr := ClusterCheckResults{
				Namespace: event[j].ObjectMeta.Namespace,
				Name:      event[j].ObjectMeta.Name,
				EventTime: event[j].LastTimestamp.Time,
				Reason:    event[j].Reason,
				Message:   event[j].Message,
				Severity:  "Warning",
			}
			pdrs = append(pdrs, pdr)
		}
	}
	return pdrs, nil
}

//NodeStatusResult Get kubernetes node status result, it will be recode.
func NodeStatusResult(nodes []v1.Node) ([]AllNodeStatusResults, error) {
	var nodestatus []AllNodeStatusResults
	for k := 0; k < len(nodes); k++ {
		if nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Status == "True" {
			continue
		}
		nodestate := AllNodeStatusResults{
			Name:          nodes[k].ObjectMeta.Name,
			HeartbeatTime: nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].LastHeartbeatTime.Time,
			Status:        nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Status,
			Reason:        nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Reason,
			Message:       nodes[k].Status.Conditions[len(nodes[k].Status.Conditions)-1].Message,
			Severity:      "Fatal",
		}

		nodestatus = append(nodestatus, nodestate)
	}
	return nodestatus, nil
}

func ResidualTime(t time.Time) string {
	d := time.Until(t)
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds < 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}
