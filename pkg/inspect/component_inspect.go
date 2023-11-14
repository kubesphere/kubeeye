package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net"
	"time"
)

type componentInspect struct {
}

func init() {
	RuleOperatorMap[constant.Component] = &componentInspect{}
}

func (c *componentInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var components kubeeyev1alpha2.ComponentRule
		err := json.Unmarshal(phase.RunRule, &components)
		if err != nil {
			return nil, err
		}
		component, err := GetInspectComponent(ctx, clients, &components)
		if err != nil {
			return nil, err
		}
		var componentResult []kubeeyev1alpha2.ComponentResultItem
		for _, item := range component {
			endpoint := fmt.Sprintf("%s.%s.svc.cluster.local:%d", item.Name, item.Namespace, item.Spec.Ports[0].Port)
			isConnected := checkConnection(endpoint)
			componentResultItem := kubeeyev1alpha2.ComponentResultItem{
				Namespace:  item.Namespace,
				Endpoint:   endpoint,
				BaseResult: kubeeyev1alpha2.BaseResult{Name: item.Name, Assert: !isConnected},
			}
			if isConnected {
				klog.Infof("success connect to：%s\n", endpoint)
			} else {
				klog.Infof("Unable to connect to: %s \n", endpoint)
				componentResultItem.Level = kubeeyev1alpha2.WarningLevel
			}
			componentResult = append(componentResult, componentResultItem)
		}

		marshal, err := json.Marshal(componentResult)
		if err != nil {
			return nil, err
		}

		return marshal, nil
	}
	return nil, nil
}

func (c *componentInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {
	var componentResult []kubeeyev1alpha2.ComponentResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &componentResult)
	if err != nil {
		return nil, err
	}
	if componentResult == nil {
		return resultCr, nil
	}

	resultCr.Spec.ComponentResult = componentResult

	return resultCr, nil
}

func checkConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func GetInspectComponent(ctx context.Context, clients *kube.KubernetesClient, components *kubeeyev1alpha2.ComponentRule) ([]corev1.Service, error) {

	services, err := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	component, _ := utils.ArrayFilter(services.Items, func(v corev1.Service) bool {
		if v.Spec.ClusterIP == "None" {
			return false
		}
		if components.IncludeComponent != nil {
			_, isExist := utils.ArrayFind(v.Name, components.IncludeComponent)
			return isExist
		}
		_, excludeExist := utils.ArrayFind(v.Name, components.ExcludeComponent)
		return !excludeExist
	})

	return component, nil
}
