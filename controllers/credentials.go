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
	"context"
	"fmt"
	"strings"

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Credentials holds the information for authenticating with the Server.
type Credentials struct {
	Username string
	Password string
}

// Validate returns an error if the credentials are invalid
func (creds *Credentials) Validate() error {
	if creds.Username == "" {
		return &CredentialsValidationError{message: "Missing connection detail 'username' in credentials"}
	}
	if creds.Password == "" {
		return &CredentialsValidationError{message: "Missing connection details 'password' in credentials"}
	}
	return nil
}

// CredentialsValidationError is returned when the provided Target credentials
// are invalid (e.g. null)
type CredentialsValidationError struct {
	message string
}

func (e CredentialsValidationError) Error() string {
	return fmt.Sprintf("Validation error with Target credentials: %s",
		e.message)
}

// Make sure the credentials for the network node are valid
func (r *NetworkNodeReconciler) buildAndValidateCredentials(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode) (creds *Credentials, err error) {
	// Retrieve the secret from Kubernetes for this network node
	credsSecret, err := r.getSecret(ctx, req, nn)
	if err != nil {
		return nil, err
	}

	// Check if address is defined on the network node
	if *nn.Spec.Target.Address == "" {
		return nil, &EmptyTargetAddressError{message: "Missing Target connection detail 'Address'"}
	}

	// TODO we could validate the address format

	creds = &Credentials{
		Username: strings.TrimSuffix(string(credsSecret.Data["username"]), "\n"),
		Password: strings.TrimSuffix(string(credsSecret.Data["password"]), "\n"),
	}

	// Verify that the secret contains the expected info.
	err = creds.Validate()
	if err != nil {
		return nil, err
	}

	return creds, nil
}

// Retrieve the secret containing the credentials for talking to the Network Node.
func (r *NetworkNodeReconciler) getSecret(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode) (credsSecret *corev1.Secret, err error) {

	if *nn.Spec.Target.CredentialsName == "" {
		return nil, &EmptyTargetSecretError{message: "The Target secret reference is empty"}
	}
	secretKey := nn.CredentialsKey()
	credsSecret = &corev1.Secret{}
	err = r.Client.Get(ctx, secretKey, credsSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, &ResolveTargetSecretRefError{message: fmt.Sprintf("The Target secret %s does not exist", secretKey)}
		}
		return nil, err
	}

	return credsSecret, nil
}
