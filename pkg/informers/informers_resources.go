package informers

import (
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func KeEyeGver() map[schema.GroupVersion][]string {
	return map[schema.GroupVersion][]string{
		v1alpha2.SchemeGroupVersion: {"inspectrules", "inspectplans", "inspecttasks", "inspectresults"},
	}
}

func K8sEyeGver() map[schema.GroupVersion][]string {
	return map[schema.GroupVersion][]string{
		corev1.SchemeGroupVersion: {"configmaps", "nodes", "pods", "services"},
	}
}
