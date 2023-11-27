/*
 Copyright 2022 The KubeSphere Authors.
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

package main

import (
	"flag"
	controllers2 "github.com/kubesphere/kubeeye/pkg/controllers"
	"github.com/kubesphere/kubeeye/pkg/informers"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"go.uber.org/zap/zapcore"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kubeeyev1alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var pluginsResultsReceiverAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&pluginsResultsReceiverAddr, "plugins-results-receiver-address", ":8888", "The address the plugin result receiver binds to")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "fa68b2a3.kubesphere.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	// get kubernetes cluster config
	kubeConfig, err := kube.GetKubeConfigInCluster()
	if err != nil {
		setupLog.Error(err, "Failed to load cluster clients")
		os.Exit(1)
	}

	// get kubernetes cluster clients
	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		setupLog.Error(err, "Failed to load cluster clients")
		os.Exit(1)
	}

	setupLog.Info("starting inspect")
	factory := informers.NewInformerFactory(clients.ClientSet, clients.VersionClientSet)

	if err = (&controllers2.InspectPlanReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		K8sClient:      clients,
		KubeEyeFactory: factory.KubeEyeInformerFactory().Kubeeye(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InspectPlan")
		os.Exit(1)
	}
	if err = (&controllers2.InspectTaskReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		K8sClients:     clients,
		KubeEyeFactory: factory.KubeEyeInformerFactory().Kubeeye(),
		K8sFactory:     factory.KubernetesInformerFactory(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InspectTask")
		os.Exit(1)
	}
	if err = (&controllers2.InspectResultReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		KubeEyeFactory: factory.KubeEyeInformerFactory().Kubeeye(),
		K8sFactory:     factory.KubernetesInformerFactory(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InspectResult")
		os.Exit(1)
	}
	if err = (&controllers2.InspectRulesReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "InspectRule")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	factory.ForResources(informers.KeEyeGver(), informers.K8sEyeGver())
	factory.Start(ctx.Done())

	factory.KubeEyeInformerFactory().WaitForCacheSync(ctx.Done())
	factory.KubernetesInformerFactory().WaitForCacheSync(ctx.Done())

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
