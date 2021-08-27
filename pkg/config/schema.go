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

package config

import (
	"encoding/json"
	"github.com/qri-io/jsonschema"
	corev1 "k8s.io/api/core/v1"
)

type TargetKind string

const (
	// TargetContainer points to the container spec
	TargetContainer TargetKind = "Container"
	// TargetPod points to the pod spec
	TargetPod TargetKind = "Pod"
)

type SchemaCheck struct {
	ID       string `yaml:"id"`
	Category string `yaml:"category"`
	//SuccessMessage string                `yaml:"successMessage"`
	PromptMessage string                `yaml:"promptMessage"`
	Containers    includeExcludeList    `yaml:"containers"`
	Target        TargetKind            `yaml:"target"`
	SchemaTarget  TargetKind            `yaml:"schemaTarget"`
	Schema        jsonschema.RootSchema `yaml:"schema"`
	JSONSchema    string                `yaml:"jsonSchema"`
}

type includeExcludeList struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func (check SchemaCheck) CheckPod(pod *corev1.PodSpec) (bool, error) {
	return check.CheckObject(pod)
}
func (check SchemaCheck) CheckObject(obj interface{}) (bool, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return false, err
	}
	errs, err := check.Schema.ValidateBytes(bytes)
	return len(errs) == 0, err
}
func (check SchemaCheck) IsActionable(target TargetKind, controllerType string, isInit bool) bool {
	if check.Target != target {
		return false
	}

	if check.Target == TargetContainer {
		isIncluded := len(check.Containers.Include) == 0
		for _, inclusion := range check.Containers.Include {
			if (inclusion == "initContainer" && isInit) || (inclusion == "container" && !isInit) {
				isIncluded = true
				break
			}
		}
		if !isIncluded {
			return false
		}
		for _, exclusion := range check.Containers.Exclude {
			if (exclusion == "initContainer" && isInit) || (exclusion == "container" && !isInit) {
				return false
			}
		}
	}
	return true
}
func (check SchemaCheck) CheckContainer(container *corev1.Container) (bool, error) {
	return check.CheckObject(container)
}
func (check SchemaCheck) CheckController(bytes []byte) (bool, error) {
	errs, err := check.Schema.ValidateBytes(bytes)
	return len(errs) == 0, err
}
func (check *SchemaCheck) Initialize(id string) error {
	check.ID = id
	if check.JSONSchema != "" {
		if err := json.Unmarshal([]byte(check.JSONSchema), &check.Schema); err != nil {
			return err
		}
	}
	return nil
}
