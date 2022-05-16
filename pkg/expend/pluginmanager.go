package expend

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net/http"
	
	kubeeyepluginsv1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeyeplugins/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/pkg/errors"
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
	_, err := http.Get(fmt.Sprintf("http://%s.%s.svc/healthz",plugin.Name, conf.KubeeyeNameSpace))
	if err != nil {
		return "", err
	}
	return conf.PluginIntalled, nil
}

func PluginsInstaller(ctx context.Context, pluginName string) (err error) {
	var installer Expends
	installer = Installer{
		CTX:        ctx,
	}
	var pluginsResources []byte

	if pluginName == "npd" || pluginName == "NPD"{
		pluginsResources = npdResources
	} else if pluginName == "kubebench" {
		pluginsResources = kubebenchResources
	} else if pluginName == "kubehunter" {
		pluginsResources = kubehunterResources
	} else if pluginName == "kubescape" {
		pluginsResources = kubescapeResources
	} else {
		return errors.Wrap(err, "Unknown plugin name")
	}
	pluginsResource := bytes.Split(pluginsResources, []byte("---"))
	
	for _, resource := range pluginsResource {
		if err := installer.Install(resource); err != nil {
			return err
		}
	}
	return nil
}

func PluginsUninstaller(ctx context.Context, pluginName string) (err error) {
	var installer Expends
	var pluginsResources []byte
	installer = Installer{
		CTX:        ctx,
	}
	if pluginName == "npd" || pluginName == "NPD"{
		pluginsResources = npdResources
	} else if pluginName == "kubebench" {
		pluginsResources = kubebenchResources
	} else if pluginName == "kubehunter" {
		pluginsResources = kubehunterResources
	} else if pluginName == "kubescape" {
		pluginsResources = kubescapeResources
	} else {
		return errors.Wrap(err, "Unknown plugin name")
	}
	pluginsResource := bytes.Split(pluginsResources, []byte("---"))
	
	for _, resource := range pluginsResource {
		if err := installer.Uninstall(resource); err != nil {
			return err
		}
	}
	return nil
}
