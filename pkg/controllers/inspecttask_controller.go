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
	kubeeyeInformers "github.com/kubesphere/kubeeye/clients/informers/externalversions/kubeeye"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/rules"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
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
	Scheme         *runtime.Scheme
	K8sClients     *kube.KubernetesClient
	KubeEyeFactory kubeeyeInformers.Interface
	K8sFactory     informers.SharedInformerFactory
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.kubesphere.io,resources=clusters,verbs=get
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspecttasks/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes;namespaces;services;secrets,verbs=list;get
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=create;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=deletecollection;list;get;watch
//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=create;get;delete
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
		err = r.updatePlanStatus(ctx, kubeeyev1alpha2.PhaseRunning, inspectTask.Labels[constant.LabelPlanName], inspectTask.Name)
		if err != nil {
			klog.Error("Failed to update plan status. ", err)
			return ctrl.Result{}, err
		}
		inspectTask.Status.StartTimestamp = &metav1.Time{Time: time.Now()}
		inspectTask.Status.Status = kubeeyev1alpha2.PhaseRunning
		err = r.Status().Update(ctx, inspectTask)
		if err != nil {
			klog.Error("Failed to update inspect plan status. ", err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	inspectTask.Status.ClusterInfo, err = r.getClusterInfo(ctx)
	if err != nil {
		klog.Error("failed to get cluster info. ", err)
		return ctrl.Result{}, err
	}
	var kubeEyeConfig conf.KubeEyeConfig
	kubeEyeConfig, err = kube.GetKubeEyeConfig(r.K8sFactory.Core())
	if err != nil {
		klog.Error("Unable to get jobConfig")
		return ctrl.Result{}, err
	}

	if inspectTask.Spec.ClusterName != nil {
		var wait sync.WaitGroup
		wait.Add(len(inspectTask.Spec.ClusterName))
		for _, cluster := range inspectTask.Spec.ClusterName {
			go func(c kubeeyev1alpha2.Cluster) {
				defer wait.Done()
				clusterClient, err := kube.GetMultiClusterClient(ctx, r.K8sClients, c.Name)
				if err != nil {
					klog.Error(err, "Failed to get multi-cluster client.")
					return
				}
				err = r.initClusterInspectConfig(ctx, clusterClient)
				if err != nil {
					klog.Errorf("failed To Initialize Cluster Configuration for Cluster Name:%s,err:%s", c, err)
					return
				}
				err = r.createInspect(ctx, c, inspectTask, clusterClient, kubeEyeConfig)
				if err != nil {
					klog.Error("failed to create inspect. ", err)
				}
			}(cluster)
		}
		wait.Wait()
	} else {
		err = r.initClusterInspectConfig(ctx, r.K8sClients)
		if err != nil {
			klog.Errorf("failed To Initialize Cluster Configuration for Cluster Name:%s,err:%s", "default", err)
			return ctrl.Result{}, err
		}
		err = r.createInspect(ctx, kubeeyev1alpha2.Cluster{Name: "default"}, inspectTask, r.K8sClients, kubeEyeConfig)
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

func (r *InspectTaskReconciler) createInspect(ctx context.Context, cluster kubeeyev1alpha2.Cluster, task *kubeeyev1alpha2.InspectTask, clients *kube.KubernetesClient, kubeEyeConfig conf.KubeEyeConfig) error {
	e := rules.NewExecuteRuleOptions(clients, task)

	mergeRule, err := e.MergeRule(r.getRules(task))
	if err != nil {
		return err
	}

	createInspectRule, err := e.CreateInspectRule(ctx, e.GenerateJob(ctx, mergeRule))
	if err != nil {
		return err
	}
	jobConfig := kubeEyeConfig.GetClusterJobConfig(cluster.Name)

	task.Status.EndTimestamp = &metav1.Time{Time: time.Now()}
	task.Status.Duration = task.Status.EndTimestamp.Sub(task.Status.StartTimestamp.Time).String()
	task.Status.InspectRuleType = func() (data []string) {
		for k, v := range e.GetRuleTotal() {
			if v > 0 {
				data = append(data, k)
			}
		}
		return data
	}()
	result := r.GenerateResult(task, cluster, clients, e.GetRuleTotal())
	deepCopyResult := result.DeepCopy()
	JobPhase, err := r.createJobsInspect(ctx, task, clients, jobConfig, createInspectRule, deepCopyResult)
	if err != nil {
		return err
	}
	task.Status.JobPhase = append(task.Status.JobPhase, JobPhase...)

	err = r.Create(ctx, &result)
	if err != nil {
		return err
	}
	err = r.cleanClusterInspectConfig(ctx, clients, task)
	if err != nil {
		return err
	}
	return nil
}

func (r *InspectTaskReconciler) GenerateResult(task *kubeeyev1alpha2.InspectTask, cluster kubeeyev1alpha2.Cluster, clients *kube.KubernetesClient, ruleNum map[string]int) kubeeyev1alpha2.InspectResult {
	var ownerRefBol = true
	resultName := fmt.Sprintf("%s-%s-result", cluster.Name, task.Name)

	list, err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelTaskName: task.Name})})
	if err == nil && len(list.Items) > 0 {
		err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelTaskName: task.Name})})
		if err != nil {
			klog.Error("failed to delete inspect result")
		}
	}

	file, err := os.Open(path.Join(constant.ResultPath, resultName))
	if err == nil {
		defer file.Close()
		err = os.Remove(path.Join(constant.ResultPath, resultName))
		if err != nil {
			klog.Error("failed to delete result ")
		}
	}

	return kubeeyev1alpha2.InspectResult{
		ObjectMeta: metav1.ObjectMeta{Name: resultName,
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

func (r *InspectTaskReconciler) createJobsInspect(ctx context.Context, task *kubeeyev1alpha2.InspectTask, kubeClient *kube.KubernetesClient, config *conf.JobConfig, jobRules []kubeeyev1alpha2.JobRule, resultData *kubeeyev1alpha2.InspectResult) ([]kubeeyev1alpha2.JobPhase, error) {
	var jobNames []kubeeyev1alpha2.JobPhase
	nodes := kube.GetNodes(ctx, kubeClient.ClientSet)

	var wg sync.WaitGroup
	var mutex sync.Mutex
	semaphore := make(chan struct{}, computedDeployNum(len(nodes), len(jobRules)))
	for _, rule := range jobRules {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(v kubeeyev1alpha2.JobRule) {
			defer func() {
				wg.Done()
				<-semaphore
			}()
			if isTimeout(task.CreationTimestamp, task.Spec.Timeout) {
				jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseFailed})
				return
			}
			_, status := inspect.RuleOperatorMap[v.RuleType]
			if status {
				if checkJobIsDeploy(nodes, getIncompleteJob(ctx, kubeClient, task, v.RuleType), v) {
					jobTask, err := createInspectJob(ctx, kubeClient, &v, task, config, v.RuleType)
					if err != nil {
						klog.Errorf("create job error. error:%s", err)
						jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseFailed})
						return
					}
					klog.Infof("Job %s starting created", v.JobName)
					resultJob := r.waitForJobCompletionGetResult(ctx, kubeClient, v.JobName, jobTask, task.Spec.Timeout)
					mutex.Lock()
					jobNames = append(jobNames, *resultJob)
					err = r.getInspectResultData(ctx, kubeClient, resultData, resultJob.JobName)
					if err != nil {
						klog.Error("failed to get inspect result data", err)
					}
					mutex.Unlock()
				} else {
					klog.Errorf("failed  to deploy job with name %s", v.JobName)
					jobNames = append(jobNames, kubeeyev1alpha2.JobPhase{JobName: v.JobName, Phase: kubeeyev1alpha2.PhaseFailed})
				}
				klog.Infof("Job %s completed", v.JobName)
			} else {
				klog.Errorf("%s not found", v.RuleType)
			}
		}(rule)
	}
	wg.Wait()

	return jobNames, nil
}

