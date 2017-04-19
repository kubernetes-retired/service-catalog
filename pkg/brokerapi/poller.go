/*
Copyright 2016 The Kubernetes Authors.

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

package brokerapi

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

// PollerCallback will be called after polling the last operation endpoint. Returning true
// will continue polling, false stops it
type PollerCallback func(*v1alpha1.Instance, *LastOperationResponse) bool

// Poller defines the interface for polling for last operation status for
// asynchronous operations.
type Poller interface {
	// CreatePoller
	CreatePoller(BrokerClient, *v1alpha1.Instance, *LastOperationRequest, PollerCallback) error
}
