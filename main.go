/*
Copyright 2021 Wim Henderickx.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/component-base/version"
	"k8s.io/klog/klogr"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	"github.com/netw-device-driver/netw-device-controller/controllers"
	"github.com/spf13/pflag"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(nddv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func setupChecks(mgr ctrl.Manager) {
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}
}

var (
	metricsAddr          string
	healthAddr           string
	enableLeaderElection bool
	nddConcurrency       int
)

// InitFlags funntion
func InitFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")

	fs.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	fs.StringVar(&healthAddr, "health-probe-bind-address", ":9440",
		"The address the health endpoint binds to.")

	fs.IntVar(&nddConcurrency, "ndd-oncurrency", 1,
		"Number of objects to process simultaneously")
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctrl.SetLogger(klogr.New())

	setupLog.Info("setting up flag in manager...")

	InitFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: healthAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "a8f17801.henderiw.be",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup the context that's going to be used in controllers and for the manager.
	ctx := ctrl.SetupSignalHandler()

	if err = (&controllers.NetworkNodeReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("NetworkNode"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, concurrency(nddConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkNode")
		os.Exit(1)
	}

	setupChecks(mgr)

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager", "version", version.Get().String())
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}

func concurrency(c int) controller.Options {
	return controller.Options{MaxConcurrentReconciles: c}
}
