package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net"
	"strings"
	"time"
)

type componentInspect struct {
}

func init() {
	RuleOperatorMap[constant.Component] = &componentInspect{}
}

func (o *componentInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig) (*kubeeyev1alpha2.JobPhase, error) {

	job := template.InspectJobsTemplate(config, jobRule.JobName, task, "", nil, constant.Component)

	_, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", job.Name, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil

}

func (o *componentInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var components string
		err := json.Unmarshal(phase.RunRule, &components)
		if err != nil {
			return nil, err
		}
		component, err := GetInspectComponent(ctx, clients, components)
		if err != nil {
			return nil, err
		}
		var componentResult []kubeeyev1alpha2.ComponentResultItem
		for _, item := range component {
			endpoint := fmt.Sprintf("%s.%s.svc.cluster.local:%d", item.Name, item.Namespace, item.Spec.Ports[0].Port)
			isConnected := checkConnection(endpoint)
			if isConnected {
				klog.Infof("success connect to：%s\n", endpoint)
			} else {
				klog.Infof("Unable to connect to: %s \n", endpoint)
				componentResult = append(componentResult, kubeeyev1alpha2.ComponentResultItem{Name: item.Name, Namespace: item.Namespace, Endpoint: endpoint})
			}

		}

		marshal, err := json.Marshal(componentResult)
		if err != nil {
			return nil, err
		}

		return marshal, nil
	}
	return nil, nil
}

func (o *componentInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) *kubeeyev1alpha2.InspectResult {
	var componentResult []kubeeyev1alpha2.ComponentResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &componentResult)
	if err != nil {
		return resultCr
	}

	resultCr.Spec.ComponentResult = componentResult

	return resultCr
}

func checkConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func GetInspectComponent(ctx context.Context, clients *kube.KubernetesClient, components string) ([]corev1.Service, error) {
	var filterComponents []string
	if components != "" {
		filterComponents = strings.Split(components, "|")
	}
	services, err := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var component []corev1.Service
	for _, item := range services.Items {
		if len(filterComponents) > 0 {
			_, b := utils.ArrayFind(item.Name, filterComponents)
			if !b {
				continue
			}
		}
		if item.Spec.ClusterIP != "None" {
			component = append(component, item)
		}

	}

	return component, err
}
