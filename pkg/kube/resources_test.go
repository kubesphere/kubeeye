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
package kube

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kubeeye/test"
	"testing"
)

func TestGetResourceFromAPI(t *testing.T) {
	k8s, dynamic := test.SetupTestAPI()
	k8s = test.SetupAddControllers(context.Background(), k8s, "test")
	resources, err := CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamic)
	assert.Equal(t, nil, err, "Error should be nil")
	assert.Equal(t, "test", resources.AuditAddress, "Should have source name")
	//assert.Equal(t,time.Now(),resources.CreationTime,"Creation time should be set")
	assert.Equal(t, 0, len(resources.Nodes), "Should be have any nodes")
	assert.Equal(t, 1, len(resources.Controllers), "Should have 1 controller")
	assert.Equal(t, 0, len(resources.ComponentStatus), "")
	assert.Equal(t, 0, len(resources.ProblemDetector), "")
	assert.Equal(t, "", resources.Controllers[0].ObjectMeta.GetName())
}
