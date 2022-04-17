package pkg

// Install plugin param
var (
	PluginResource    = "plugins"
	PluginConfig      = "kubeeye-manager-config"
	PrefixManifestKey = "pluginmanifest-"
	// plugin gvr
	Group    = "kubeeye.kubesphere.io"
	Version  = "v1alpha1"
	Resource = "pluginsubscriptions"
	// plugin manager
	MaxConcurrentReconciles = 3
	MaxCheckPodCount        = 15
	IntervalsTime           = 20
)

//plugin install status
const (
	PluginIntalled     string = "installed"
	PluginInstalling   string = "installing"
	PluginUninstalled  string = "uninstalled"
	PluginUninstalling string = "uninstalling"
)
