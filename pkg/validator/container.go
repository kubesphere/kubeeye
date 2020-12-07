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
	corev1 "k8s.io/api/core/v1"
	"kubeye/pkg/config"
	"kubeye/pkg/kube"
)

func ValidateContainer(ctx context.Context, conf *config.Configuration, controller kube.GenericWorkload, container *corev1.Container, isInit bool) (ContainerResult, error) {
	results, err := applyContainerSchemaChecks(ctx, conf, controller, container, isInit)
	if err != nil {
		return ContainerResult{}, err
	}

	cRes := ContainerResult{
		Results: results,
	}
	return cRes, nil
}
func ValidateAllContainers(ctx context.Context, conf *config.Configuration, controller kube.GenericWorkload) ([]ContainerResult, error) {
	results := []ContainerResult{}
	pod := controller.PodSpec
	//for _, container := range pod.InitContainers {
	//	result, err := ValidateContainer(ctx, conf, controller, &container, true)
	//	if err != nil {
	//		return nil, err
	//	}
	//	results = append(results, result)
	//}
	for _, container := range pod.Containers {
		result, err := ValidateContainer(ctx, conf, controller, &container, false)
		if err != nil {
			return nil, err
		}

		for key, deleteTrue := range result.Results {
			if true == deleteTrue.Success {
				delete(result.Results, key)
				continue
			}
		}
		results = append(results, result)
	}
	return results, nil
}
