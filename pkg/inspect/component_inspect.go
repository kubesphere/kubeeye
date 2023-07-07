package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type componentInspect struct {
}

func init() {
	RuleOperatorMap[constant.Component] = &componentInspect{}
}

func (o *componentInspect) CreateJobTask(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobPhase, error) {

	var jobNames []kubeeyev1alpha2.JobPhase

	job := template.InspectJobsTemplate(ctx, clients, jobRule.JobName, task, "", nil, constant.Component)

	_, err := clients.ClientSet.BatchV1().Jobs("kubeeye-system").Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning})

	return jobNames, nil
}

func (o *componentInspect) RunInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, currentJobName string, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(task.Spec.Rules, func(m kubeeyev1alpha2.JobRule) bool {
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

func (o *componentInspect) GetResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var componentResult []kubeeyev1alpha2.ComponentResultItem
	err := json.Unmarshal(result.BinaryData[constant.Result], &componentResult)
	if err != nil {
		return err
	}
	var ownerRefBol = true

	var inspectResult kubeeyev1alpha2.InspectResult
	inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.Component)
	inspectResult.OwnerReferences = []metav1.OwnerReference{{
		APIVersion:         task.APIVersion,
		Kind:               task.Kind,
		Name:               task.Name,
		UID:                task.UID,
		Controller:         &ownerRefBol,
		BlockOwnerDeletion: &ownerRefBol,
	}}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.ComponentResult = componentResult
	err = c.Create(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
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
