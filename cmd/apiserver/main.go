package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/clients/informers/externalversions"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/router"
	_ "github.com/kubesphere/kubeeye/swaggerDocs"
	"github.com/pkg/errors"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"time"
)

// @title           KubeEye API
// @version         1.0
// @description     This is a kubeeye api server.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      172.31.73.216:30882
// @BasePath  /kapis/kubeeye.kubesphere.io/v1alpha2

func main() {

	r := gin.Default()

	r.GET("/readyz", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	ctx, cancelFunc := context.WithCancel(context.TODO())
	errCh := make(chan error)
	defer close(errCh)

	var kc kube.KubernetesClient
	kubeConfig, err := kube.GetKubeConfigInCluster()
	if err != nil {
		klog.Error(err)
		errCh <- err
	}

	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		klog.Error(err)
		errCh <- err
	}
	factory := externalversions.NewSharedInformerFactory(clients.VersionClientSet, 5*time.Second)

	forResources := []string{"inspectrules", "inspectresults", "inspectplans", "inspecttasks"}

	for _, resource := range forResources {
		f, _ := factory.ForResource(schema.GroupVersionResource{
			Group:    kubeeyev1alpha2.GroupVersion.Group,
			Version:  kubeeyev1alpha2.GroupVersion.Version,
			Resource: resource,
		})
		f.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				name := GetResourcesName(obj)
				fmt.Println(fmt.Sprintf("add cr,name:%s", name))
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				name := GetResourcesName(oldObj)
				fmt.Println(fmt.Sprintf("update cr,name:%s", name))
			},
			DeleteFunc: func(obj interface{}) {
				name := GetResourcesName(obj)
				fmt.Println(fmt.Sprintf("delete cr,name:%s", name))
			},
		})
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	router.RegisterRouter(ctx, r, clients, factory.Kubeeye())

	srv := &http.Server{
		Addr:    "0.0.0.0:9090",
		Handler: r,
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	go func() {
		// 服务连接
		if err = srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			klog.Errorf("listen: %s\n", err)
			errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			klog.Info("结束咯！！！")
			os.Exit(1)
		case errCtx := <-errCh:
			cancelFunc()
			klog.Infof("哦何，出错了！！！ err:%s", errCtx)
			os.Exit(1)
		}
	}
}

func GetResourcesName(obj interface{}) string {
	switch obj.(type) {
	case *kubeeyev1alpha2.InspectPlan:
		plan := obj.(*kubeeyev1alpha2.InspectPlan)
		return plan.Name
	case *kubeeyev1alpha2.InspectResult:
		result := obj.(*kubeeyev1alpha2.InspectResult)
		return result.Name
	case *kubeeyev1alpha2.InspectTask:
		task := obj.(*kubeeyev1alpha2.InspectTask)
		return task.Name
	case *kubeeyev1alpha2.InspectRule:
		rule := obj.(*kubeeyev1alpha2.InspectRule)
		return rule.Name
	}
	return ""
}
