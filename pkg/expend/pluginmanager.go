package expend

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	kubeeyepluginsv1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeyeplugins/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/conf"
)

//go:embed deploymentfiles/npd-resources.yaml
var npdResources []byte

//go:embed deploymentfiles/kubebench.yaml
var kubebenchResources []byte

//go:embed deploymentfiles/kubehunter.yaml
var kubehunterResources []byte

//go:embed deploymentfiles/kubescape.yaml
var kubescapeResources []byte

func PluginHealth(plugin *kubeeyepluginsv1alpha1.PluginSubscription) (string, error) {
	_, err := http.Get(fmt.Sprintf("http://%s.%s.svc/healthz", plugin.Name, conf.KubeeyeNameSpace))
	if err != nil {
		return "", err
	}
	return conf.PluginInstalled, nil
}

func PluginsInstaller(ctx context.Context, pluginName string, pluginResources string) (err error) {
	var installer Expends
	installer = Installer{
		CTX: ctx,
	}
	pluginsResource := strings.Split(pluginResources, "---")

	for _, resource := range pluginsResource {
		if err := installer.Install(resource); err != nil {
			return err
		}
	}
	return nil
}

func PluginsUninstaller(ctx context.Context, pluginName string, pluginResources string) (err error) {
	var installer Expends
	installer = Installer{
		CTX: ctx,
	}
	pluginsResource := strings.Split(pluginResources, "---")

	for _, resource := range pluginsResource {
		if err := installer.Uninstall(resource); err != nil {
			return err
		}
	}
	return nil
}
