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

import "fmt"

// EmptyTargetAddressError is returned when the Target address field
// for a network node is empty
type EmptyTargetAddressError struct {
	message string
}

func (e EmptyTargetAddressError) Error() string {
	return fmt.Sprintf("Empty Target address %s",
		e.message)
}

// EmptyTargetSecretError is returned when the Target secret
// for a network node is empty
type EmptyTargetSecretError struct {
	message string
}

func (e EmptyTargetSecretError) Error() string {
	return fmt.Sprintf("No Target CredentialsName defined %s",
		e.message)
}

// ResolveTargetSecretRefError is returned when the Target secret
// for a network node is defined but cannot be found
type ResolveTargetSecretRefError struct {
	message string
}

func (e ResolveTargetSecretRefError) Error() string {
	return fmt.Sprintf("Target CredentialsName secret doesn't exist %s",
		e.message)
}

// ResolveDeviceDriverRefError is returned when the device driver lookup fails
type ResolveDeviceDriverRefError struct {
	message string
}

func (e ResolveDeviceDriverRefError) Error() string {
	return fmt.Sprintf("Device Driver retreive error %s",
		e.message)
}

// CreateNetworkDeviceError is return when the network device cannot get created
type CreateNetworkDeviceError struct {
	message string
}

func (e CreateNetworkDeviceError) Error() string {
	return fmt.Sprintf("Network Device error %s",
		e.message)
}

// UpdateNetworkDeviceError is return when the network device cannot get updated
type UpdateNetworkDeviceError struct {
	message string
}

func (e UpdateNetworkDeviceError) Error() string {
	return fmt.Sprintf("Network Device error %s",
		e.message)
}

// DeleteNetworkDeviceError is return when the network device cannot get deleted
type DeleteNetworkDeviceError struct {
	message string
}

func (e DeleteNetworkDeviceError) Error() string {
	return fmt.Sprintf("Network Device error %s",
		e.message)
}

// GetNetworkDeviceError is return when the network device cannot be retrieved
type GetNetworkDeviceError struct {
	message string
}

func (e GetNetworkDeviceError) Error() string {
	return fmt.Sprintf("Network Device error %s",
		e.message)
}

// SaveNetworkDeviceError is return when the network device cannot be saved
type SaveNetworkDeviceError struct {
	message string
}

func (e SaveNetworkDeviceError) Error() string {
	return fmt.Sprintf("Network Device error %s",
		e.message)
}

// CreateDeploymentError is return when the deployement cannot get created
type CreateDeploymentError struct {
	message string
}

func (e CreateDeploymentError) Error() string {
	return fmt.Sprintf("Deployment error %s",
		e.message)
}

// UpdateDeploymentError is return when the deployement cannot get updated
type UpdateDeploymentError struct {
	message string
}

func (e UpdateDeploymentError) Error() string {
	return fmt.Sprintf("Deployment error %s",
		e.message)
}

// DeleteDeploymentError is return when the deployement cannot get deleted
type DeleteDeploymentError struct {
	message string
}

func (e DeleteDeploymentError) Error() string {
	return fmt.Sprintf("Deployment error %s",
		e.message)
}
