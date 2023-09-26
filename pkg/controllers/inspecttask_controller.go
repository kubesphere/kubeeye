/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/rules"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"math"
	"os"
	"path"
	"sync"
	"time"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/inspect"
	"github.com/kubesphere/kubeeye/pkg/kube"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InspectTaskReconciler reconciles a InspectTask object
type InspectTaskReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	K8sClients *kube.KubernetesClient
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.kubesphere.io,resources=clusters,verbs=get
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes;namespaces;services,verbs=list;get
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=deletecollection
//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=create;get
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs="*"

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *InspectTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//logger := log.FromContext(ctx).WithName(req.NamespacedName.String())

	inspectTask := &kubeeyev1alpha2.InspectTask{}
	err := r.Get(ctx, req.NamespacedName, inspectTask)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Error("failed to get inspect task. ", err)
		return ctrl.Result{}, err
	}
	if inspectTask.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, inspectTask.Finalizers); !b {
			inspectTask.Finalizers = append(inspectTask.Finalizers, Finalizers)
			err = r.Client.Update(ctx, inspectTask)
			if err != nil {
				klog.Error("Failed to inspect plan add finalizers", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		newFinalizers := utils.SliceRemove(Finalizers, inspectTask.Finalizers)
		inspectTask.Finalizers = newFinalizers.([]string)
		klog.Infof("inspect task is being deleted")
		err = r.Client.Update(ctx, inspectTask)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers. ", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if inspectTask.Status.Status.IsSucceeded() || inspectTask.Status.Status.IsFailed() {
		return ctrl.Result{}, nil
	}

	if inspectTask.Status.StartTimestamp.IsZero() {
		inspectTask.Status.StartTimestamp = &metav1.Time{Time: time.Now()}
		inspectTask.Status.Status = kubeeyev1alpha2.PhaseRunning
		err = r.Status().Update(ctx, inspectTask)
		if err != nil {
			klog.Error("Failed to update inspect plan status. ", err)
			return ctrl.Result{}, err
		}
		err = r.updatePlanStatus(ctx, kubeeyev1alpha2.PhaseRunning, inspectTask.Labels[constant.LabelPlanName], inspectTask.Name)
		if err != nil {

		}
		return ctrl.Result{}, nil
	}
	inspectTask.Status.ClusterInfo, err = r.getClusterInfo(ctx)
	if err != nil {
		klog.Error("failed to get cluster info. ", err)
		return ctrl.Result{}, err
	}
	var kubeEyeConfig conf.KubeEyeConfig
	kubeEyeConfig, err = kube.GetKubeEyeConfig(ctx, r.K8sClients)
	if err != nil {
		klog.Error("Unable to get jobConfig")
		return ctrl.Result{}, err
	}

	getRules := r.getRules(ctx, inspectTask)
	if err != nil {
		klog.Error("failed get to inspectrules.", err)
		return ctrl.Result{}, err
	}

	if inspectTask.Spec.ClusterName != nil {
		var wait sync.WaitGroup
		wait.Add(len(inspectTask.Spec.ClusterName))
		for _, cluster := range inspectTask.Spec.ClusterName {
			go func(v kubeeyev1alpha2.Cluster) {
				defer wait.Done()
				clusterClient, err := kube.GetMultiClusterClient(ctx, r.K8sClients, v.Name)
				if err != nil {
					klog.Error(err, "Failed to get multi-cluster client.")
					return
				}
				err = r.initClusterInspect(ctx, clusterClient)
				if err != nil {
					klog.Errorf("failed To Initialize Cluster Configuration for Cluster Name:%s,err:%s", v, err)
					return
				}
				err = r.CreateInspect(ctx, v, inspectTask, getRules, clusterClient, kubeEyeConfig)
				if err != nil {
					klog.Error("failed to create inspect. ", err)
				}
			}(cluster)
		}
		wait.Wait()
	} else {
		err = r.initClusterInspect(ctx, r.K8sClients)
		if err != nil {
			klog.Errorf("failed To Initialize Cluster Configuration for Cluster Name:%s,err:%s", "default", err)
			return ctrl.Result{}, err
		}
		err = r.CreateInspect(ctx, kubeeyev1alpha2.Cluster{Name: "default"}, inspectTask, getRules, r.K8sClients, kubeEyeConfig)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	taskStatus := GetStatus(inspectTask)
	inspectTask.Status.Status = taskStatus
	err = r.Status().Update(ctx, inspectTask)
	if err != nil {
		klog.Error("failed to update inspect task. ", err)
		return ctrl.Result{}, err
	}
	klog.Infof("all job finished for taskName:%s", inspectTask.Name)

	err = r.updatePlanStatus(ctx, taskStatus, inspectTask.Labels[constant.LabelPlanName], inspectTask.Name)
	if err != nil {
		klog.Error("failed to update inspect plan comeToAnEnd status. ", err)
	}

	return ctrl.Result{}, nil

}

func createInspectRule(ctx context.Context, clients *kube.KubernetesClient, ruleGroup []kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask) ([]kubeeyev1alpha2.JobRule, error) {
	r := sortRuleOpaToAtLast(ruleGroup)
	marshal, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	_, err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Get(ctx, task.Name, metav1.GetOptions{})
	if err == nil {
		_ = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Delete(ctx, task.Name, metav1.DeleteOptions{})
	}

	configMapTemplate := template.BinaryConfigMapTemplate(task.Name, constant.DefaultNamespace, marshal, true, map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"})
	_, err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Create(ctx, configMapTemplate, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *InspectTaskReconciler) CreateInspect(ctx context.Context, cluster kubeeyev1alpha2.Cluster, task *kubeeyev1alpha2.InspectTask, ruleLists []kubeeyev1alpha2.InspectRule, clients *kube.KubernetesClient, kubeEyeConfig conf.KubeEyeConfig) error {

	inspectRule, inspectRuleNum, err := rules.ParseRules(ctx, clients, task.Name, ruleLists)
	if err != nil {
		return err
	}
	rule, err := createInspectRule(ctx, clients, inspectRule, task)
	if err != nil {
		return err
	}
	jobConfig := kubeEyeConfig.GetJobConfig(cluster.Name)
	JobPhase, err := r.createJobsInspect(ctx, task, clients, jobConfig, rule)
	if err != nil {
		return err
	}
	task.Status.JobPhase = append(task.Status.JobPhase, JobPhase...)
	task.Status.EndTimestamp = &metav1.Time{Time: time.Now()}
	task.Status.Duration = task.Status.EndTimestamp.Sub(task.Status.StartTimestamp.Time).String()
	task.Status.InspectRuleType = func() (data []string) {
		for k, v := range inspectRuleNum {
			if v > 0 {
				data = append(data, k)
			}
		}
		return data
	}()
	err = r.getInspectResultData(ctx, clients, task, cluster, inspectRuleNum)
	if err != nil {
		return err
	}
	return nil
}

func sortRuleOpaToAtLast(rule []kubeeyev1alpha2.JobRule) []kubeeyev1alpha2.JobRule {

	finds, b, OpaRule := utils.ArrayFinds(rule, func(i kubeeyev1alpha2.JobRule) bool {
		return i.RuleType == constant.Opa
	})
	if b {
		rule = append(rule[:finds], rule[finds+1:]...)
		rule = append(rule, OpaRule)
	}

	return rule
}
func GetStatus(task *kubeeyev1alpha2.InspectTask) kubeeyev1alpha2.Phase {
	if task.Status.JobPhase == nil {
		return kubeeyev1alpha2.PhaseFailed
	}
	_, ok, _ := utils.ArrayFinds(task.Status.JobPhase, func(m kubeeyev1alpha2.JobPhase) bool {
		return m.Phase.IsFailed()
	})
	if ok {
		return kubeeyev1alpha2.PhaseFailed
	}
	return kubeeyev1alpha2.PhaseSucceeded
}

func (r *InspectTaskReconciler) getClusterInfo(ctx context.Context) (kubeeyev1alpha2.ClusterInfo, error) {
	var clusterInfo kubeeyev1alpha2.ClusterInfo
	versionInfo, err := r.K8sClients.ClientSet.Discovery().ServerVersion()
	if err != nil {
		klog.Error(err, "Failed to get Kubernetes serverVersion.\n")
	}
	var serverVersion string
	if versionInfo != nil {
		serverVersion = versionInfo.Major + "." + versionInfo.Minor
	}
	_, nodesCount, err := kube.GetObjectCounts(ctx, r.K8sClients, conf.Nodes, conf.NoGroup)
	if err != nil {
		klog.Error(err, "Failed to get node number.")
	}
	_, namespacesCount, err := kube.GetObjectCounts(ctx, r.K8sClients, conf.Namespaces, conf.NoGroup)
	if err != nil {
		klog.Error(err, "Failed to get ns number.")
	}
	clusterInfo = kubeeyev1alpha2.ClusterInfo{ClusterVersion: serverVersion, NodesCount: nodesCount, NamespacesCount: namespacesCount}
	return clusterInfo, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectTask{}).
		Complete(r)
}

func (r *InspectTaskReconciler) createJobsInspect(ctx context.Context, inspectTask *kubeeyev1alpha2.InspectTask, clusterClient *kube.KubernetesClient, config *conf.JobConfig, inspectRule []kubeeyev1alpha2.JobRule) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	nodes := kube.GetNodes(ctx, clusterClient.ClientSet)
	concurrency := 5
	runNumber := math.Round(float64(len(nodes)) + float64(len(inspectRule))*0.1)
	if runNumber > 5 {
		concurrency = int(runNumber)
	}
	var wg sync.WaitGroup
	var mutex sync.Mutex
	semaphore := make(chan struct{}, concurrency)
	for _, rule := range inspectRule {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(v kubeeyev1alpha2.JobRule) {
			defer func() {
				wg.Done()
				<-semaphore
			}()
			if isTimeout(inspectTask.CreationTimestamp, inspectTask.Spec.Timeout) {
				jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseFailed})
				return
			}
			if err := isExistsJob(ctx, clusterClient, v.JobName); err != nil {
				mutex.Lock()
				jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseSucceeded})
				mutex.Unlock()
				return
			}
			inspectInterface, status := inspect.RuleOperatorMap[v.RuleType]
			if status {
				klog.Infof("Job %s created", v.JobName)
				jobTask, err := inspectInterface.CreateJobTask(ctx, clusterClient, &v, inspectTask, config)
				if err != nil {
					klog.Errorf("create job error. error:%s", err)
					jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseFailed})
					return
				}
				resultJob := r.waitForJobCompletionGetResult(ctx, clusterClient, v.JobName, jobTask, inspectTask.Spec.Timeout)
				mutex.Lock()
				jobNames = append(jobNames, *resultJob)
				mutex.Unlock()
				klog.Infof("Job %s completed", v.JobName)
			} else {
				klog.Errorf("%s not found", v.RuleType)
			}
		}(rule)
	}
	wg.Wait()

	err := r.cleanConfig(ctx, clusterClient, inspectTask.Spec.ClusterName)
	if err != nil {
		return nil, err
	}
	return jobNames, nil
}

