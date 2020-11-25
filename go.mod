module kubeye

go 1.15

require (
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/pkg/errors v0.8.1
	github.com/qri-io/jsonschema v0.1.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/yaml v1.2.0 // indirect
)
