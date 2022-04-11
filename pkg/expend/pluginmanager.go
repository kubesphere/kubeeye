package expend

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/kubesphere/kubeeye/plugins/plugin-manage/pkg"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

var ClientSet *kubernetes.Clientset

// InstallOrUninstallPlugin, if isInstall is true, it will installed,otherwize uninstall
func InstallOrUninstallPlugin(ctx context.Context, namespace, pluginName string, isInstall bool) error {
	var installer Expends
	clients, err := GetK8SClients("")
	if err != nil {
		return err
	}
	installer = Installer{
		CTX:     ctx,
		Clients: clients,
	}
	ClientSet = clients.ClientSet.(*kubernetes.Clientset)

	resources, err := pkg.GetPluginManifest(ClientSet, namespace, pluginName)
	if err != nil {
		return err
	}

	pluginCRDResources := []byte(resources)
	pluginCRDResourceList := bytes.Split(pluginCRDResources, []byte("---"))

	for _, resource := range pluginCRDResourceList {
		if isInstall {
			err = installer.install(resource)
		} else {
			err = installer.uninstall(resource)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// IsPluginPodRunning check pod is running or not
// Preconditions: the define plugin name is the prefix of the container name
func IsPluginPodRunning(namespace, prefixPodName string) bool {
	var podName string
	var isRunning bool
	// wait pod creating
	time.Sleep(time.Second * time.Duration(pkg.IntervalsTime))
	pods, err := ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, prefixPodName) {
			podName = pod.Name
			switch pod.Status.Phase {
			case corev1.PodRunning:
				isRunning = true
			case corev1.PodPending:
			default:
				isRunning = false
			}
			break
		}
	}
	if !isRunning {
		return tickerGetPodStatus(namespace, podName)
	}
	return true
}

//tickerGetPodStatus Query the latest status of pods regularly
func tickerGetPodStatus(namespace, podName string) bool {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	isRunning := make(chan bool)
	var count int
	go func(isRunning chan bool) {
		for _ = range ticker.C {
			pod, err := ClientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
			if err != nil {
				if kubeErr.IsNotFound(err) {
					fmt.Printf("Pod %s in namespace %s not found\n", podName, namespace)
				}
				return
			}

			switch pod.Status.Phase {
			case corev1.PodRunning:
				isRunning <- true
			case corev1.PodPending:
			default:
				isRunning <- false
			}
			count++
			if count == pkg.MaxCheckPodCount {
				isRunning <- false
			}
		}
	}(isRunning)

	result := <-isRunning
	return result
}