func isExistsJob(ctx context.Context, clients *kube.KubernetesClient, jobName string) error {
	_, err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil && kubeErr.IsNotFound(err) {
		return nil
	}
	klog.Errorf("job already exists for name:%s", jobName)
	return fmt.Errorf("job already exists for name:%s", jobName)
}

func (r *InspectTaskReconciler) waitForJobCompletionGetResult(ctx context.Context, clients *kube.KubernetesClient, jobName string, jobPhase *kubeeyev1alpha2.JobPhase, timeout string) *kubeeyev1alpha2.JobPhase {
	for {
		klog.Infof("wait job run complete for name:%s", jobName)
		jobInfo, err := clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			klog.Infof("failed to get job info for name:%s,err:%s", jobName, err)
			jobPhase.Phase = kubeeyev1alpha2.PhaseFailed
			return jobPhase
		}
		if isTimeout(jobInfo.CreationTimestamp, timeout) {
			klog.Infof("job executed timeout for name:%s", jobName)
			jobPhase.Phase = kubeeyev1alpha2.PhaseFailed
			return jobPhase
		}
		if jobInfo.Status.Conditions != nil && jobInfo.Status.Conditions[0].Type == v1.JobFailed {
			klog.Infof("failed to job executed successful for name:%s", jobName)
			jobPhase.Phase = kubeeyev1alpha2.PhaseFailed
			return jobPhase
		}
		if jobInfo.Status.CompletionTime != nil && !jobInfo.Status.CompletionTime.IsZero() && jobInfo.Status.Active == 0 {
			jobPhase.Phase = kubeeyev1alpha2.PhaseSucceeded
			background := metav1.DeletePropagationBackground
			_ = clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &background})
			return jobPhase
		}

		time.Sleep(10 * time.Second)
	}

}