func computedDeployNum(nodeNum int, jobRulesNum int) int {
	concurrency := 5
	runNumber := math.Round(float64(nodeNum) + float64(jobRulesNum)*0.1)
	if runNumber > 5 {
		concurrency = int(runNumber)
	}
	return concurrency
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
			err = clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &background})
			if err != nil {
				klog.Infof("failed to delete job:%s , err:%s", jobName, err)
			}
			return jobPhase
		}

		time.Sleep(10 * time.Second)
	}

}

func (r *InspectTaskReconciler) getInspectResultData(ctx context.Context, clients *kube.KubernetesClient, resultData *kubeeyev1alpha2.InspectResult, jobName string) error {
	configMap, err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Get(ctx, jobName,
		metav1.GetOptions{})
	if err != nil {
		return err
	}

	ruleType := configMap.Labels[constant.LabelRuleType]
	nodeName := configMap.Labels[constant.LabelNodeName]
	inspectInterface, status := inspect.RuleOperatorMap[ruleType]
	if status {
		klog.Infof("starting get %s result data", jobName)
		_, err = inspectInterface.GetResult(nodeName, configMap, resultData)
		if err != nil {
			klog.Error(err)
		}
	}

	err = saveResultFile(resultData)
	if err != nil {
		return err
	}

	err = clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).Delete(ctx, jobName, metav1.DeleteOptions{})
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
func (r *InspectTaskReconciler) initClusterInspectConfig(ctx context.Context, clients *kube.KubernetesClient) error {

	_, err := clients.ClientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: constant.DefaultNamespace}}, metav1.CreateOptions{})
	if err != nil && !kubeErr.IsAlreadyExists(err) {
		return err
	}

	_, err = clients.ClientSet.RbacV1().ClusterRoles().Create(ctx, template.GetClusterRoleTemplate(), metav1.CreateOptions{})
	if err != nil && !kubeErr.IsAlreadyExists(err) {
		return err
	}
	_, err = clients.ClientSet.RbacV1().ClusterRoleBindings().Create(ctx, template.GetClusterRoleBindingTemplate(), metav1.CreateOptions{})
	if err != nil && !kubeErr.IsAlreadyExists(err) {
		return err
	}

	_, err = clients.ClientSet.CoreV1().ServiceAccounts(constant.DefaultNamespace).Create(ctx, template.GetServiceAccountTemplate(), metav1.CreateOptions{})
	if err != nil && !kubeErr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *InspectTaskReconciler) cleanClusterInspectConfig(ctx context.Context, clients *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask) error {
	// clean temp inspect rule
	err := clients.ClientSet.CoreV1().ConfigMaps(constant.DefaultNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{constant.LabelInspectRuleGroup: "inspect-rule-temp"}),
	})
	if err != nil && !kubeErr.IsNotFound(err) {
		return err
	}

	err = clients.ClientSet.CoreV1().ServiceAccounts(constant.DefaultNamespace).Delete(ctx, template.GetServiceAccountTemplate().Name, metav1.DeleteOptions{})
	if err != nil && !kubeErr.IsNotFound(err) {
		return err
	}
	err = clients.ClientSet.RbacV1().ClusterRoleBindings().Delete(ctx, template.GetClusterRoleBindingTemplate().Name, metav1.DeleteOptions{})
	if err != nil && !kubeErr.IsNotFound(err) {
		return err
	}
	err = clients.ClientSet.RbacV1().ClusterRoles().Delete(ctx, template.GetClusterRoleTemplate().Name, metav1.DeleteOptions{})
	if err != nil && !kubeErr.IsNotFound(err) {
		return err
	}
	return nil
}

