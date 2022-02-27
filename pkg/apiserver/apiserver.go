package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/kubesphere/kubeeye/pkg/apiserver/v1"
	"github.com/kubesphere/kubeeye/pkg/kube"
	//"kubesphere.io/kubesphere/pkg/simple/client/k8s"

	//"k8s.io/apimachinery/pkg/runtime/schema"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"github.com/kubesphere/kubeeye/pkg/apiserver/request"
	"github.com/kubesphere/kubeeye/pkg/apiserver/filters"
	"k8s.io/apimachinery/pkg/util/sets"
	//urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	//"kubesphere.io/kubesphere/pkg/utils/iputil"

	//"kubesphere.io/kubesphere/pkg/utils/iputil"
	"k8s.io/klog"
	"net/http"
	rt "runtime"
	"time"
)

type APIServer struct {
	Server    *http.Server
	container *restful.Container
	//KubernetesClient k8s.Client
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.installKubeSphereAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container
	s.buildHandlerChain(stopCh)
	klog.Infof("PrepareRun success")
	return nil
}

func (s *APIServer) installKubeSphereAPIs() {
	urlruntime.Must(v1.AddToContainer(s.container))
}

func (s *APIServer) Run(ctx context.Context) (err error) {
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(shutdownCtx)
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		//iputil.RemoteIp(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func (s *APIServer) buildHandlerChain(stopCh <-chan struct{}) {
	klog.Infof("!!!!!!!!!!!!!!!!! buildHandlerChain coming")
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
		GrouplessAPIPrefixes: sets.NewString("api", "kapi"),
	}

	handler := s.Server.Handler
	kubeConfig, err := kube.GetKubeConfig("")
	klog.Infof("!!!!!!!!!!!!!!!!! kubeConfig :%v",kubeConfig)
	if err != nil {
		klog.Infof("!!!!!!!!!!!!!!!!! GetKubeConfig err:%+v",err)
	}

	handler = filters.WithKubeAPIServer(handler, kubeConfig, nil)
	klog.Infof("!!!!!!!!!!!!!!!!! filters.WithAuthentication.handler:%+v",handler)
	handler = filters.WithRequestInfo(handler, requestInfoResolver)
	s.Server.Handler = handler
}