func (r *InspectTaskReconciler) getInspectResultData(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask, cluster kubeeyev1alpha2.Cluster, ruleNum map[string]int) error {
	configs, err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{constant.LabelTaskName: task.Name}),
	})

	if err != nil {
		return err
	}
	var ownerRefBol = true
	inspectResult := kubeeyev1alpha2.InspectResult{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s-result", cluster.Name, task.Name),
			Labels: map[string]string{constant.LabelTaskName: task.Name},
			Annotations: map[string]string{
				constant.AnnotationStartTime:     task.Status.StartTimestamp.Format("2006-01-02 15:04:05"),
				constant.AnnotationEndTime:       task.Status.EndTimestamp.Format("2006-01-02 15:04:05"),
				constant.AnnotationInspectPolicy: string(task.Spec.InspectPolicy),
			},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         task.APIVersion,
				Kind:               task.Kind,
				Name:               task.Name,
				UID:                task.UID,
				Controller:         &ownerRefBol,
				BlockOwnerDeletion: &ownerRefBol,
			}},
		},
		Spec: kubeeyev1alpha2.InspectResultSpec{
			InspectRuleTotal: ruleNum,
			InspectCluster:   cluster,
		},
	}

	resultData := inspectResult.DeepCopy()
	for _, phase := range task.Status.JobPhase {
		if phase.Phase.IsSucceeded() {
			_, exists, configMap := utils.ArrayFinds(configs.Items, func(m corev1.ConfigMap) bool {
				return m.Name == phase.JobName
			})
			if exists {
				ruleType := configMap.Labels[constant.LabelRuleType]
				nodeName := configMap.Labels[constant.LabelNodeName]
				inspectInterface, status := inspect.RuleOperatorMap[ruleType]
				if status {
					klog.Infof("starting get %s result data", phase.JobName)
					_, err = inspectInterface.GetResult(nodeName, &configMap, resultData)
					if err != nil {
						klog.Error(err)
					}
				}
			}
		}
	}
	err = saveResultFile(resultData)
	if err != nil {
		return err
	}

	err = r.Create(ctx, &inspectResult)
	if err != nil {
		return err
	}

	err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelTaskName: task.Name})})
	if err != nil {
		return err
	}

	return nil
}

