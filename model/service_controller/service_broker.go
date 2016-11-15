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

package model

// http://apidocs.cloudfoundry.org/239/service_brokers/create_a_service_broker.html

import (
	"k8s.io/client-go/1.5/pkg/runtime"
)

type ServiceBroker struct {
	GUID         string
	Name         string
	BrokerURL    string
	AuthUsername string
	AuthPassword string
	// SpaceGUID    string

	Created int64 `json:",string"`
	Updated int64
	SelfURL string
	// For k8s object completeness
	runtime.TypeMeta `json:",inline"`
}

type CreateServiceBrokerRequest struct {
	Name         string `json:"name"`
	BrokerURL    string `json:"broker_url"`
	AuthUsername string `json:"auth_username"`
	AuthPassword string `json:"auth_password"`
	SpaceGUID    string `json:"space_guid"` // CF-specific - FIXME
}

type CreateServiceBrokerResponse struct {
	Metadata ServiceBrokerMetadata `json:"metadata"`
	Entity   ServiceBrokerEntity   `json:"entity"`
}

type ServiceBrokerMetadata struct {
	GUID      string `json:"guid"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
	URL       string `json:"url"`
}

type ServiceBrokerEntity struct {
	Name         string `json:"name"`
	BrokerURL    string `json:"broker_url"`
	AuthUsername string `json:"auth_username"`
	// space_guid
}
