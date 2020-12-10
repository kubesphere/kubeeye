package kube

import (
	"context"
	"github.com/stretchr/testify/assert"
	"kubeye/test"
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
