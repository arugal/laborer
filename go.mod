module github.com/arugal/laborer

go 1.15

require (
	github.com/antihax/optional v1.0.0
	github.com/docker/docker v1.4.2-0.20190822205725-ed20165a37b4
	github.com/scultura-org/harborapi v0.0.0-20201101061223-00bd5186364a
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	gomodules.xyz/jsonpatch/v2 v2.1.0
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	k8s.io/component-base v0.20.1
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.7.0
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190822205725-ed20165a37b4
