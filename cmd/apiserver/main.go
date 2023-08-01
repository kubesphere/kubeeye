package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/router"
	_ "github.com/kubesphere/kubeeye/swaggerDocs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

// @title           KubeEye API
// @version         1.0
// @description     This is a kubeeye api server.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /kapis/kubeeye.kubesphere.io/v1alpha1

func main() {

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
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

	router.RegisterRouter(ctx, r, clients)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	go func() {
		// 服务连接
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
