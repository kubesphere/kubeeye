package expend

import (
	"context"
)

type Expends interface {
	install()
	uninstall()
}

type Installer struct {
	ctx        context.Context
	kubeconfig string
}

type Resources []byte

func (installer Installer) install(resource Resources) {
	ctx := installer.ctx
	kubeconfig := installer.kubeconfig

	// create npd resources
	err := CreateResource(kubeconfig, ctx, resource)
	if err != nil {
		panic(err)
	}

}

func (installer Installer) uninstall(resource Resources) {
	ctx := installer.ctx
	kubeconfig := installer.kubeconfig

	// delete npd resources
	err := RemoveResource(kubeconfig, ctx, resource)
	if err != nil {
		panic(err)
	}
}