func (r *InspectTaskReconciler) updatePlanStatus(ctx context.Context, phase kubeeyev1alpha2.Phase, planName string, taskName string) error {

	plan, err := r.KubeEyeFactory.V1alpha2().InspectPlans().Lister().Get(planName)
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

func (r *InspectTaskReconciler) getRules(task *kubeeyev1alpha2.InspectTask) (rules []kubeeyev1alpha2.InspectRule) {
	for _, v := range task.Spec.RuleNames {
		rule, err := r.KubeEyeFactory.V1alpha2().InspectRules().Lister().Get(v.Name)
		if err != nil {
			klog.Error(err, "get rule error")
			continue
		}
		rules = append(rules, *rule)
	}

	return rules
}

func createInspectJob(ctx context.Context, clients *kube.KubernetesClient, jobRule *kubeeyev1alpha2.JobRule, task *kubeeyev1alpha2.InspectTask, config *conf.JobConfig, ruleType string) (*kubeeyev1alpha2.JobPhase, error) {

	nodeName, err := GetDeploySchedule(jobRule.RunRule)
	if err != nil && ruleType != constant.ServiceConnect && ruleType != constant.Component {
		return nil, fmt.Errorf("%s:%s", ruleType, err.Error())
	}

	o := template.JobTemplateOptions{
		JobConfig: config,
		JobName:   jobRule.JobName,
		Task:      task,
		NodeName:  nodeName,
		RuleType:  ruleType,
	}

	_, err = clients.ClientSet.BatchV1().Jobs(constant.DefaultNamespace).Create(ctx, template.GeneratorJobTemplate(o), metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Failed to create Jobs  for node name:%s,err:%s", err, err)
		return nil, err
	}
	return &kubeeyev1alpha2.JobPhase{JobName: jobRule.JobName, Phase: kubeeyev1alpha2.PhaseRunning}, nil
}

func GetDeploySchedule(r []byte) (string, error) {
	var data []map[string]interface{}
	err := json.Unmarshal(r, &data)
	if len(data) == 0 || err != nil {
		return "", fmt.Errorf("rule is empty")
	}
	v := data[0]["nodeName"]

	if !utils.IsEmptyValue(v) {
		return v.(string), nil
	}

	return "", nil
}

func checkJobIsDeploy(allNode []corev1.Node, inComplete []corev1.Pod, job kubeeyev1alpha2.JobRule) bool {
	nodeStatus := make(map[string]bool, len(allNode))
	for _, n := range allNode {
		if kube.IsNodesReady(n) {
			nodeStatus[n.Name] = kube.IsNodesReady(n)
		}
	}

	if len(nodeStatus) == 0 {
		klog.Error("暂无可部署的节点", job.JobName)
		return false
	}

	nodeName, err := GetDeploySchedule(job.RunRule)
	if err != nil || utils.IsEmptyValue(nodeName) {
		klog.Info("空Node,正常部署")
		return true
	}

	_, isDeploy := nodeStatus[nodeName]
	if !isDeploy {
		klog.Errorf("找不到可部署的节点：%s,jobName:%s", nodeName, job.JobName)
		return false
	}

	_, exist, _ := utils.ArrayFinds(inComplete, func(m corev1.Pod) bool {
		return nodeName == m.Spec.NodeName
	})
	if exist {
		klog.Errorf("节点：%s存在未完成的任务,jobName:%s", nodeName, job.JobName)
		return false
	}
	klog.Infof("正常部署%s", nodeName)
	return true
}

func getIncompleteJob(ctx context.Context, kubeClient *kube.KubernetesClient, task *kubeeyev1alpha2.InspectTask, ruleType string) []corev1.Pod {
	list, err := kubeClient.ClientSet.CoreV1().Pods(constant.DefaultNamespace).List(ctx, metav1.ListOptions{LabelSelector: labels.FormatLabels(map[string]string{constant.LabelPlanName: task.Labels[constant.LabelPlanName], constant.LabelRuleType: ruleType})})
	if err != nil {
		klog.Error("failed to getIncompleteJob ")
		return nil
	}
	return list.Items
}
