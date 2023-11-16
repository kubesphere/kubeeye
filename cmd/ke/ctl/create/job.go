package create

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"os"
)

type Options struct {
	Clients      *kube.KubernetesClient
	TaskName     string
	ResultName   string
	JobType      string
	k8sInformers informers.SharedInformerFactory
}

func NewJobOptions() *Options {
	return &Options{}
}

func NewJobCmd() *cobra.Command {
	o := NewJobOptions()
	jobCmd := &cobra.Command{
		Use:   "job",
		Short: "create inspect job task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Run(cmd.Context()); err != nil {
				return err
			}
			fmt.Println(o.TaskName, "inspect success")
			return nil
		},
	}
	o.addFlags(jobCmd)
	return jobCmd
}

func (o *Options) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ResultName, "result-name", "", " result config name")
	cmd.Flags().StringVar(&o.TaskName, "task-name", "", "task name")
	cmd.Flags().StringVar(&o.JobType, "job-type", "", "execute job type")
}

func CheckAgr(o *Options) error {
	if utils.IsEmptyValue(o.TaskName) || utils.IsEmptyValue(o.ResultName) {
		return fmt.Errorf("taskName  or resultName Incomplete parameters")
	}
	return nil
}

func (o *Options) Run(cmd context.Context) error {
	err := CheckAgr(o)
	if err != nil {
		return err
	}
	k8sConfig, err := kube.GetKubeConfigInCluster()
	if err != nil {
		klog.Error(fmt.Sprintf("Failed to load cluster clients. err:%s", err))
		return err
	}
	var kc kube.KubernetesClient
	o.Clients, err = kc.K8SClients(k8sConfig)
	if err != nil {
		klog.Error(err, ",Failed to load cluster clients")
		return err
	}
	factory := informers.NewSharedInformerFactory(o.Clients.ClientSet, 0)

	k8sGVRs := map[schema.GroupVersion][]string{
		{Group: "", Version: "v1"}: {
			"namespaces",
			"nodes",
			"pods",
			"services",
			"configmaps",
		},
	}
	for groupVersion, resourcesNames := range k8sGVRs {
		resource := schema.GroupVersionResource{
			Group:   groupVersion.Group,
			Version: groupVersion.Version,
		}
		for i := range resourcesNames {
			resource.Resource = resourcesNames[i]
			_, err = factory.ForResource(resource)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

	}

	factory.Start(cmd.Done())

	o.k8sInformers = factory

	err = o.jobInspect(cmd)
	if err != nil {
		return fmt.Errorf("kubeeye inspect failed with error: %s,%v", err, err)
	}
	return nil
}

func (o *Options) jobInspect(ctx context.Context) error {
	var jobRule []kubeeyev1alpha2.JobRule

	rule, err := o.Clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"})})
	if err != nil {
		klog.Errorf("failed to get  inspect Rule. err:%s", err)
		return err
	}
	for _, item := range rule.Items {
		var tempRule []kubeeyev1alpha2.JobRule
		data := item.BinaryData[constant.Data]
		err = json.Unmarshal(data, &tempRule)
		if err != nil {
			return err
		}
		jobRule = append(jobRule, tempRule...)
	}

	inspectInterface, status := inspect.RuleOperatorMap[o.JobType]
	if status {
		result, err := inspectInterface.RunInspect(ctx, jobRule, o.Clients, o.ResultName, o.k8sInformers)
		if err != nil {
			return err
		}
		node := o.findJobRunNode(ctx)
		resultConfigMap := template.BinaryConfigMapTemplate(o.ResultName, constant.DefaultNamespace, result, true, map[string]string{constant.LabelTaskName: o.TaskName, constant.LabelNodeName: node, constant.LabelRuleType: o.JobType})
		_, err = o.Clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Create(ctx, resultConfigMap, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("create configMap failed. err:%s", err)
		}
	}
	return nil
}

func (o *Options) findJobRunNode(ctx context.Context) string {
	pods, err := o.Clients.ClientSet.CoreV1().Pods(constant.DefaultNamespace).List(ctx, v1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{"job-name": o.ResultName})})
	if err != nil {
		klog.Error("failed to get pods ", err)
		return ""
	}

	return pods.Items[0].Spec.NodeName
}
