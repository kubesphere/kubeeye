/*
Copyright 2019 The KubeSphere Authors.

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

package app

import (
	"context"
	"github.com/kubesphere/kubeeye/cmd/apiserver/app/options"
	"github.com/kubesphere/kubeeye/pkg/utils/signal"
	"k8s.io/klog"
)

func NewAPIServerCommand(){
	s := options.NewServerRunOptions()
	err := Run(s, signal.SetupSignalHandler())
	if err != nil {
		klog.Fatal("apiServer Run failed ", err)
		return
	}
}

func Run(s *options.ServerRunOptions, ctx context.Context) error {
	apiServer, err := s.NewAPIServer()
	if err != nil {
		return err
	}

	err = apiServer.PrepareRun()
	if err != nil {
		return err
	}
	return apiServer.Run(ctx)
}
