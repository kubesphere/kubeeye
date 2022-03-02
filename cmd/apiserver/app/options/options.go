/*
Copyright 2020 KubeSphere Authors

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

package options

import (
	"embed"
	"fmt"
	"io/fs"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"os"
)
////go:embed dist/*
var assets embed.FS

// errorResponse
type errorResponder struct{}

func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
	klog.Error(err)
	responsewriters.InternalError(w, req, err)
}

// ServerRunOptions apiServer Options
type ServerRunOptions struct {
	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// tls cert file
	TlsCertFile string

	// tls private key file
	TlsPrivateKey string
}

// NewServerRunOptions creates server option
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  9088,
		SecurePort:    0,
		TlsCertFile:   "",
		TlsPrivateKey: "",
	}
	return &s
}

// loadConfig get cluster config
func loadConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	return clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
}

// NewAPIServer creates an APIServer instance using given options
func NewAPIServer(s *ServerRunOptions) error {

	subFS, _ := fs.Sub(assets, "dist")
	assetsFs := http.FileServer(http.FS(subFS))

	config, err := loadConfig()
	if err != nil {
		return err
	}

	kubernetes, _ := url.Parse(config.Host)
	defaultTransport, err := rest.TransportFor(config)
	if err != nil {
		return err
	}

	var handleKubeAPIfunc = func(w http.ResponseWriter, req *http.Request) {
		klog.Info(req.URL)
		s := *req.URL
		s.Host = kubernetes.Host
		s.Scheme = kubernetes.Scheme

		// make sure we don't override kubernetes's authorization
		req.Header.Del("Authorization")
		httpProxy := proxy.NewUpgradeAwareHandler(&s, defaultTransport, true, false, &errorResponder{})
		httpProxy.UpgradeTransport = proxy.NewUpgradeRequestRoundTripper(defaultTransport, defaultTransport)
		httpProxy.ServeHTTP(w, req)
		return
	}

	// create http server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", handleKubeAPIfunc)
	mux.HandleFunc("/apis/", handleKubeAPIfunc)
	mux.Handle("/", assetsFs)

	if s.TlsCertFile != "" && s.TlsPrivateKey != "" {
		klog.Infof("Start listening on %s", s.InsecurePort)
		err = http.ListenAndServeTLS(fmt.Sprintf(":%d", s.InsecurePort+1), s.TlsCertFile, s.TlsPrivateKey, mux)
	} else {
		klog.Infof("Start listening on %s", s.InsecurePort)
		err = http.ListenAndServe(fmt.Sprintf(":%d", s.InsecurePort), mux)

	}
	return err
}
