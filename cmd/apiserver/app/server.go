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
	"k8s.io/klog"
)

func NewAPIServerCommand(){
	s := options.NewServerRunOptions()
	Run(s, context.Background())
}

func Run(s *options.ServerRunOptions, ctx context.Context) error {

	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		klog.Error("Failed to NewAPIServer %v", err)
		return err
	}

	err = apiserver.PrepareRun(ctx.Done())
	if err != nil {
		klog.Error("PrepareRun err %v",err)
		return err
	}
	klog.Infof("Run prep!!!!!!!!!!!!!")
	return apiserver.Run(ctx)
}
