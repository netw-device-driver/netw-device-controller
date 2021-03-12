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
