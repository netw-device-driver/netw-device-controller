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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkDeviceSpec defines the desired state of NetworkDevice
type NetworkDeviceSpec struct {
	// Target defines how we connect to the network node
	Target TargetDetails `json:"target,omitempty"`
}

// DiscoveryStatus defines the states the device driver will report
type DiscoveryStatus string

const (
	// DiscoveryStatusNone means the state is unknown
	DiscoveryStatusNone DiscoveryStatus = ""

	// DiscoveryStatusNotReady means there is insufficient information available to
	// discover the networkDevice
	DiscoveryStatusNotReady DiscoveryStatus = "Not Ready"

	// DiscoveryStatusDiscovery means we are running the discovery on the networkDevice to
	// learn about the hardware components
	DiscoveryStatusDiscovery DiscoveryStatus = "Discovery"

	// DiscoveryStatusReady means the networkDevice can be consumed
	DiscoveryStatusReady DiscoveryStatus = "Ready"

	// DiscoveryStatusDeleting means we are in the process of cleaning up the networkDevice
	// ready for deletion
	DiscoveryStatusDeleting DiscoveryStatus = "Deleting"
)

// CredentialsStatus contains the reference and version of the last
// set of credentials the controller was able to validate.
type CredentialsStatus struct {
	Reference *corev1.SecretReference `json:"credentials,omitempty"`
	Version   string                  `json:"credentialsVersion,omitempty"`
}

// HardwareDetails collects all of the information about hardware
// discovered on the Network Node.
type HardwareDetails struct {
	// the Kind of hardware
	Kind string `json:"kind,omitempty"`
	// the Mac address of the hardware
	MacAddress string `json:"macAddress,omitempty"`
	// the Serial Number of the hardware
	SerialNumber string `json:"serialNumber,omitempty"`
}

// NetworkDeviceStatus defines the observed state of NetworkDevice
type NetworkDeviceStatus struct {
	// DiscoveryStatus holds the discovery status of the networkNode
	// +kubebuilder:validation:Enum="";Enabled;Disabled;Discovery;Deleting
	DiscoveryStatus DiscoveryStatus `json:"discoveryStatus"`

	// OperationalStatus holds the operational status of the networkNode
	// +kubebuilder:validation:Enum="";Up;Down
	OperationalStatus OperationalStatus `json:"operationalStatus"`

	// ErrorType indicates the type of failure encountered when the
	// OperationalStatus is OperationalStatusDown
	// +kubebuilder:validation:Enum="";target error;credential error;discovery error;provisioning error
	ErrorType ErrorType `json:"errorType,omitempty"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// The HardwareDetails discovered on the Network Node.
	HardwareDetails HardwareDetails `json:"hardwareDetails,omitempty"`

	// the last credentials we were able to validate as working
	GoodCredentials CredentialsStatus `json:"goodCredentials,omitempty"`

	// the last credentials we sent to the provisioning backend
	TriedCredentials CredentialsStatus `json:"triedCredentials,omitempty"`

	// the last error message reported by the provisioning subsystem
	ErrorMessage string `json:"errorMessage"`

	// ErrorCount records how many times the host has encoutered an error since the last successful operation
	// +kubebuilder:default:=0
	ErrorCount int `json:"errorCount"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NetworkDevice is the Schema for the networkdevices API
type NetworkDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkDeviceSpec   `json:"spec,omitempty"`
	Status NetworkDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkDeviceList contains a list of NetworkDevice
type NetworkDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkDevice{}, &NetworkDeviceList{})
}
