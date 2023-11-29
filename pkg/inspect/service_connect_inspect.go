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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"net"
	"time"
)

type serviceConnectInspect struct {
}

func init() {
	RuleOperatorMap[constant.ServiceConnect] = &serviceConnectInspect{}
}

func (c *serviceConnectInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, informers informers.SharedInformerFactory, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})

	if exist {
		var components []kubeeyev1alpha2.ServiceConnectRuleItem
		err := json.Unmarshal(phase.RunRule, &components)
		if err != nil {
			return nil, err
		}
		component, err := GetInspectComponent(ctx, clients, components)
		if err != nil {
			return nil, err
		}
		var componentResult []kubeeyev1alpha2.ServiceConnectResultItem
		for _, item := range component {
			isConnected := c.checkConnection(item.Rule)
			componentResultItem := kubeeyev1alpha2.ServiceConnectResultItem{
				Endpoint:   item.Rule,
				BaseResult: kubeeyev1alpha2.BaseResult{Name: item.Name, Assert: !isConnected},
			}
			if isConnected {
				klog.Infof("success connect to：%s\n", item.Rule)
			} else {
				klog.Infof("Unable to connect to: %s \n", item.Rule)
				componentResultItem.Level = item.Level
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

func (c *serviceConnectInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {
	var componentResult []kubeeyev1alpha2.ServiceConnectResultItem
	err := json.Unmarshal(resultCm.BinaryData[constant.Data], &componentResult)
	if err != nil {
		return nil, err
	}
	if componentResult == nil {
		return resultCr, nil
	}

	resultCr.Spec.ServiceConnectResult = componentResult

	return resultCr, nil
}

func (c *serviceConnectInspect) checkConnection(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func GetInspectComponent(ctx context.Context, clients *kube.KubernetesClient, serviceConnectRule []kubeeyev1alpha2.ServiceConnectRuleItem) (map[string]kubeeyev1alpha2.ServiceConnectRuleItem, error) {
	var inspectService = make(map[string]kubeeyev1alpha2.ServiceConnectRuleItem)
	list, err := clients.ClientSet.CoreV1().Services(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, service := range serviceConnectRule {
		if !utils.IsEmptyValue(service.Workspace) {
			namespaces := GetNameSpacesForWorkSpace(ctx, clients, service.Workspace)
			for _, namespace := range namespaces {
				for _, s := range GetServicesForNameSpace(list.Items, namespace.Name) {
					inspectService[fmt.Sprintf("%s/%s", s.Name, s.Namespace)] = kubeeyev1alpha2.ServiceConnectRuleItem{
						RuleItemBases: kubeeyev1alpha2.RuleItemBases{
							Name:  s.Name,
							Rule:  fmt.Sprintf("%s.%s.svc.cluster.local:%d", s.Name, s.Namespace, s.Spec.Ports[0].Port),
							Level: service.Level,
						},
					}

				}
			}
		} else if !utils.IsEmptyValue(service.Namespace) {
			for _, s := range GetServicesForNameSpace(list.Items, service.Namespace) {
				inspectService[fmt.Sprintf("%s/%s", s.Name, s.Namespace)] = kubeeyev1alpha2.ServiceConnectRuleItem{
					RuleItemBases: kubeeyev1alpha2.RuleItemBases{
						Name:  s.Name,
						Rule:  fmt.Sprintf("%s.%s.svc.cluster.local:%d", s.Name, s.Namespace, s.Spec.Ports[0].Port),
						Level: service.Level,
					},
				}

			}
		} else {
			if s, ok := GetServices(list.Items, service.Name); ok {
				inspectService[fmt.Sprintf("%s/%s", s.Name, s.Namespace)] = kubeeyev1alpha2.ServiceConnectRuleItem{
					RuleItemBases: kubeeyev1alpha2.RuleItemBases{
						Name:  s.Name,
						Rule:  fmt.Sprintf("%s.%s.svc.cluster.local:%d", s.Name, s.Namespace, s.Spec.Ports[0].Port),
						Level: service.Level,
					},
				}
			}
		}
	}
	return inspectService, nil
}

func GetNameSpacesForWorkSpace(ctx context.Context, clients *kube.KubernetesClient, workspace string) []corev1.Namespace {
	var namespaces []corev1.Namespace
	list, err := clients.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelSystemWorkspace: workspace})})
	if err != nil {
		return namespaces
	}
	namespaces = append(namespaces, list.Items...)
	return namespaces
}

func GetServicesForNameSpace(services []corev1.Service, namespace string) []corev1.Service {

	filter, _ := utils.ArrayFilter(services, func(v corev1.Service) bool {
		return namespace == v.Namespace && v.Spec.Type != corev1.ServiceTypeExternalName && v.Spec.ClusterIP != "None"
	})

	return filter
}

func GetServices(services []corev1.Service, name string) (corev1.Service, bool) {
	_, ok, service := utils.ArrayFinds(services, func(v corev1.Service) bool {
		return name == v.Namespace && v.Spec.Type != corev1.ServiceTypeExternalName && v.Spec.ClusterIP != "None"
	})
	return service, ok
}
