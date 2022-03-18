package testdata

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var UnstructuredResource = &unstructured.Unstructured{
	Object: map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "testname",
			"namespace": "testns",
		},
	},
}