func saveResultFile(resultData *kubeeyev1alpha2.InspectResult) error {
	file, err := os.OpenFile(path.Join(constant.ResultPath, resultData.Name), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		klog.Error(err, "open file error")
		return err
	}
	defer file.Close()
	marshal, err := json.Marshal(resultData)
	if err != nil {
		klog.Error(err, "marshal error")
	}
	_, err = file.Write(marshal)
	if err != nil {
		klog.Error(err, "write file error")
		return err
	}
	return nil
}

func isTimeout(startTime metav1.Time, t string) bool {
	duration, err := time.ParseDuration(t)
	if err != nil {
		duration = constant.DefaultTimeout
	}
	return startTime.Add(duration).Before(time.Now())
}

// InitClusterInspect Initialize the relevant configuration items required for multi-cluster inspection
func (r *InspectTaskReconciler) initClusterInspect(ctx context.Context, clients *kube.KubernetesClient) error {

	_, err := clients.ClientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: constant.DefaultNamespace}}, metav1.CreateOptions{})
	if err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}

	_, err = clients.ClientSet.RbacV1().ClusterRoles().Create(ctx, template.GetClusterRoleTemplate(), metav1.CreateOptions{})
	if err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}
	_, err = clients.ClientSet.RbacV1().ClusterRoleBindings().Create(ctx, template.GetClusterRoleBindingTemplate(), metav1.CreateOptions{})
	if err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}

	_, err = clients.ClientSet.CoreV1().ServiceAccounts(constant.DefaultNamespace).Create(ctx, template.GetServiceAccountTemplate(), metav1.CreateOptions{})
	if err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (r *InspectTaskReconciler) cleanConfig(ctx context.Context, clients *kube.KubernetesClient, clusterName []kubeeyev1alpha2.Cluster) error {
	err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"}),
	})
	if err != nil {
		return err
	}
	err = clients.ClientSet.CoreV1().ServiceAccounts(constant.DefaultNamespace).Delete(ctx, template.GetServiceAccountTemplate().Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clients.ClientSet.RbacV1().ClusterRoleBindings().Delete(ctx, template.GetClusterRoleBindingTemplate().Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clients.ClientSet.RbacV1().ClusterRoles().Delete(ctx, template.GetClusterRoleTemplate().Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (r *InspectTaskReconciler) updatePlanStatus(ctx context.Context, phase kubeeyev1alpha2.Phase, planName string, taskName string) error {
	plan := &kubeeyev1alpha2.InspectPlan{}
	err := r.Get(ctx, types.NamespacedName{Name: planName}, plan)
	if err != nil {
		klog.Error(err, "get plan error")
		return err
	}
	for i, name := range plan.Status.TaskNames {
		if name.Name == taskName {
			plan.Status.TaskNames[i].TaskStatus = phase
			break
		}
	}
	timeNow := metav1.Now()
	if phase.IsRunning() {
		plan.Status.LastTaskStartTime = &timeNow
	} else {
		plan.Status.LastTaskEndTime = &timeNow
	}
	plan.Status.LastTaskStatus = phase
	err = r.Status().Update(ctx, plan)
	if err != nil {
		klog.Error(err, "update plan status error")
		return err
	}
	return nil
}

func (r *InspectTaskReconciler) getRules(ctx context.Context, task *kubeeyev1alpha2.InspectTask) (rules []kubeeyev1alpha2.InspectRule) {
	for _, name := range task.Spec.RuleNames {
		rule, err := r.K8sClients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			klog.Error(err, "get rule error")
			continue
		}
		rules = append(rules, *rule)
	}

	return rules
}
