package expend

import (
	"bytes"
	"context"
	_ "embed"
)

func InstallNPD(ctx context.Context, kubeconfig string) error {

	installer := Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}
	NPDResource := bytes.Split(npdResources, []byte("---"))

	for _, resource := range NPDResource {
		if err := installer.Install(string(resource)); err != nil {
			return err
		}
	}
	return nil
}

func UninstallNPD(ctx context.Context, kubeconfig string) error {
	installer := Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}
	NPDResource := bytes.Split(npdResources, []byte("---"))

	for _, resource := range NPDResource {
		if err := installer.Uninstall(string(resource)); err != nil {
			return err
		}
	}
	return nil
}
