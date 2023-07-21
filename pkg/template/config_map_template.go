package template

import (
	"github.com/kubesphere/kubeeye/constant"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BinaryFileConfigMapTemplate(name string, namespace string, binaryData []byte, onRely bool, reference ...metav1.OwnerReference) *corev1.ConfigMap {
	return BinaryConfigMapTemplate(name, namespace, binaryData, onRely, map[string]string{constant.LabelConfigType: constant.BaseFile}, reference...)
}

func BinaryConfigMapTemplate(name string, namespace string, binaryData []byte, onRely bool, labels map[string]string, reference ...metav1.OwnerReference) *corev1.ConfigMap {
	resultConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: reference,
			Labels:          labels,
		},
		Immutable:  &onRely,
		BinaryData: map[string][]byte{constant.Data: binaryData},
	}

	return resultConfigMap
}
