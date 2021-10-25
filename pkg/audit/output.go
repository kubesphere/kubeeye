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

package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
)

// defaultOutput get the results from channels, and then print out.
func defaultOutput(certExpires []kube.Certificate) {
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)

	fmt.Fprintln(w, "\nKIND\tNAMESPACE\tNAME\tMESSAGE")
	defer close(kube.DeploymentsResultsChan)
	deploymentsResults := <-kube.DeploymentsResultsChan
	for _, deploymentResult := range deploymentsResults.ValidateResults {
		if len(deploymentResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				deploymentResult.Type,
				deploymentResult.Namespace,
				deploymentResult.Name,
				deploymentResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.DaemonSetsResultsChan)
	daemonSetsResults := <-kube.DaemonSetsResultsChan
	for _, daemonSetResult := range daemonSetsResults.ValidateResults {
		if len(daemonSetResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				daemonSetResult.Type,
				daemonSetResult.Namespace,
				daemonSetResult.Name,
				daemonSetResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.StatefulSetsResultsChan)
	statefulSetsResults := <-kube.StatefulSetsResultsChan
	for _, statefulSetResult := range statefulSetsResults.ValidateResults {
		if len(statefulSetResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				statefulSetResult.Type,
				statefulSetResult.Namespace,
				statefulSetResult.Name,
				statefulSetResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.JobsResultsChan)
	jobsResults := <-kube.JobsResultsChan
	for _, jobResult := range jobsResults.ValidateResults {
		if len(jobResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				jobResult.Type,
				jobResult.Namespace,
				jobResult.Name,
				jobResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.CronjobsResultsChan)
	cronjobsResults := <-kube.CronjobsResultsChan
	for _, cronjobResult := range cronjobsResults.ValidateResults {
		if len(cronjobResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				cronjobResult.Type,
				cronjobResult.Namespace,
				cronjobResult.Name,
				cronjobResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.RolesResultsChan)
	rolesResults := <-kube.RolesResultsChan
	for _, roleResult := range rolesResults.ValidateResults {
		if len(roleResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				roleResult.Type,
				roleResult.Namespace,
				roleResult.Name,
				roleResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.ClusterRolesResultsChan)
	clusterRolesResults := <-kube.ClusterRolesResultsChan
	for _, clusterRoleResult := range clusterRolesResults.ValidateResults {
		if len(clusterRoleResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				clusterRoleResult.Type,
				"",
				clusterRoleResult.Name,
				clusterRoleResult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.NodesResultsChan)
	nodesResults := <-kube.NodesResultsChan
	for _, noderesult := range nodesResults.ValidateResults {
		if len(noderesult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				noderesult.Type,
				"",
				noderesult.Name,
				noderesult.Message,
			)
			fmt.Fprintln(w, s)
			continue
		}
	}

	defer close(kube.EventsResultsChan)
	eventsResults := <-kube.EventsResultsChan

	for _, eventsResult := range eventsResults.ValidateResults {
		if len(eventsResult.Message) != 0 {
			s := fmt.Sprintf("%s\t%s\t%s\t%-8v",
				eventsResult.Type,
				eventsResult.Namespace,
				eventsResult.Name,
				eventsResult.Message,
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
}

// JSONOutput get the results from channels, output by sjon.
func JSONOutput(certExpires []kube.Certificate) {
	var output []kube.ResultReceiver

	defer close(kube.DeploymentsResultsChan)
	deploymentsResults := <-kube.DeploymentsResultsChan
	for _, deploymentResult := range deploymentsResults.ValidateResults {
		if len(deploymentResult.Message) != 0 {
			output = append(output, deploymentResult)
		}
	}

	defer close(kube.DaemonSetsResultsChan)
	daemonSetsResults := <-kube.DaemonSetsResultsChan
	for _, daemonSetResult := range daemonSetsResults.ValidateResults {
		if len(daemonSetResult.Message) != 0 {
			output = append(output, daemonSetResult)
		}
	}

	defer close(kube.StatefulSetsResultsChan)
	statefulSetsResults := <-kube.StatefulSetsResultsChan
	for _, statefulSetResult := range statefulSetsResults.ValidateResults {
		if len(statefulSetResult.Message) != 0 {
			output = append(output, statefulSetResult)
		}
	}

	defer close(kube.JobsResultsChan)
	jobsResults := <-kube.JobsResultsChan
	for _, jobResult := range jobsResults.ValidateResults {
		if len(jobResult.Message) != 0 {
			output = append(output, jobResult)
		}
	}

	defer close(kube.CronjobsResultsChan)
	cronjobsResults := <-kube.CronjobsResultsChan
	for _, cronjobResult := range cronjobsResults.ValidateResults {
		if len(cronjobResult.Message) != 0 {
			output = append(output, cronjobResult)
		}
	}

	defer close(kube.RolesResultsChan)
	rolesResults := <-kube.RolesResultsChan
	for _, roleResult := range rolesResults.ValidateResults {
		if len(roleResult.Message) != 0 {
			output = append(output, roleResult)
		}
	}

	defer close(kube.ClusterRolesResultsChan)
	clusterRolesResults := <-kube.ClusterRolesResultsChan
	for _, clusterRoleResult := range clusterRolesResults.ValidateResults {
		if len(clusterRoleResult.Message) != 0 {
			output = append(output, clusterRoleResult)
		}
	}

	var certExpiresOutput kube.ResultReceiver
	if len(certExpires) != 0 {
		for _, certExpire := range certExpires {
			if len(certExpire.Expires) != 0 {
				certExpiresOutput.Name = certExpire.Name
				certExpiresOutput.Type = "certExpire"
				certExpiresOutput.Message = append(certExpiresOutput.Message, certExpire.Expires, certExpire.Residual)

				output = append(output, certExpiresOutput)
			}
		}
	}

	// output json
	jsonOutput, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(jsonOutput))
}

// CSVOutput get the results from channels, write to csv file.
func CSVOutput(certExpires []kube.Certificate) {
	var output []kube.ResultReceiver

	defer close(kube.DeploymentsResultsChan)
	deploymentsResults := <-kube.DeploymentsResultsChan
	for _, deploymentResult := range deploymentsResults.ValidateResults {
		if len(deploymentResult.Message) != 0 {
			output = append(output, deploymentResult)
		}
	}

	defer close(kube.DaemonSetsResultsChan)
	daemonSetsResults := <-kube.DaemonSetsResultsChan
	for _, daemonSetResult := range daemonSetsResults.ValidateResults {
		if len(daemonSetResult.Message) != 0 {
			output = append(output, daemonSetResult)
		}
	}

	defer close(kube.StatefulSetsResultsChan)
	statefulSetsResults := <-kube.StatefulSetsResultsChan
	for _, statefulSetResult := range statefulSetsResults.ValidateResults {
		if len(statefulSetResult.Message) != 0 {
			output = append(output, statefulSetResult)
		}
	}

	defer close(kube.JobsResultsChan)
	jobsResults := <-kube.JobsResultsChan
	for _, jobResult := range jobsResults.ValidateResults {
		if len(jobResult.Message) != 0 {
			output = append(output, jobResult)
		}
	}

	defer close(kube.CronjobsResultsChan)
	cronjobsResults := <-kube.CronjobsResultsChan
	for _, cronjobResult := range cronjobsResults.ValidateResults {
		if len(cronjobResult.Message) != 0 {
			output = append(output, cronjobResult)
		}
	}

	defer close(kube.RolesResultsChan)
	rolesResults := <-kube.RolesResultsChan
	for _, roleResult := range rolesResults.ValidateResults {
		if len(roleResult.Message) != 0 {
			output = append(output, roleResult)
		}
	}

	defer close(kube.ClusterRolesResultsChan)
	clusterRolesResults := <-kube.ClusterRolesResultsChan
	for _, clusterRoleResult := range clusterRolesResults.ValidateResults {
		if len(clusterRoleResult.Message) != 0 {
			output = append(output, clusterRoleResult)
		}
	}

	//var nodeStatusOutput kube.ResultReceiver
	//if len(nodeStatus) != 0 {
	//	for _, nodestatus := range nodeStatus {
	//		if len(nodestatus.Message) != 0 {
	//			nodeStatusOutput.Name = nodestatus.Name
	//			nodeStatusOutput.Type = "node"
	//			nodeStatusOutput.Message = append(nodeStatusOutput.Message, nodestatus.Message)
	//			nodeStatusOutput.Reason = nodestatus.Reason
	//
	//			output = append(output, nodeStatusOutput)
	//		}
	//	}
	//}
	//
	//var clusterCheckOutput kube.ResultReceiver
	//if len(clusterCheckResults) != 0 {
	//	for _, clusterCheckResult := range clusterCheckResults {
	//		if len(clusterCheckResult.Message) != 0 {
	//			clusterCheckOutput.Name = clusterCheckResult.Name
	//			clusterCheckOutput.Namespace = clusterCheckResult.Namespace
	//			clusterCheckOutput.Type = "cluster"
	//			clusterCheckOutput.Message = append(clusterCheckOutput.Message, clusterCheckResult.Message)
	//			clusterCheckOutput.Reason = clusterCheckResult.Reason
	//
	//			output = append(output, clusterCheckOutput)
	//		}
	//	}
	//}

	var certExpiresOutput kube.ResultReceiver
	if len(certExpires) != 0 {
		for _, certExpire := range certExpires {
			if len(certExpire.Expires) != 0 {
				certExpiresOutput.Name = certExpire.Name
				certExpiresOutput.Type = "certExpire"
				certExpiresOutput.Message = append(certExpiresOutput.Message, certExpire.Expires, certExpire.Residual)

				output = append(output, certExpiresOutput)
			}
		}
	}

	filename := "kubeEyeAuditResult.csv"
	// create csv file
	newFile, err := os.Create(filename)
	if err != nil {
		createError := errors.Wrap(err, "create file kubeEyeAuditResult.csv failed.")
		panic(createError)
	}

	defer newFile.Close()

	// write UTF-8 BOM to prevent print gibberish.
	newFile.WriteString("\xEF\xBB\xBF")

	// NewWriter returns a new Writer that writes to w.
	w := csv.NewWriter(newFile)
	header := []string{"name", "namespace", "kind", "message", "reason"}
	data := [][]string{
		header,
	}
	for _, receiver := range output {
		var testname string
		for _, msg := range receiver.Message {
			if testname == "" {
				contexts := []string{
					receiver.Name,
					receiver.Namespace,
					receiver.Type,
					msg,
					receiver.Reason,
				}
				data = append(data, contexts)
				testname = receiver.Name
			} else {
				contexts := []string{
					"",
					"",
					"",
					msg,
					receiver.Reason,
				}
				data = append(data, contexts)
			}

		}
	}

	// WriteAll writes multiple CSV records to w using Write and then calls Flush,
	if err := w.WriteAll(data); err != nil {
		fmt.Println("The result is exported to kubeeyeauditResult.CSV, please check it for audit result.")
	}
}
