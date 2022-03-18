package expend

import (
	"bytes"
	"context"
	_ "embed"
)

//go:embed plugins/npd-resources.yaml
var npdResources []byte

func InstallNPD(ctx context.Context, kubeconfig string) error {
	var installer Expends
	installer = Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}
	NPDResource := bytes.Split(npdResources, []byte("---"))

	for _, resource := range NPDResource {
		if err := installer.install(resource); err != nil {
			return err
		}
	}
	return nil
}

func UninstallNPD(ctx context.Context, kubeconfig string) error {
	var installer Expends
	installer = Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}
	NPDResource := bytes.Split(npdResources, []byte("---"))

	for _, resource := range NPDResource {
		if err := installer.uninstall(resource); err != nil {
			return err
		}
	}
	return nil
}
