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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkDeviceSpec defines the desired state of NetworkDevice
type NetworkDeviceSpec struct {
	// Address defines how we connect to the network node
	Address string `json:"address,omitempty"`
}

// DeviceDetails collects information about the deiscovered device
type DeviceDetails struct {
	// Host name
	HostName string `json:"hostname,omitempty"`

	// the Kind of hardware
	Kind string `json:"kind,omitempty"`

	// SW version
	SwVersion string `json:"swVersion,omitempty"`

	// the Mac address of the hardware
	MacAddress string `json:"macAddress,omitempty"`

	// the Serial Number of the hardware
	SerialNumber string `json:"serialNumber,omitempty"`
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
)

// NetworkDeviceStatus defines the observed state of NetworkDevice
type NetworkDeviceStatus struct {
	// The discovered DeviceDetails
	DeviceDetails DeviceDetails `json:"hardwareDetails,omitempty"`

	// DiscoveryStatus holds the discovery status of the networkNode
	// +kubebuilder:validation:Enum="";Ready;Not Ready;Discovery
	// +kubebuilder:default:="Not Ready"
	DiscoveryStatus DiscoveryStatus `json:"discoveryStatus"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
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
