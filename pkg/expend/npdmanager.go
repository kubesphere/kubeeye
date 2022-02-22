package expend

import (
	"context"
	_ "embed"
)

//go:embed deploymentfiles/npd-daemonSet.yaml
var npdDaemonset []byte

//go:embed deploymentfiles/npd-configmap.yaml
var npdConfigmap []byte

type NPD struct {
	ctx        context.Context
	kubeconfig string
}

func (npd NPD) install() {
	ctx := npd.ctx
	kubeconfig := npd.kubeconfig

	// create npd configmap
	err := CreateResource(kubeconfig, ctx, npdConfigmap)
	if err != nil {
		panic(err)
	}
	// create npd daemonset
	err = CreateResource(kubeconfig, ctx, npdDaemonset)
	if err != nil {
		panic(err)
	}

}

func (npd NPD) uninstall() {
	ctx := npd.ctx
	kubeconfig := npd.kubeconfig

	// delete npd configmap
	err := RemoveResource(kubeconfig, ctx, npdConfigmap)
	if err != nil {
		panic(err)
	}
	// delete npd daemonset
	err = RemoveResource(kubeconfig, ctx, npdDaemonset)
	if err != nil {
		panic(err)
	}
}
