// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceDetails) DeepCopyInto(out *DeviceDetails) {
	*out = *in
	if in.HostName != nil {
		in, out := &in.HostName, &out.HostName
		*out = new(string)
		**out = **in
	}
	if in.Kind != nil {
		in, out := &in.Kind, &out.Kind
		*out = new(string)
		**out = **in
	}
	if in.SwVersion != nil {
		in, out := &in.SwVersion, &out.SwVersion
		*out = new(string)
		**out = **in
	}
	if in.MacAddress != nil {
		in, out := &in.MacAddress, &out.MacAddress
		*out = new(string)
		**out = **in
	}
	if in.SerialNumber != nil {
		in, out := &in.SerialNumber, &out.SerialNumber
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceDetails.
func (in *DeviceDetails) DeepCopy() *DeviceDetails {
	if in == nil {
		return nil
	}
	out := new(DeviceDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceDriver) DeepCopyInto(out *DeviceDriver) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceDriver.
func (in *DeviceDriver) DeepCopy() *DeviceDriver {
	if in == nil {
		return nil
	}
	out := new(DeviceDriver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DeviceDriver) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceDriverDetails) DeepCopyInto(out *DeviceDriverDetails) {
	*out = *in
	if in.Kind != nil {
		in, out := &in.Kind, &out.Kind
		*out = new(DeviceDriverKind)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceDriverDetails.
func (in *DeviceDriverDetails) DeepCopy() *DeviceDriverDetails {
	if in == nil {
		return nil
	}
	out := new(DeviceDriverDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceDriverList) DeepCopyInto(out *DeviceDriverList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DeviceDriver, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceDriverList.
func (in *DeviceDriverList) DeepCopy() *DeviceDriverList {
	if in == nil {
		return nil
	}
	out := new(DeviceDriverList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DeviceDriverList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceDriverSpec) DeepCopyInto(out *DeviceDriverSpec) {
	*out = *in
	if in.Container != nil {
		in, out := &in.Container, &out.Container
		*out = new(corev1.Container)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceDriverSpec.
func (in *DeviceDriverSpec) DeepCopy() *DeviceDriverSpec {
	if in == nil {
		return nil
	}
	out := new(DeviceDriverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrpcServerDetails) DeepCopyInto(out *GrpcServerDetails) {
	*out = *in
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrpcServerDetails.
func (in *GrpcServerDetails) DeepCopy() *GrpcServerDetails {
	if in == nil {
		return nil
	}
	out := new(GrpcServerDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkNode) DeepCopyInto(out *NetworkNode) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkNode.
func (in *NetworkNode) DeepCopy() *NetworkNode {
	if in == nil {
		return nil
	}
	out := new(NetworkNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkNode) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkNodeList) DeepCopyInto(out *NetworkNodeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NetworkNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkNodeList.
func (in *NetworkNodeList) DeepCopy() *NetworkNodeList {
	if in == nil {
		return nil
	}
	out := new(NetworkNodeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NetworkNodeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkNodeSpec) DeepCopyInto(out *NetworkNodeSpec) {
	*out = *in
	if in.Target != nil {
		in, out := &in.Target, &out.Target
		*out = new(TargetDetails)
		(*in).DeepCopyInto(*out)
	}
	if in.DeviceDriver != nil {
		in, out := &in.DeviceDriver, &out.DeviceDriver
		*out = new(DeviceDriverDetails)
		(*in).DeepCopyInto(*out)
	}
	if in.GrpcServer != nil {
		in, out := &in.GrpcServer, &out.GrpcServer
		*out = new(GrpcServerDetails)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkNodeSpec.
func (in *NetworkNodeSpec) DeepCopy() *NetworkNodeSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkNodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkNodeStatus) DeepCopyInto(out *NetworkNodeStatus) {
	*out = *in
	if in.OperationalStatus != nil {
		in, out := &in.OperationalStatus, &out.OperationalStatus
		*out = new(OperationalStatus)
		**out = **in
	}
	if in.ErrorType != nil {
		in, out := &in.ErrorType, &out.ErrorType
		*out = new(ErrorType)
		**out = **in
	}
	if in.ErrorCount != nil {
		in, out := &in.ErrorCount, &out.ErrorCount
		*out = new(int)
		**out = **in
	}
	if in.ErrorMessage != nil {
		in, out := &in.ErrorMessage, &out.ErrorMessage
		*out = new(string)
		**out = **in
	}
	if in.DiscoveryStatus != nil {
		in, out := &in.DiscoveryStatus, &out.DiscoveryStatus
		*out = new(DiscoveryStatus)
		**out = **in
	}
	if in.DeviceDetails != nil {
		in, out := &in.DeviceDetails, &out.DeviceDetails
		*out = new(DeviceDetails)
		(*in).DeepCopyInto(*out)
	}
	if in.GrpcServer != nil {
		in, out := &in.GrpcServer, &out.GrpcServer
		*out = new(GrpcServerDetails)
		(*in).DeepCopyInto(*out)
	}
	if in.LastUpdated != nil {
		in, out := &in.LastUpdated, &out.LastUpdated
		*out = (*in).DeepCopy()
	}
	if in.UsedNetworkNodeSpec != nil {
		in, out := &in.UsedNetworkNodeSpec, &out.UsedNetworkNodeSpec
		*out = new(NetworkNodeSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.UsedDeviceDriverSpec != nil {
		in, out := &in.UsedDeviceDriverSpec, &out.UsedDeviceDriverSpec
		*out = new(DeviceDriverSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkNodeStatus.
func (in *NetworkNodeStatus) DeepCopy() *NetworkNodeStatus {
	if in == nil {
		return nil
	}
	out := new(NetworkNodeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetDetails) DeepCopyInto(out *TargetDetails) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(string)
		**out = **in
	}
	if in.Proxy != nil {
		in, out := &in.Proxy, &out.Proxy
		*out = new(string)
		**out = **in
	}
	if in.CredentialsName != nil {
		in, out := &in.CredentialsName, &out.CredentialsName
		*out = new(string)
		**out = **in
	}
	if in.TLSCredentialsName != nil {
		in, out := &in.TLSCredentialsName, &out.TLSCredentialsName
		*out = new(string)
		**out = **in
	}
	if in.SkipVerify != nil {
		in, out := &in.SkipVerify, &out.SkipVerify
		*out = new(bool)
		**out = **in
	}
	if in.Insecure != nil {
		in, out := &in.Insecure, &out.Insecure
		*out = new(bool)
		**out = **in
	}
	if in.Encoding != nil {
		in, out := &in.Encoding, &out.Encoding
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetDetails.
func (in *TargetDetails) DeepCopy() *TargetDetails {
	if in == nil {
		return nil
	}
	out := new(TargetDetails)
	in.DeepCopyInto(out)
	return out
}
