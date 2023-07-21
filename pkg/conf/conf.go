package conf

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	AppsGroup                = "apps"
	NoGroup                  = ""
	BatchGroup               = "batch"
	RoleGroup                = "rbac.authorization.k8s.io"
	APIVersionV1             = "v1"
	Nodes                    = "nodes"
	Deployments              = "deployments"
	Pods                     = "pods"
	Daemonsets               = "daemonsets"
	Statefulsets             = "statefulsets"
	Jobs                     = "jobs"
	Cronjobs                 = "cronjobs"
	Namespaces               = "namespaces"
	Events                   = "events"
	Roles                    = "roles"
	Clusterroles             = "clusterroles"
	Group                    = "kubeeyeplugins.kubesphere.io"
	Version                  = "v1alpha1"
	Resource                 = "pluginsubscriptions"
	KubeeyeNameSpace         = "kubeeye-system"
	PluginInstalled   string = "installed"
	PluginPause       string = "pause"
	PluginInstalling  string = "installing"
	PluginUninstalled string = "uninstalled"
)

type KubeEyeConfig struct {
	Job *JobConfig `json:"job,omitempty"`
}

type JobConfig struct {
	Image           string                      `json:"image,omitempty"`
	ImagePullPolicy string                      `json:"imagePullPolicy,omitempty"`
	BackLimit       *int32                      `json:"backLimit,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
}
