module github.com/netw-device-driver/netw-device-controller

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/metal3-io/baremetal-operator v0.0.0-20210310165305-c74264a6ddcb
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
	k8s.io/component-base v0.19.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.7.0
)
