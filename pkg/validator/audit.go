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
	"sync"

	"github.com/kubesphere/kubeeye/regorules"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	certutil "k8s.io/client-go/util/cert"

	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kubesphere/kubeeye/pkg/kube"
)

var resultChan = make(chan regorules.Result)

func Cluster(ctx context.Context) error {
	resources, err := kube.CreateResourceProvider(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get cluster information")
	}

	clusterCheckResults, err2 := ProblemDetectorResult(resources.ProblemDetector)
	if err2 != nil {
		return errors.Wrap(err2, "Failed to get clusterCheckResults information")
	}

	nodeStatus, err3 := NodeStatusResult(resources.Nodes)
	if err3 != nil {
		return errors.Wrap(err3, "Failed to get nodeStatus information")
	}

	// Get kube-apiserver certificate expiration
	var certExpires []Certificate
	cmd := fmt.Sprintf("cat /etc/kubernetes/pki/%s", "apiserver.crt")
	output, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if output != nil {
		certs, _ := certutil.ParseCertsPEM([]byte(output))
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

	wg := &sync.WaitGroup{}
	defer close(resultChan)

	ValidatePods(ctx, resources, wg)

	var podResults regorules.ResultsList
	// get results by goroutine
	go func(resultChan chan regorules.Result, podResults *regorules.ResultsList) {
		for {
			select {
			case result := <-resultChan:
				podResults.Results = append(podResults.Results, result)
				wg.Done()
			}
		}

	}(resultChan, &podResults)
	wg.Wait()

	goodPractice := podResults.Results

	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	if len(nodeStatus) != 0 {
		fmt.Fprintln(w, "NODENAME\tSEVERITY\tHEARTBEATTIME\tREASON\tMESSAGE")
		for _, nodestatus := range nodeStatus {
			s := fmt.Sprintf("%s\t%s\t%s\t%s\t%-8v",
				nodestatus.Name,
				nodestatus.Severity,
				nodestatus.HeartbeatTime.Format(time.RFC3339),
				nodestatus.Reason,
				nodestatus.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	if len(clusterCheckResults) != 0 {
		fmt.Fprintln(w, "\nNAMESPACE\tSEVERITY\tPODNAME\tEVENTTIME\tREASON\tMESSAGE")
		for _, clusterCheckResult := range clusterCheckResults {
			s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v",
				clusterCheckResult.Namespace,
				clusterCheckResult.Severity,
				clusterCheckResult.Name,
				clusterCheckResult.EventTime.Format(time.RFC3339),
				clusterCheckResult.Reason,
				clusterCheckResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	if len(goodPractice) != 0 {
		fmt.Fprintln(w, "\nNAMESPACE\tNAME\tKIND\tMESSAGE")
		for _, goodpractice := range goodPractice {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				goodpractice.Namespace,
				goodpractice.Name,
				goodpractice.Kind,
				goodpractice.PromptMessage,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}
	if len(certExpires) != 0 {
		fmt.Fprintln(w, "\nNAME\tEXPIRES\tRESIDUAL")
		for _, certExpire := range certExpires {
			s := fmt.Sprintf("%s\t%s\t%-8v",
				certExpire.Name,
				certExpire.Expires,
				certExpire.Residual,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}
	w.Flush()

	//auditData := AuditData{
	//	AuditTime:       k.CreationTime.Format(time.RFC3339),
	//	AuditAddress:      k.AuditAddress,
	//BasicComponentStatus: basicComponentStatus,
	//BasicClusterInformation: BasicClusterInformation{
	//	K8sVersion:   k.ServerVersion,
	//	PodNum:       len(k.Pods),
	//	NodeNum:      len(k.Nodes),
	//	NamespaceNum: len(k.Namespaces),
	//},

	//ClusterConfigurationResults: goodPractice,
	//AllNodeStatusResults:        nodeStatus,
	//ClusterCheckResults:         clusterCheckResults,
	//}

	//jsonBytes, err := json.Marshal(auditData)
	//outputBytes, err := yaml.JSONToYAML(jsonBytes)
	//os.Stdout.Write(outputBytes)
	return nil

}

//Get kubernetes pod result
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

//Get kubernetes node status result
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
