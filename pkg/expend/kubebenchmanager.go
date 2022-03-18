package expend

import (
	"bytes"
	"context"
	_ "embed"
	"io/ioutil"
	"net/http"
)

func InstallKubeBench(ctx context.Context, kubeconfig string) error {
	var installer Expends
	installer = Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}

	KubeBenchCRDUrl := "https://raw.githubusercontent.com/kubesphere/kubeeye/main/plugins/kube-bench/deploy/kubeeye-plugins-kubebench.yaml"
	KubeBenchResourceUrl := "https://raw.githubusercontent.com/kubesphere/kubeeye/main/plugins/kube-bench/config/samples/kubeeye_v1alpha1_kubebench.yaml"

	KubeBenchCRD, err := http.Get(KubeBenchCRDUrl)
	if err != nil {
		return err
	}
	defer KubeBenchCRD.Body.Close()

	KubeBenchConfigResource, err := http.Get(KubeBenchResourceUrl)
	if err != nil {
		return err
	}
	defer KubeBenchConfigResource.Body.Close()

	KubeBenchCRDResources, err := ioutil.ReadAll(KubeBenchCRD.Body)
	KubeBenchCRDResource := bytes.Split(KubeBenchCRDResources, []byte("---"))
	for _, resource := range KubeBenchCRDResource {
		if err := installer.install(resource); err != nil {
			return err
		}
	}

	KubeBenchResources, err := ioutil.ReadAll(KubeBenchConfigResource.Body)
	KubeBenchResource := bytes.Split(KubeBenchResources, []byte("---"))
	for _, resource := range KubeBenchResource {
		if err := installer.install(resource); err != nil {
			return err
		}
	}

	return nil
}

func UninstallKubeBench(ctx context.Context, kubeconfig string) error {
	var installer Expends
	installer = Installer{
		CTX:        ctx,
		Kubeconfig: kubeconfig,
	}

	KubeBenchCRDUrl := "https://raw.githubusercontent.com/ruiyaoOps/kubeeye-plugins/main/kube-bench/deploy/kubeeye-plugins-kubebench.yaml"
	KubeBenchCRD, err := http.Get(KubeBenchCRDUrl)
	if err != nil {
		return err
	}
	defer KubeBenchCRD.Body.Close()

	KubeBenchCRDResources, err := ioutil.ReadAll(KubeBenchCRD.Body)
	KubeBenchCRDResource := bytes.Split(KubeBenchCRDResources, []byte("---"))
	for _, resource := range KubeBenchCRDResource {
		if err := installer.uninstall(resource); err != nil {
			return err
		}
	}

	return nil
}