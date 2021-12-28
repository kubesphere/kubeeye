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
	"context"
	"fmt"

	regorules2 "github.com/kubesphere/kubeeye/pkg/regorules"
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
		regorules2.GetRegoRules(additionalregoruleputh)
	}(additionalregoruleputh)

	// todo
	// Get kube-apiserver certificate expiration, it will be recode.
	var certExpires []kube.Certificate
	cmd := fmt.Sprintf("cat /etc/kubernetes/pki/%s", "apiserver.crt")
	combinedoutput, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if combinedoutput != nil {
		certs, _ := certutil.ParseCertsPEM([]byte(combinedoutput))
		if len(certs) != 0 {
			certExpire := kube.Certificate{
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
	validateResultChan := ValidateResources(ctx)

	// Set the output mode, support default output JSON and CSV.
	switch output {
	case "JSON", "json", "Json":
		JSONOutput(validateResultChan)
	case "CSV", "csv", "Csv":
		CSVOutput(validateResultChan)
	default:
		defaultOutput(validateResultChan)
	}
	return nil
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
