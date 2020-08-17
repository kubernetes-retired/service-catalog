/*
Copyright 2019 The Kubernetes Authors.

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

package v2

import (
	"fmt"
	"net/http"
)

// HTTPStatusCodeError is an error type that provides additional information
// based on the Open Service Broker API conventions for returning information
// about errors.  If the response body provided by the broker to any client
// operation is malformed, an error of this type will be returned with the
// ResponseError field set to the unmarshalling error.
//
// These errors may optionally provide a machine-readable error message and
// human-readable description.
//
// The IsHTTPError method checks whether an error is of this type.
//
// Checks for important errors in the API specification are implemented as
// utility methods:
//
// - IsGoneError
// - IsConflictError
// - IsAsyncRequiredError
// - IsAppGUIDRequiredError
type HTTPStatusCodeError struct {
	// StatusCode is the HTTP status code returned by the broker.
	StatusCode int
	// ErrorMessage is a machine-readable error string that may be returned by
	// the broker.
	ErrorMessage *string
	// Description is a human-readable description of the error that may be
	// returned by the broker.
	Description *string
	// ResponseError is set to the error that occurred when unmarshalling a
	// response body from the broker.
	ResponseError error
}

func (e HTTPStatusCodeError) Error() string {
	errorMessage := "<nil>"
	description := "<nil>"

	if e.ErrorMessage != nil {
		errorMessage = *e.ErrorMessage
	}
	if e.Description != nil {
		description = *e.Description
	}
	return fmt.Sprintf("Status: %v; ErrorMessage: %v; Description: %v; ResponseError: %v", e.StatusCode, errorMessage, description, e.ResponseError)
}

// IsHTTPError returns whether the error represents an HTTPStatusCodeError.  A
// client method returning an HTTP error indicates that the broker returned an
// error code and a correctly formed response body.
func IsHTTPError(err error) (*HTTPStatusCodeError, bool) {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if ok {
		return &statusCodeError, ok
	}

	statusCodeErrorPointer, ok := err.(*HTTPStatusCodeError)
	if ok {
		return statusCodeErrorPointer, ok
	}

	return nil, ok
}

// IsGoneError returns whether the error represents an HTTP GONE status.
func IsGoneError(err error) bool {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if !ok {
		return false
	}

	return statusCodeError.StatusCode == http.StatusGone
}

// IsConflictError returns whether the error represents a conflict.
func IsConflictError(err error) bool {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if !ok {
		return false
	}

	return statusCodeError.StatusCode == http.StatusConflict
}

// Constants are used to check for spec-mandated errors and their messages
const (
	AsyncErrorMessage               = "AsyncRequired"
	AsyncErrorDescription           = "This service plan requires client support for asynchronous service operations."
	AppGUIDRequiredErrorMessage     = "RequiresApp"
	AppGUIDRequiredErrorDescription = "This service supports generation of credentials through binding an application only."
	ConcurrencyErrorMessage         = "ConcurrencyError"
	ConcurrencyErrorDescription     = "The Service Broker does not support concurrent requests that mutate the same resource."
)

// IsAsyncRequiredError returns whether the error corresponds to the
// conventional way of indicating that a service requires asynchronous
// operations to perform an action.
func IsAsyncRequiredError(err error) bool {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if !ok {
		return false
	}

	if statusCodeError.StatusCode != http.StatusUnprocessableEntity {
		return false
	}

	if statusCodeError.ErrorMessage == nil || statusCodeError.Description == nil {
		return false
	}

	if *statusCodeError.ErrorMessage != AsyncErrorMessage {
		return false
	}

	return *statusCodeError.Description == AsyncErrorDescription
}

// IsAppGUIDRequiredError returns whether the error corresponds to the
// conventional way of indicating that a service only supports credential-type
// bindings.
func IsAppGUIDRequiredError(err error) bool {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if !ok {
		return false
	}

	if statusCodeError.StatusCode != http.StatusUnprocessableEntity {
		return false
	}

	if statusCodeError.ErrorMessage == nil || statusCodeError.Description == nil {
		return false
	}

	if *statusCodeError.ErrorMessage != AppGUIDRequiredErrorMessage {
		return false
	}

	return *statusCodeError.Description == AppGUIDRequiredErrorDescription
}

// IsConcurrencyError returns whether the error corresponds to the
// conventional way of indicating that a service broker does not support
// concurrent requests to modify the same resource
func IsConcurrencyError(err error) bool {
	statusCodeError, ok := err.(HTTPStatusCodeError)
	if !ok {
		return false
	}

	if statusCodeError.StatusCode != http.StatusUnprocessableEntity {
		return false
	}

	if statusCodeError.ErrorMessage == nil || statusCodeError.Description == nil {
		return false
	}

	if *statusCodeError.ErrorMessage != ConcurrencyErrorMessage {
		return false
	}

	return *statusCodeError.Description == ConcurrencyErrorDescription
}

// AlphaAPIMethodsNotAllowedError is an error type signifying that alpha API
// methods are not allowed for this client's API Version or alpha opt-in.
type AlphaAPIMethodsNotAllowedError struct {
	reason string
}

func (e AlphaAPIMethodsNotAllowedError) Error() string {
	return fmt.Sprintf(
		"alpha API methods not allowed: %s",
		e.reason,
	)
}

// OperationNotAllowedError is an error type signifying that an operation
// is not allowed for this client.
type OperationNotAllowedError struct {
	reason string
}

func (e OperationNotAllowedError) Error() string {
	return fmt.Sprintf(
		"operation not allowed: %s",
		e.reason,
	)
}

// GetInstanceNotAllowedError is an error type signifying that doing a GET to
// fetch a service instance is not allowed for this client.
type GetInstanceNotAllowedError struct {
	reason string
}

func (e GetInstanceNotAllowedError) Error() string {
	return fmt.Sprintf(
		"GetInstance not allowed: %s",
		e.reason,
	)
}

// GetBindingNotAllowedError is an error type signifying that doing a GET to
// fetch a binding is not allowed for this client.
type GetBindingNotAllowedError struct {
	reason string
}

func (e GetBindingNotAllowedError) Error() string {
	return fmt.Sprintf(
		"GetBinding not allowed: %s",
		e.reason,
	)
}

// AsyncBindingOperationsNotAllowedError is an error type signifying that asynchronous
// binding operations (bind/unbind/poll) are not allowed for this client.
type AsyncBindingOperationsNotAllowedError struct {
	reason string
}

func (e AsyncBindingOperationsNotAllowedError) Error() string {
	return fmt.Sprintf("Asynchronous binding operations are not allowed: %s", e.reason)
}

// IsAsyncBindingOperationsNotAllowedError returns whether the error represents asynchronous
// binding operations (bind/unbind/poll) not being allowed for this client.
func IsAsyncBindingOperationsNotAllowedError(err error) bool {
	_, ok := err.(AsyncBindingOperationsNotAllowedError)
	return ok
}
