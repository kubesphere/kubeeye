package main

import (
	"context"
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
		_, err = factory.ForResource(schema.GroupVersionResource{
			Group:    kubeeyev1alpha2.GroupVersion.Group,
			Version:  kubeeyev1alpha2.GroupVersion.Version,
			Resource: resource,
		})
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	router.RegisterRouter(ctx, r, clients, factory.Kubeeye())

	srv := &http.Server{
		Addr:    ":9090",
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
