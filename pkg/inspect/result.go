package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetFileChangeResult(ctx context.Context, c client.Client, jobs *v1.Job, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	//
	var inspectResult kubeeyev1alpha2.InspectResult
	err := c.Get(ctx, types.NamespacedName{
		Namespace: task.Namespace,
		Name:      fmt.Sprintf("%s-%s", task.Name, constant.FileChange),
	}, &inspectResult)
	var nodeInfoResult kubeeyev1alpha2.NodeInfoResult
	jsonErr := json.Unmarshal(result.BinaryData[constant.Result], &nodeInfoResult)
	if jsonErr != nil {
		klog.Error("failed to get result", jsonErr)
	}
	if err != nil {
		if kubeErr.IsNotFound(err) {
			var ownerRefBol = true
			resultRef := metav1.OwnerReference{
				APIVersion:         task.APIVersion,
				Kind:               task.Kind,
				Name:               task.Name,
				UID:                task.UID,
				Controller:         &ownerRefBol,
				BlockOwnerDeletion: &ownerRefBol,
			}
			inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
			inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.FileChange)
			inspectResult.Namespace = task.Namespace
			inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
			inspectResult.Spec.NodeInfoResult = map[string]kubeeyev1alpha2.NodeInfoResult{jobs.Spec.Template.Spec.NodeName: nodeInfoResult}
			err = c.Create(ctx, &inspectResult)
			if err != nil {
				klog.Error("Failed to create inspect result", err)
				return err
			}
			return nil
		}

	}
	inspectResult.Spec.NodeInfoResult[jobs.Spec.Template.Spec.NodeName] = nodeInfoResult
	err = c.Update(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to update inspect result", err)
		return err
	}
	return nil

}

func GetPrometheusResult(ctx context.Context, c client.Client, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var prometheus [][]map[string]string
	err := json.Unmarshal(result.BinaryData[constant.Result], &prometheus)
	if err != nil {
		return err
	}
	var ownerRefBol = true
	resultRef := metav1.OwnerReference{
		APIVersion:         task.APIVersion,
		Kind:               task.Kind,
		Name:               task.Name,
		UID:                task.UID,
		Controller:         &ownerRefBol,
		BlockOwnerDeletion: &ownerRefBol,
	}

	var inspectResult kubeeyev1alpha2.InspectResult
	inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.Prometheus)
	inspectResult.Namespace = task.Namespace
	inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.PrometheusResult = prometheus
	err = c.Create(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
}

func GetOpaResult(ctx context.Context, c client.Client, result *corev1.ConfigMap, task *kubeeyev1alpha2.InspectTask) error {
	var opaResult kubeeyev1alpha2.KubeeyeOpaResult
	err := json.Unmarshal(result.BinaryData[constant.Result], &opaResult)
	if err != nil {
		return err
	}
	var ownerRefBol = true
	resultRef := metav1.OwnerReference{
		APIVersion:         task.APIVersion,
		Kind:               task.Kind,
		Name:               task.Name,
		UID:                task.UID,
		Controller:         &ownerRefBol,
		BlockOwnerDeletion: &ownerRefBol,
	}

	var inspectResult kubeeyev1alpha2.InspectResult
	inspectResult.Name = fmt.Sprintf("%s-%s", task.Name, constant.Opa)
	inspectResult.Namespace = task.Namespace
	inspectResult.OwnerReferences = []metav1.OwnerReference{resultRef}
	inspectResult.Labels = map[string]string{constant.LabelName: task.Name}
	inspectResult.Spec.OpaResult = opaResult
	err = c.Create(ctx, &inspectResult)
	if err != nil {
		klog.Error("Failed to create inspect result", err)
		return err
	}
	return nil
}
