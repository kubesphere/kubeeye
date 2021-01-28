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
	"github.com/stretchr/testify/assert"
	conf "kubeeye/pkg/config"
	"kubeeye/pkg/kube"
	"kubeeye/test"
	"testing"
)

func TestInvalidIPCPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet": conf.SeverityWarning,
		},
	}

	k8s, _ := test.SetupTestAPI()
	k8s = test.SetupAddControllers(context.Background(), k8s, "test")
	p := test.MockPod()
	p.Spec.HostIPC = true
	workload, err := kube.NewGenericWorkloadFromPod(p, nil)
	assert.NoError(t, err)

	expectedResults := ResultSet{
		"hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC should not be configured", Success: false, Severity: "warning", Category: "Security"},
	}

	actualPodResult, err := ValidatePod(context.Background(), &c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, 1, len(actualPodResult.Results.GetWarnings()))
	assert.EqualValues(t, expectedResults, actualPodResult.Results)
}

func TestInvalidNeworkPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostNetworkSet": conf.SeverityWarning,
		},
	}

	k8s, _ := test.SetupTestAPI()
	k8s = test.SetupAddControllers(context.Background(), k8s, "test")
	p := test.MockPod()
	p.Spec.HostNetwork = true
	workload, err := kube.NewGenericWorkloadFromPod(p, nil)
	assert.NoError(t, err)

	expectedResults := ResultSet{
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network should not be configured", Success: false, Severity: "warning", Category: "Networking"},
	}

	actualPodResult, err := ValidatePod(context.Background(), &c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, 1, len(actualPodResult.Results.GetWarnings()))
	assert.EqualValues(t, expectedResults, actualPodResult.Results)
}

func TestInvalidPIDPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostPIDSet": conf.SeverityWarning,
		},
	}

	k8s, _ := test.SetupTestAPI()
	k8s = test.SetupAddControllers(context.Background(), k8s, "test")
	p := test.MockPod()
	p.Spec.HostPID = true
	workload, err := kube.NewGenericWorkloadFromPod(p, nil)
	assert.NoError(t, err)

	expectedResults := ResultSet{
		"hostPIDSet": {ID: "hostPIDSet", Message: "Host PID should not be configured", Success: false, Severity: "warning", Category: "Security"},
	}

	actualPodResult, err := ValidatePod(context.Background(), &c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, 1, len(actualPodResult.Results.GetWarnings()))
	assert.EqualValues(t, expectedResults, actualPodResult.Results)
}
