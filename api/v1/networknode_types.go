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
	"k8s.io/apimachinery/pkg/types"
)

const (
	// NetworkNodeFinalizer is the name of the finalizer added to
	// network node to block delete operations until the physical node can be
	// deprovisioned.
	NetworkNodeFinalizer string = "networknode.ndd.henderiw.be"
)

// TargetDetails contains the information necessary to communicate with
// the network node.
type TargetDetails struct {
	// Address holds the IP:port for accessing the network node
	Address *string `json:"address"`

	// Proxy used to communicate to the target network node
	Proxy *string `json:"proxy,omitempty"`

	// The name of the secret containing the credentials (requires
	// keys "username" and "password").
	CredentialsName *string `json:"credentialsName"`

	// The name of the secret containing the credentials (requires
	// keys "TLSCA" and "TLSCert", " TLSKey").
	TLSCredentialsName *string `json:"tlsCredentialsName,omitempty"`

	// SkipVerify disables verification of server certificates when using
	// HTTPS to connect to the Target. This is required when the server
	// certificate is self-signed, but is insecure because it allows a
	// man-in-the-middle to intercept the connection.
	SkipVerify *bool `json:"skpVerify,omitempty"`

	// Insecure runs the communication in an insecure manner
	Insecure *bool `json:"insecure,omitempty"`

	// Encoding defines the gnmi encoding
	Encoding *string `json:"encoding,omitempty"`
}

// DeviceDriverDetails defines the device driver details to connect to the network node
type DeviceDriverDetails struct {
	// Kind defines the device driver kind
	// +kubebuilder:default:=gnmi
	Kind *DeviceDriverKind `json:"kind"`
}

type GrpcServerDetails struct {
	// Port defines the port of the GRPC server for the device driver
	// +kubebuilder:default:=9999
	Port *int `json:"port"`
}

// NetworkNodeSpec defines the desired state of NetworkNode
type NetworkNodeSpec struct {
	// Target defines how we connect to the network node
	Target *TargetDetails `json:"target,omitempty"`

	// DeviceDriver defines the device driver details to connect to the network node
	DeviceDriver *DeviceDriverDetails `json:"deviceDriver,omitempty"`

	// GrpcServerPort defines the grpc server port
	GrpcServer *GrpcServerDetails `json:"grpcServer,omitempty"`
}

// OperationalStatus represents the state of the network node
type OperationalStatus string

const (
	// OperationalStatusNone means the state is unknown
	OperationalStatusNone OperationalStatus = ""

	// OperationalStatusUp is the status value for when the network node is
	// configured correctly and is manageable/operational.
	OperationalStatusUp OperationalStatus = "Up"

	// OperationalStatusDown is the status value for when the network node
	// has any sort of error.
	OperationalStatusDown OperationalStatus = "Down"
)

// IsValid discovery status
func (os OperationalStatus) IsValid() bool {
	switch os {
	case OperationalStatusNone:
	case OperationalStatusUp:
	case OperationalStatusDown:
	default:
		return false
	}
	return true
}

// String2OperationalStatus retuns pointer to enum
func String2OperationalStatus(s string) *OperationalStatus {
	os := OperationalStatus(s)
	if !os.IsValid() {
		panic("Provided error type is not valid")
	}
	return &os
}

// ErrorType indicates the class of problem that has caused the Network Node resource
// to enter an error state.
type ErrorType string

const (
	// NoneError is an error type when no error exists
	NoneError ErrorType = ""

	// TargetError is an error condition occurring when the
	// target supplied are not correct or not existing
	TargetError ErrorType = "target error"

	// CredentialError is an error condition occurring when the
	// credentials supplied are not correct or not existing
	CredentialError ErrorType = "credential error"

	// DeploymentError is an error condition occuring when the controller
	// fails to provision or deprovision the ddriver deployment.
	DeploymentError ErrorType = "deployment error"

	// DeviceDriverError is an error condition occuring when the controller
	// fails to retrieve the ddriver information.
	DeviceDriverError ErrorType = "device driver error"
)

// IsValid discovery status
func (et ErrorType) IsValid() bool {
	switch et {
	case NoneError:
	case TargetError:
	case CredentialError:
	case DeploymentError:
	case DeviceDriverError:
	default:
		return false
	}
	return true
}

