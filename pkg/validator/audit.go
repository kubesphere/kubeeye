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
	conf "kubeye/pkg/config"
	"kubeye/pkg/kube"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

func Cluster(configuration string, ctx context.Context) error {
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
		fmt.Fprintln(w, "\nNAMESPACE\tSEVERITY\tNODENAME\tEVENTTIME\tREASON\tMESSAGE")
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
			s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v",
				goodpractice.Namespace,
				goodpractice.Severity,
				goodpractice.Name,
				goodpractice.Kind,
				goodpractice.CreatedTime,
				goodpractice.Message,
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
