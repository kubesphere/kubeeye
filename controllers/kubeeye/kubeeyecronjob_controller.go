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

package kubeeye

import (
	"context"
	"time"

	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	kubeeyeclientset "github.com/kubesphere/kubeeye/client/clientset/versioned"
	kubeeyeinformer "github.com/kubesphere/kubeeye/client/informers/externalversions/kubeeye/v1alpha1"
	kubeeyev1alpha1listers "github.com/kubesphere/kubeeye/client/listers/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/kubeeye"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=kubeeyecronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=kubeeyecronjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeeye.kubesphere.io,resources=kubeeyecronjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=*

var (
	nextScheduleDelta = 100 * time.Millisecond
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "KubeEye synced successfully"
	controllerName        = "KubeEye-controller"
)

type KubeeyeCronjobController struct {
	kubeeye.BaseController
	k8sClient     kubernetes.Interface
	kubeeyeclient kubeeyeclientset.Interface
	kubeeyeLister kubeeyev1alpha1listers.ClusterInsightLister
	kubeeyeSynced cache.InformerSynced
	// now is a function that returns current time, done to facilitate unit tests
	now func() time.Time
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func (k *KubeeyeCronjobController) Start(ctx context.Context) error {
	return k.Run(5, ctx.Done())
}

func (k *KubeeyeCronjobController) reconcile(key string) error {
	clusterInsight, err := k.kubeeyeLister.Get(key)
	if err != nil {
		return err
	}

	if !clusterInsight.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted, the reconciler can do nothing.
		return nil
	}

	now := k.now()
	sched, err := cron.ParseStandard(clusterInsight.Spec.AuditPeriod)
	if err != nil {
		klog.V(2).InfoS("Unparseable schedule", "kubeeyecronjob", klog.KRef(clusterInsight.GetNamespace(), clusterInsight.GetName()), "schedule", clusterInsight.Spec.AuditPeriod, "err", err)
		return err
	}

	scheduledTime := nextScheduledTimeDuration(sched, now)
	klog.Infof("Next audit time: %v", now.Add(*scheduledTime))

	if clusterInsight.Status.LastScheduleTime.IsZero() && clusterInsight.Status.AuditPercent == 0 && clusterInsight.Status.PluginsResults == nil {
		klog.Infof("no LastScheduleTime")
		k.Workqueue.AddAfter(key, 10*time.Second)
	} else if clusterInsight.Status.LastScheduleTime.IsZero() && (clusterInsight.Status.AuditPercent > 0 && clusterInsight.Status.AuditPercent < 100) {
		klog.Infof("wait for KubeEye audit finish")
		k.Workqueue.AddAfter(key, 10*time.Second)
	} else if clusterInsight.Status.LastScheduleTime.IsZero() && clusterInsight.Status.AuditPercent == 0 && clusterInsight.Status.PluginsResults != nil {
		klog.Infof("clear plugins results in clusterInsight, Don't worry, the plugins will be re-executed")
		clusterInsight = kubeeye.ClearClusterInsightStatus(clusterInsight)
		if err := k.updateKubeeyeStatus(clusterInsight, now); err != nil {
			klog.V(2).InfoS("update clusterInsight failed", "kubeeyecronjob", klog.KRef(clusterInsight.GetNamespace(), clusterInsight.GetName()), "schedule", clusterInsight.Spec.AuditPeriod, "err", err)
			k.recorder.Event(clusterInsight, corev1.EventTypeWarning, "KubeEyeCronjob failed", "KubeEyeCronjob failed to update clusterInsight")
			return err
		}
	} else {
		if clusterInsight.Status.LastScheduleTime.Add(*scheduledTime).Before(now) {
			clusterInsight = kubeeye.ClearClusterInsightStatus(clusterInsight)
			if err := k.updateKubeeyeStatus(clusterInsight, now); err != nil {
				klog.V(2).InfoS("update clusterInsight failed", "kubeeyecronjob", klog.KRef(clusterInsight.GetNamespace(), clusterInsight.GetName()), "schedule", clusterInsight.Spec.AuditPeriod, "err", err)
				k.recorder.Event(clusterInsight, corev1.EventTypeWarning, "KubeEyeCronjob failed", "KubeEyeCronjob failed to update clusterInsight")
				return err
			}
		} else {
			k.Workqueue.AddAfter(key, clusterInsight.Status.LastScheduleTime.Add(*scheduledTime).Sub(now))
		}
	}

	klog.Info("sync KubeEyeCronjob successful")
	return nil
}

func (k *KubeeyeCronjobController) updateKubeeyeStatus(clusterInsight *v1alpha1.ClusterInsight, now time.Time) error {
	updateOptions := metav1.UpdateOptions{}
	_, err := k.kubeeyeclient.KubeeyeV1alpha1().ClusterInsights().UpdateStatus(context.Background(), clusterInsight, updateOptions)
	if err != nil {
		klog.V(2).InfoS("clear clusterInsight failed", "kubeeyecronjob", klog.KRef(clusterInsight.GetNamespace(), clusterInsight.GetName()), "schedule", clusterInsight.Spec.AuditPeriod, "err", err)
		return err
	}
	klog.Infof("clear clusterInsight successful")
	return nil
}

// nextScheduledTimeDuration returns the time duration to requeue based on
// the schedule and current time. It adds a 100ms padding to the next requeue to account
// for Network Time Protocol(NTP) time skews. If the time drifts are adjusted which in most
// realistic cases would be around 100s, scheduled cron will still be executed without missing
// the schedule.
func nextScheduledTimeDuration(sched cron.Schedule, now time.Time) *time.Duration {
	t := sched.Next(now).Add(nextScheduleDelta).Sub(now)
	return &t
}

func AddKubeeyeController(mgr manager.Manager, clients *kube.KubernetesClient, kubeeyeclient *kubeeyeclientset.Clientset,
	informerFactory kubeeye.InformerFactory) {
	kubeeyeInformer := informerFactory.KubeeyeSharedInformerFactory()

	kubeeyeCronjobController := NewKubeeyeCronjobController(mgr, clients, kubeeyeclient, kubeeyeInformer.Kubeeye().V1alpha1().ClusterInsights())
	addController(mgr, "kubeeyecontroller", kubeeyeCronjobController)
}

func NewKubeeyeCronjobController(mgr manager.Manager, clients *kube.KubernetesClient, kubeeyeclient kubeeyeclientset.Interface, kubeeyeInformer kubeeyeinformer.ClusterInsightInformer) *KubeeyeCronjobController {

	k := &KubeeyeCronjobController{
		BaseController: kubeeye.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
			Synced:    []cache.InformerSynced{kubeeyeInformer.Informer().HasSynced},
			Name:      controllerName,
		},
		k8sClient:     clients.ClientSet,
		kubeeyeclient: kubeeyeclient,
		kubeeyeLister: kubeeyeInformer.Lister(),
		now:           time.Now,
		recorder:      mgr.GetEventRecorderFor("kubeeyecronjob-controller"),
	}
	k.Handler = k.reconcile
	klog.Info("Setting up event handlers")
	kubeeyeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: k.Enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			k.Enqueue(newObj)
		},
		DeleteFunc: k.Enqueue,
	})
	return k
}

var addSuccessfullyControllers = sets.NewString()

func addController(mgr manager.Manager, name string, controller manager.Runnable) {
	if err := mgr.Add(controller); err != nil {
		klog.Fatalf("Unable to create %v controller: %v", name, err)
	}
	addSuccessfullyControllers.Insert(name)
}
