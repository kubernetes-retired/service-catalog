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

package webhookutil

// WebhookError holds info about error message and its status code
type WebhookError struct {
	errMsg  string
	errCode int32
}

// NewWebhookError returns new error
func NewWebhookError(msg string, code int32) *WebhookError {
	return &WebhookError{
		errMsg:  msg,
		errCode: code,
	}
}

// Error returns error message
func (m *WebhookError) Error() string {
	return m.errMsg
}

// Code returns error status code
func (m *WebhookError) Code() int32 {
	return m.errCode
}
