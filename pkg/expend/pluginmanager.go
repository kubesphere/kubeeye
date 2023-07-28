package expend

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed deploymentfiles/npd-resources.yaml
var npdResources []byte

////go:embed deploymentfiles/kubebench.yaml
////var kubebenchResources []byte
//
////go:embed deploymentfiles/kubehunter.yaml
//var kubehunterResources []byte
//
////go:embed deploymentfiles/kubescape.yaml
////var kubescapeResources []byte

func PluginsInstaller(ctx context.Context, pluginName string, pluginResources string) (err error) {
	installer := Installer{
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
	installer := Installer{
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
