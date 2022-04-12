package expend

import (
	"context"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
)

type Expends interface {
	install(resource Resources) error
	uninstall(resource Resources) error
}

type Installer struct {
	CTX context.Context
	// Kubeconfig can be deleted later
	Kubeconfig string
	Clients    *kube.KubernetesClient
}

type Resources []byte

func (installer Installer) install(resource Resources) error {
	ctx := installer.CTX
	// create  resources
	err := ResourceCreater(installer.Clients, ctx, resource)
	if err == nil {
		return err
	}
	return nil
}

func (installer Installer) uninstall(resource Resources) error {
	ctx := installer.CTX
	// delete  resources
	err := ResourceRemover(installer.Clients, ctx, resource)
	if err != nil {
		return err
	}
	return nil
}

func GetK8SClients(kubeconfig string) (*kube.KubernetesClient, error) {
	kubeConfig, err := kube.GetKubeConfig(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config file")
	}

	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		return nil, err
	}
	return clients, nil
}