// String2ErrorType retuns pointer to enum
func String2ErrorType(s string) *ErrorType {
	et := ErrorType(s)
	if !et.IsValid() {
		panic("Provided error type is not valid")
	}
	return &et
}

// DeviceDetails collects information about the deiscovered device
type DeviceDetails struct {
	// Host name
	HostName *string `json:"hostname,omitempty"`

	// the Kind of hardware
	Kind *string `json:"kind,omitempty"`

	// SW version
	SwVersion *string `json:"swVersion,omitempty"`

	// the Mac address of the hardware
	MacAddress *string `json:"macAddress,omitempty"`

	// the Serial Number of the hardware
	SerialNumber *string `json:"serialNumber,omitempty"`
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

// IsValid discovery status
func (ds DiscoveryStatus) IsValid() bool {
	switch ds {
	case DiscoveryStatusNone:
	case DiscoveryStatusNotReady:
	case DiscoveryStatusDiscovery:
	case DiscoveryStatusReady:
	default:
		return false
	}
	return true
}

// String2DiscoveryStatus retuns pointer to enum
func String2DiscoveryStatus(s string) *DiscoveryStatus {
	ds := DiscoveryStatus(s)
	if !ds.IsValid() {
		panic("Provided discovery status is not valid")
	}
	return &ds
}

// SetDiscoveryStatus updates the DiscoveryStatus field and returns
// true when a change is made or false when no change is made.
/*
func (nn *NetworkNode) SetDiscoveryStatus(status DiscoveryStatus) bool {
	if nn.Status.DiscoveryStatus != String2DiscoveryStatus(fmt.Sprintf("%s", status)) {
		nn.Status.DiscoveryStatus = String2DiscoveryStatus(fmt.Sprintf("%s", status))
		return true
	}
	return false
}
*/

// SetDiscoveryStatus updates the DiscoveryStatus field and returns
// true when a change is made or false when no change is made.
func (nn *NetworkNode) SetDiscoveryStatus(status DiscoveryStatus) bool {
	if nn.Status.DiscoveryStatus == nil {
		nn.Status.DiscoveryStatus = new(DiscoveryStatus)
		*nn.Status.DiscoveryStatus = status
		return true
	}
	if *nn.Status.DiscoveryStatus != status {
		*nn.Status.DiscoveryStatus = status
		return true
	}

	return false
}

// NetworkNodeStatus defines the observed state of NetworkNode
type NetworkNodeStatus struct {
	// OperationalStatus holds the operational status of the networkNode
	// +kubebuilder:validation:Enum="";Up;Down
	// +kubebuilder:default:=Down
	OperationalStatus *OperationalStatus `json:"operationalStatus"`

	// ErrorType indicates the type of failure encountered when the
	// OperationalStatus is OperationalStatusDown
	// +kubebuilder:validation:Enum="";target error;credential error;container error;device driver error
	// +kubebuilder:default:=""
	ErrorType *ErrorType `json:"errorType,omitempty"`

	// ErrorCount records how many times the host has encoutered an error since the last successful operation
	// +kubebuilder:default:=0
	ErrorCount *int `json:"errorCount"`

	// the last error message reported by the provisioning subsystem
	// +kubebuilder:default:=""
	ErrorMessage *string `json:"errorMessage"`

	// DiscoveryStatus holds the discovery status of the networkNode
	// +kubebuilder:validation:Enum="";Ready;Not Ready;Discovery
	DiscoveryStatus *DiscoveryStatus `json:"discoveryStatus,omitempty"`

	// The discovered DeviceDetails
	DeviceDetails *DeviceDetails `json:"deviceDetails,omitempty"`

	// GrpcServerPort defines the grpc server port
	GrpcServer *GrpcServerDetails `json:"grpcServer,omitempty"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// UsedNetworkNodeSpec identifies the used networkNode spec when operational state up
	UsedNetworkNodeSpec *NetworkNodeSpec `json:"usedNetworkNodeSpec,omitempty"`

	// UsedDeviceDriverSpec identifies the used deviceDriver spec in operational state up
	UsedDeviceDriverSpec *DeviceDriverSpec `json:"usedDeviceDriverSpec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="OperationalStatus",type="string",JSONPath=".status.operationalStatus",description="Operational status"
// +kubebuilder:printcolumn:name="DiscoveryStatus",type="string",JSONPath=".status.discoveryStatus",description="Discovery status"
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".status.deviceDetails.kind",description="Kind of device"
// +kubebuilder:printcolumn:name="SwVersion",type="string",JSONPath=".status.deviceDetails.swVersion",description="SW version of the device"
// +kubebuilder:printcolumn:name="MacAddress",type="string",JSONPath=".status.deviceDetails.macAddress",description="macAddress of the device"
// +kubebuilder:printcolumn:name="serialNumber",type="string",JSONPath=".status.deviceDetails.serialNumber",description="serialNumber of the device"
// +kubebuilder:printcolumn:name="grpcServerPort",type="string",JSONPath=".status.grpcServer.port",description="grpc server port to connect to the devic driver"

// NetworkNode is the Schema for the networknodes API
type NetworkNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkNodeSpec   `json:"spec,omitempty"`
	Status NetworkNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkNodeList contains a list of NetworkNode
type NetworkNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkNode{}, &NetworkNodeList{})
}

