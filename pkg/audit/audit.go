// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Audit struct {
	sync.Mutex
	TaskQueue   *TaskQueue
	TaskResults map[string]map[string]*kubeeyev1alpha2.AuditResult
	K8sClient   *kube.KubernetesClient
	Cli         client.Client
}

func (k *Audit) StartAudit(ctx context.Context) {
	for {
		obj, shutdown := k.TaskQueue.Get()
		if shutdown {
			return
		}
		taskName, ok := obj.(types.NamespacedName)
		if !ok {
			k.TaskQueue.Done(obj)
			continue
		}

		go k.TriggerAudit(ctx, taskName)

	}
}

func (k *Audit) TriggerAudit(ctx context.Context, taskName types.NamespacedName) {
	auditTask := &kubeeyev1alpha2.AuditTask{
		ObjectMeta: metav1.ObjectMeta{Name: taskName.Name, Namespace: taskName.Namespace},
	}
	err := k.Cli.Get(ctx, client.ObjectKeyFromObject(auditTask), auditTask)
	if err != nil {
		klog.Error(err, "failed to get audit task")
		return
	}
	timeout, err := time.ParseDuration(auditTask.Spec.Timeout)
	if err != nil {
		klog.Error(err, "failed to parse timeout")
		timeout = constant.DefaultTimeout
	}
	k.Mutex.Lock()
	k.TaskResults[taskName.Name] = make(map[string]*kubeeyev1alpha2.AuditResult, len(auditTask.Spec.Auditors))
	k.Mutex.Unlock()

	auditorSvcMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: constant.AuditorServiceAddrConfigMap, Namespace: taskName.Namespace},
	}
	err = k.Cli.Get(ctx, client.ObjectKeyFromObject(auditorSvcMap), auditorSvcMap)
	if err != nil {
		klog.Error(err, " failed to get audit service configmap")
		return
	}

	auditorMap := auditorSvcMap.Data
	for _, auditor := range auditTask.Spec.Auditors {
		if auditor == "kubeeye" {
			go k.KubeeyeAudit(taskName.Name, ctx)
		} else {
			go k.PluginAudit(ctx, taskName.Name, string(auditor), auditorMap)
		}
	}

	done := func() (bool, error) {
		err := k.Cli.Get(ctx, client.ObjectKeyFromObject(auditTask), auditTask)
		if err != nil {
			klog.Error(err, "failed to get audit task")
			return false, nil
		}
		if auditTask.Status.Phase == kubeeyev1alpha2.PhaseSucceeded {
			k.TaskQueue.Done(taskName)
			return true, nil
		}
		return false, nil
	}
	err = wait.PollImmediate(3*time.Second, timeout, done)
	if err != nil {
		k.TaskQueue.Done(taskName)
		return
	}
}

func (k *Audit) KubeeyeAudit(taskName string, ctx context.Context) {
	klog.Infof("%s : start kubeeye audit", taskName)
	auditResult := &kubeeyev1alpha2.AuditResult{Name: "kubeeye", Phase: kubeeyev1alpha2.PhaseRunning}

	k.Mutex.Lock()
	k.TaskResults[taskName]["kubeeye"] = auditResult
	k.Mutex.Unlock()
	// start kubeeye audit
	K8SResources, validationResultsChan, Percent := ValidationResults(ctx, k.K8sClient, "")

	kubeeyeResult := kubeeyev1alpha2.KubeeyeAuditResult{}
	var results []kubeeyev1alpha2.ResultItems
	ctxCancel, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				kubeeyeResult.Percent = Percent.AuditPercent // update kubeeye audit percent
				ext := runtime.RawExtension{}
				marshal, err := json.Marshal(kubeeyeResult)
				if err != nil {
					klog.Error(err, " failed marshal kubeeye result")
					return
				}
				ext.Raw = marshal
				auditResult.Result = ext
			case <-ctx.Done():
				return
			}
		}
	}(ctxCancel)

	for r := range validationResultsChan {
		for _, result := range r {
			results = append(results, result)
		}
	}

	cancel()
	scoreInfo := CalculateScore(results, K8SResources)
	kubeeyeResult.Percent = 100
	kubeeyeResult.ScoreInfo = scoreInfo
	kubeeyeResult.ExtraInfo = kubeeyev1alpha2.ExtraInfo{
		WorkloadsCount: K8SResources.WorkloadsCount,
		NamespacesList: K8SResources.NameSpacesList,
	}
	kubeeyeResult.ResultItem = results

	ext := runtime.RawExtension{}
	marshal, err := json.Marshal(kubeeyeResult)
	if err != nil {
		klog.Error(err, " failed marshal kubeeye result")
		return
	}
	ext.Raw = marshal
	auditResult.Result = ext

	auditResult.Phase = kubeeyev1alpha2.PhaseSucceeded
	klog.Infof("%s : finish kubeeye audit", taskName)
}

func (k *Audit) PluginsResultsReceiver(pluginsResultsReceiverAddr string) {

	ServeMux := http.NewServeMux()
	ServeMux.Handle("/plugins", http.HandlerFunc(k.PluginsResult))
	err := http.ListenAndServe(pluginsResultsReceiverAddr, ServeMux)
	if err != nil {
		klog.Error("failed to listen plugin audit result")
		return
	}
}

// receive extra plugin result
func (k *Audit) PluginsResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	taskName := r.PostFormValue("taskname")
	pluginName := r.PostFormValue("pluginname")
	pluginResult := r.PostFormValue("pluginresult")
	ext := runtime.RawExtension{}

	ext.Raw = []byte(pluginResult)
	result := &kubeeyev1alpha2.AuditResult{
		Result: ext,
		Name:   pluginName,
		Phase:  kubeeyev1alpha2.PhaseSucceeded,
	}
	k.Mutex.Lock()
	k.TaskResults[taskName][pluginName] = result
	k.Mutex.Unlock()
	klog.Infof(" task %s receive %s plugin result", taskName, pluginName)
	w.WriteHeader(http.StatusOK)
}

func (k *Audit) PluginAudit(ctx context.Context, taskName string, pluginName string, auditorMap map[string]string) {
	service, ok := auditorMap[pluginName]
	if !ok {
		service = fmt.Sprintf("%s.%s.svc", pluginName, constant.DefaultNamespace)
	}

	url := fmt.Sprintf("http://%s/start?taskname=%s&kubeeyesvc=%s", service, taskName, auditorMap["kubeeye"])
	_, err := http.Get(url)
	if err != nil {
		klog.Error(err, "%s : failed to request %s audit", taskName, pluginName)
	}
}
