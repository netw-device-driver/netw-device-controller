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

package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	labelNetworkNodeNamespace = "namespace"
	labelNetworkNodeName      = "host"
	labelErrorType            = "error_type"
	labelNetworkNodeDataType  = "networkNode_data_type"
)

var reconcileCounters = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "ndd_reconcile_total",
	Help: "The number of times network nodes have been reconciled",
}, []string{labelNetworkNodeNamespace, labelNetworkNodeName})
var reconcileErrorCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ndd_reconcile_error_total",
	Help: "The number of times the operator has failed to reconcile a network node",
})
var credentialsMissing = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ndd_credentials_missing_total",
	Help: "Number of times a network node's credentials are found to be missing",
})
var credentialsInvalid = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ndd_credentials_invalid_total",
	Help: "Number of times a network node's credentials are found to be invalid",
})
var unhandledCredentialsError = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ndd_credentials_unhandled_error_total",
	Help: "Number of times getting a network node's credentials fails in an unexpected way",
})
var deviceDriverError = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "ndd_device_driver_error_total",
	Help: "Number of times a network node's device driver are found to be missing",
})

func init() {
	metrics.Registry.MustRegister(
		reconcileCounters,
		reconcileErrorCounter,
		credentialsMissing,
		credentialsInvalid,
		unhandledCredentialsError,
		deviceDriverError,
	)
}

func networkNodeMetricLabels(request ctrl.Request) prometheus.Labels {
	return prometheus.Labels{
		labelNetworkNodeNamespace: request.Namespace,
		labelNetworkNodeName:      request.Name,
	}
}
