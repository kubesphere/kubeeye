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
	"github.com/pkg/errors"
	"kubeye/pkg/config"
	"kubeye/pkg/kube"
	"time"
)

func ValidatePods(ctx context.Context, conf *config.Configuration, kubeResource *kube.ResourceProvider) ([]PodResult, error) {
	//controllers value includ kind(pod, daemonset, deployment), podSpec, ObjectMeta and OriginalObjectJSON
	podToAudit := kubeResource.Controllers

	results := []PodResult{}

	for _, pod := range podToAudit {
		result, err := ValidatePod(ctx, conf, pod)
		var messages []string
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get result")
		}

		if len(result.ContainerResults[0].Results) == 0 || result.ContainerResults == nil {
			continue
		}
		for key, _ := range result.ContainerResults[0].Results {
			messages = append(messages, key)
		}
		result.Message = messages
		result.Severity = "Warning"
		results = append(results, result)
	}
	return results, nil
}

func ValidatePod(ctx context.Context, c *config.Configuration, pod kube.GenericWorkload) (PodResult, error) {
	_, err := applyPodSchemaChecks(c, pod)
	if err != nil {
		return PodResult{}, err
	}
	pRes := PodResult{
		//Results:          podResults,
		ContainerResults: []ContainerResult{},
	}

	pRes.ContainerResults, err = ValidateAllContainers(ctx, c, pod)
	if err != nil {
		return pRes, err
	}

	result := PodResult{
		CreatedTime:      time.Now().Format(time.RFC3339),
		Kind:             pod.Kind,
		Name:             pod.ObjectMeta.GetName(),
		Namespace:        pod.ObjectMeta.GetNamespace(),
		ContainerResults: pRes.ContainerResults,
		Severity:         "Warning",
	}
	return result, nil

}
