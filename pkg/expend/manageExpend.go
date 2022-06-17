package expend

import (
	"context"
)

type Expends interface {
	Install(resource string) error
	Uninstall(resource string) error
}

type Installer struct {
	CTX context.Context
	// Kubeconfig can be deleted later
	Kubeconfig string
}

func (installer Installer) Install(resource string) error {
	// create  resources
	err := ResourceCreater(installer, resource)
	if err == nil {
		return err
	}
	return nil
}

func (installer Installer) Uninstall(resource string) error {
	// delete  resources
	err := ResourceRemover(installer, resource)
	if err != nil {
		return err
	}
	return nil
}
