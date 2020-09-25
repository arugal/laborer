module github.com/arugal/laborer

go 1.15

require (
	github.com/docker/docker v1.4.2-0.20190822205725-ed20165a37b4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/component-base v0.17.3
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190822205725-ed20165a37b4
