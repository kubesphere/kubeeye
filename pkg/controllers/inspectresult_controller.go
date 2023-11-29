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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	kubeeyeInformers "github.com/kubesphere/kubeeye/clients/informers/externalversions/kubeeye"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/message"
	"github.com/kubesphere/kubeeye/pkg/output"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"os"
	"path"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InspectResultReconciler reconciles a InspectResult object
type InspectResultReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	KubeEyeFactory kubeeyeInformers.Interface
	K8sFactory     informers.SharedInformerFactory
}

//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=inspectresults/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InspectResult object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *InspectResultReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result := &kubeeyev1alpha2.InspectResult{}
	err := r.Get(ctx, req.NamespacedName, result)
	if err != nil {
		if kubeErr.IsNotFound(err) {
			klog.Infof("inspect rule is not found;name:%s\n", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if result.DeletionTimestamp.IsZero() {
		if _, b := utils.ArrayFind(Finalizers, result.Finalizers); !b {
			result.Finalizers = append(result.Finalizers, Finalizers)
			err = r.Client.Update(ctx, result)
			if err != nil {
				klog.Error("Failed to inspect result add finalizers", err)
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

	} else {
		newFinalizers := utils.SliceRemove(Finalizers, result.Finalizers)
		result.Finalizers = newFinalizers.([]string)
		klog.Infof("inspect task is being deleted")
		err = os.Remove(path.Join(constant.ResultPathPrefix, result.Name))
		if err != nil {
			klog.Error(err, "failed to delete file")
		}
		err = r.Client.Update(ctx, result)
		if err != nil {
			klog.Error("Failed to inspect plan add finalizers. ", err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if result.Status.Complete {
		return ctrl.Result{}, nil
	}

	taskName := result.GetLabels()[constant.LabelTaskName]

	task, err := r.KubeEyeFactory.V1alpha2().InspectTasks().Lister().Get(taskName)
	if err != nil {
		klog.Error("Failed to get inspect task", err)
		return ctrl.Result{}, err
	}
	startTime := result.GetAnnotations()[constant.AnnotationStartTime]
	endTime := result.GetAnnotations()[constant.AnnotationEndTime]

	parseStart, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}
	parseEnd, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}

	result.Status.Policy = task.Spec.InspectPolicy
	result.Status.Duration = parseEnd.Sub(parseStart).String()
	result.Status.TaskStartTime = startTime
	result.Status.TaskEndTime = endTime
	result.Status.Complete = true
	countLevelNum, err := r.CountLevelNum(result.Name)
	if err != nil {
		klog.Error("Failed to count level num", err)
		return ctrl.Result{}, err
	}
	result.Status.Level = countLevelNum

	err = r.Client.Status().Update(ctx, result)
	if err != nil {
		klog.Error("Failed to update inspect result status", err)
		return ctrl.Result{}, err
	}

	go r.SendMessage(result)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InspectResultReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubeeyev1alpha2.InspectResult{}).
		Complete(r)
}

func (r *InspectResultReconciler) CountLevelNum(resultName string) (map[kubeeyev1alpha2.Level]*int, error) {
	file, err := os.ReadFile(path.Join(constant.ResultPathPrefix, resultName))
	if err != nil {
		return nil, err
	}

	var result kubeeyev1alpha2.InspectResult

	err = json.Unmarshal(file, &result)
	if err != nil {
		return nil, err
	}

	levelTotal := make(map[kubeeyev1alpha2.Level]*int)
	levelTotal[kubeeyev1alpha2.DangerLevel] = &result.Spec.OpaResult.Dangerous
	levelTotal[kubeeyev1alpha2.WarningLevel] = &result.Spec.OpaResult.Warning
	levelTotal[kubeeyev1alpha2.IgnoreLevel] = &result.Spec.OpaResult.Ignore
	totalResultLevel(result.Spec.FileChangeResult, levelTotal)

	totalResultLevel(result.Spec.FileFilterResult, levelTotal)

	totalResultLevel(result.Spec.SysctlResult, levelTotal)

	totalResultLevel(result.Spec.SystemdResult, levelTotal)

	totalResultLevel(result.Spec.NodeInfo, levelTotal)

	totalResultLevel(result.Spec.PrometheusResult, levelTotal)

	totalResultLevel(result.Spec.ServiceConnectResult, levelTotal)
	totalResultLevel(result.Spec.ComponentResult, levelTotal)

	totalResultLevel(result.Spec.CommandResult, levelTotal)

	return levelTotal, nil
}
func totalResultLevel(data interface{}, mapLevel map[kubeeyev1alpha2.Level]*int) {

	maps, err := utils.ArrayStructToArrayMap(data)
	if err != nil {
		return
	}
	Autoincrement := func(level kubeeyev1alpha2.Level) *int {
		if mapLevel[level] == nil {
			mapLevel[level] = new(int)
		}
		*mapLevel[level]++
		return mapLevel[level]
	}
	for _, m := range maps {
		_, exist := m["assert"]
		if exist {
			s, isOk := m["issues"]
			a := 1
			if isOk {
				a = len(s.([]interface{}))
			}
			v, ok := m["level"]
			for i := 0; i < a; i++ {
				if !ok {
					Autoincrement(kubeeyev1alpha2.DangerLevel)
				} else {
					Autoincrement(kubeeyev1alpha2.Level(v.(string)))
				}
			}

		}
	}
}

func (r *InspectResultReconciler) SendMessage(result *kubeeyev1alpha2.InspectResult) {

	kc, err := kube.GetKubeEyeConfig(r.K8sFactory.Core())
	if err != nil {
		klog.Error("GetKubeEyeConfig error", err)
		return
	}
	if kc.Message == nil || !kc.Message.Enable {
		return
	}

	n := GetIssuesNumber(result)
	if kc.Message.Mode == "" || kc.Message.Mode == conf.AbnormalMode {
		if n == 0 {
			return
		}
	}
	klog.Info("sending email")
	htmlTemplate, err := template.GetInspectResultHtmlTemplate()
	if err != nil {
		klog.Error("GetInspectResultHtmlTemplate error", err)
		return
	}
	err, m := output.HtmlOut(result.Name)
	if err != nil {
		klog.Error("get html render data error", err)
		return
	}
	data := bytes.NewBufferString("")
	err = htmlTemplate.Execute(data, m)
	if err != nil {
		klog.Error("render html template error", err)
		return
	}

	messageHandler := message.NewEmailMessageOptions(&kc.Message.Email, r.Client)
	dispatcher := message.RegisterHandler(messageHandler)
	dispatcher.DispatchMessageEvent(&conf.MessageEvent{
		Title:     fmt.Sprintf("%s集群巡检完成,共发现%d个问题", result.Spec.InspectCluster.Name, n),
		Timestamp: time.Now(),
		Content:   data.Bytes(),
	})
}

func GetIssuesNumber(result *kubeeyev1alpha2.InspectResult) (n int) {
	for _, l := range result.Status.Level {
		if l != nil {
			n += *l
		}
	}
	return n
}
