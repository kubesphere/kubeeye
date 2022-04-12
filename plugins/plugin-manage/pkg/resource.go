package pkg

import (
	"context"
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPluginManifest(client *kubernetes.Clientset, namespace, pluginName string) (string, error) {
	configmap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), PluginConfig, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	manifest, ok := configmap.Data[fmt.Sprintf("%s%s", PrefixManifestKey, pluginName)]
	if !ok {
		return "", errors.New("Failed to get plugin configmap")
	}
	return manifest, nil
}
