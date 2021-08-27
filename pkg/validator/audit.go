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
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	certutil "k8s.io/client-go/util/cert"
	conf "kubeeye/pkg/config"
	"kubeeye/pkg/kube"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func Cluster(configuration string, ctx context.Context, allInformation bool) error {
	k, err := kube.CreateResourceProvider(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get cluster information")
	}

	basicComponentStatus, err1 := ComponentStatusResult(k.ComponentStatus)
	if err1 != nil {
		return errors.Wrap(err1, "Failed to get BasicComponentStatus information")
	}

	clusterCheckResults, err2 := ProblemDetectorResult(k.ProblemDetector)
	if err2 != nil {
		return errors.Wrap(err2, "Failed to get clusterCheckResults information")
	}

	nodeStatus, err3 := NodeStatusResult(k.Nodes)
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

	var config conf.Configuration
	var goodPractice []PodResult
	if len(configuration) != 0 {
		fp, err := filepath.Abs(configuration)
		if err != nil {
			return errors.Wrap(err, "Failed to look up current directory")
		}
		config1, err := conf.ParseFile1(fp)
		goodPractice1, err := ValidatePods(ctx, &config1, k)
		goodPractice = append(goodPractice, goodPractice1...)

	}
	config, err = conf.ParseFile()
	goodPractice2, err := ValidatePods(ctx, &config, k)
	goodPractice = append(goodPractice, goodPractice2...)
	if err != nil {
		errors.Wrap(err, "Failed to get goodPractice information")
	}

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

	if len(basicComponentStatus) != 0 {
		fmt.Fprintln(w, "\nNAME\tSEVERITY\tTIME\tMESSAGE")
		for _, basiccomponentStatus := range basicComponentStatus {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				basiccomponentStatus.Name,
				basiccomponentStatus.Severity,
				basiccomponentStatus.Time,
				basiccomponentStatus.Message,
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
		fmt.Fprintln(w, "\nNAMESPACE\tSEVERITY\tNAME\tKIND\tTIME\tMESSAGE")
		for _, goodpractice := range goodPractice {
			var message []string
			if allInformation {
				for _, tmpMessage := range goodpractice.ContainerResults[0].Results {
					message = append(message, tmpMessage.Message, "")
				}
				if len(goodpractice.Results) != 0 {
					for _, tmpResult := range goodpractice.Results {
						if tmpResult.Success == false {
							message = append(message, tmpResult.Message, "")
						}
					}
					message = message[:len(message)-1]
				} else {
					message = message[:len(message)-1]
				}

			} else {
				message = goodpractice.Message
			}
			s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v",
				goodpractice.Namespace,
				goodpractice.Severity,
				goodpractice.Name,
				goodpractice.Kind,
				goodpractice.CreatedTime,
				message,
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

//Get kubernetes core component status result
func ComponentStatusResult(cs []v1.ComponentStatus) ([]BasicComponentStatus, error) {
	var crs []BasicComponentStatus
	for i := 0; i < len(cs); i++ {
		if strings.Contains(cs[i].Conditions[0].Message, "ok") == true || strings.Contains(cs[i].Conditions[0].Message, "true") == true {
			continue
		}

		cr := BasicComponentStatus{
			Time:     time.Now().Format(time.RFC3339),
			Name:     cs[i].ObjectMeta.Name,
			Message:  cs[i].Conditions[0].Message,
			Severity: "Fatal",
		}
		crs = append(crs, cr)
	}
	return crs, nil
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
