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

// DeviceDriverKind represents the state of the network node
type DeviceDriverKind string

const (
	// DeviceDriverKindGnmi operates using the gnmi specification
	DeviceDriverKindGnmi DeviceDriverKind = "gnmi"

	// DeviceDriverKindNetconf operates using the netconf specification
	DeviceDriverKindNetconf DeviceDriverKind = "netconf"
)

// DeviceDriverSpec defines the desired state of DeviceDriver
type DeviceDriverSpec struct {
	// Image defines the image to be used for the device driver
	Image string `json:"image,omitempty"`

	// Resources define the resource requirements for the device driver
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true

// DeviceDriver is the Schema for the devicedrivers API
type DeviceDriver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeviceDriverSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// DeviceDriverList contains a list of DeviceDriver
type DeviceDriverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceDriver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceDriver{}, &DeviceDriverList{})
}
