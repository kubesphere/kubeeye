package expend

import (
	"context"
)

type Expends interface {
	install(resource Resources) error
	uninstall(resource Resources) error
}

type Installer struct {
	CTX        context.Context
	Kubeconfig string
}

type Resources []byte

func (installer Installer) install(resource Resources) error {
	ctx := installer.CTX
	kubeconfig := installer.Kubeconfig

	// create npd resources
	err := CreateResource(kubeconfig, ctx, resource)
	if err != nil {
		return err
	}
	return nil
}

func (installer Installer) uninstall(resource Resources) error {
	ctx := installer.CTX
	kubeconfig := installer.Kubeconfig

	// delete npd resources
	err := RemoveResource(kubeconfig, ctx, resource)
	if err != nil {
		return err
	}
	return nil
}
