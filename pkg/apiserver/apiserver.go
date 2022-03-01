package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"

	"github.com/kubesphere/kubeeye/pkg/apiserver/filters"
	"github.com/kubesphere/kubeeye/pkg/apiserver/request"
	"github.com/kubesphere/kubeeye/pkg/utils/iputil"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"net/http"
	rt "runtime"
	"time"
)

type APIServer struct {
	Server    *http.Server
	container *restful.Container
}

func (s *APIServer) PrepareRun() error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	kubeConfig, err := kube.GetKubeConfig("")
	if err != nil {
		return err
	}
	s.buildHandlerChain(kubeConfig)
	return nil
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
		iputil.RemoteIp(req.Request),
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

func (s *APIServer) buildHandlerChain(conf *rest.Config) {
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api", "apis"),
		GrouplessAPIPrefixes: sets.NewString("api"),
	}

	handler := s.Server.Handler
	handler = filters.WithKubeAPIServer(handler, conf, &errorResponder{})
	handler = filters.WithRequestInfo(handler, requestInfoResolver)
	s.Server.Handler = handler
}

type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}
