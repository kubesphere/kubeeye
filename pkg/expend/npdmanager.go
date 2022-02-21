package expend

import (
	"bytes"
	"context"
	_ "embed"
)

//go:embed deploymentfiles/npd-resources.yaml
var npdResources []byte

func InstallNPD(ctx context.Context, kubeconfig string) {
	NPDResource := bytes.Split(npdResources, []byte("---"))
	for _, resources := range NPDResource {
		CreateResource(kubeconfig, ctx, resources)
	}
}

func UninstallNPD(ctx context.Context, kubeconfig string) {
	NPDResource := bytes.Split(npdResources, []byte("---"))
	for _, resources := range NPDResource {
		RemoveResource(kubeconfig, ctx, resources)
	}
}