// CredentialsKey returns a NamespacedName suitable for loading the
// Secret containing the credentials associated with the host.
func (nn *NetworkNode) CredentialsKey() types.NamespacedName {
	return types.NamespacedName{
		Name:      *nn.Spec.Target.CredentialsName,
		Namespace: nn.ObjectMeta.Namespace,
	}
}

// SetOperationalStatus updates the OperationalStatus field and returns
// true when a change is made or false when no change is made.
func (nn *NetworkNode) SetOperationalStatus(status OperationalStatus) bool {
	if nn.Status.OperationalStatus == nil {
		nn.Status.OperationalStatus = new(OperationalStatus)
		*nn.Status.OperationalStatus = status
		return true
	}
	if *nn.Status.OperationalStatus != status {
		*nn.Status.OperationalStatus = status
		return true
	}

	return false
}

// SetErrorType updates the Error Type field and returns
// true when a change is made or false when no change is made.
func (nn *NetworkNode) SetErrorType(errorType ErrorType) bool {
	if nn.Status.ErrorType == nil {
		nn.Status.ErrorType = new(ErrorType)
		*nn.Status.ErrorType = errorType
		return true
	}
	if *nn.Status.ErrorType != errorType {
		*nn.Status.ErrorType = errorType
		return true
	}
	return false
}

// SetErrorMessage updates the Error message field and returns
// true when a change is made or false when no change is made.
func (nn *NetworkNode) SetErrorMessage(errorMessage *string) bool {
	if nn.Status.ErrorMessage == nil {
		nn.Status.ErrorMessage = new(string)
		*nn.Status.ErrorMessage = *errorMessage
	} else {
		if nn.Status.ErrorMessage != errorMessage {
			nn.Status.ErrorMessage = errorMessage
		}
	}
	return false
}

// SetUsedNetworkNodeSpec updates the used network node spec
func (nn *NetworkNode) SetUsedNetworkNodeSpec(nnSpec *NetworkNodeSpec) {
	nn.Status.UsedNetworkNodeSpec = new(NetworkNodeSpec)
	nn.Status.UsedNetworkNodeSpec = nnSpec
}

// SetUsedDeviceDriverSpec updates the used device driver spec
func (nn *NetworkNode) SetUsedDeviceDriverSpec(c *corev1.Container) {
	ddSpec := &DeviceDriverSpec{
		Container: c,
	}
	nn.Status.UsedDeviceDriverSpec = ddSpec

}

// NewEvent creates a new event associated with the object and ready
// to be published to the kubernetes API.
func (nn *NetworkNode) NewEvent(reason, message string) corev1.Event {
	t := metav1.Now()
	return corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: reason + "-",
			Namespace:    nn.ObjectMeta.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       "NetworkNode",
			Namespace:  nn.Namespace,
			Name:       nn.Name,
			UID:        nn.UID,
			APIVersion: GroupVersion.String(),
		},
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "netw-deviceDriver-controller",
		},
		FirstTimestamp:      t,
		LastTimestamp:       t,
		Count:               1,
		Type:                corev1.EventTypeNormal,
		ReportingController: "ndd.henderiw.be/netw-deviceDriver-controller",
	}
}

func DiscoveryStatusPtr(c DiscoveryStatus) *DiscoveryStatus { return &c }
