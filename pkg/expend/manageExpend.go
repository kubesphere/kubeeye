package expend

import (
	"context"
)

type Expends interface {
	Install(resource []byte) error
	Uninstall(resource []byte) error
}

type Installer struct {
	CTX context.Context
	// Kubeconfig can be deleted later
	Kubeconfig string
}

func (installer Installer) Install(resource []byte) error {
	// create  resources
	err := ResourceCreater(installer, resource)
	if err == nil {
		return err
	}
	return nil
}

func (installer Installer) Uninstall(resource []byte) error {
	// delete  resources
	err := ResourceRemover(installer, resource)
	if err != nil {
		return err
	}
	return nil
}
