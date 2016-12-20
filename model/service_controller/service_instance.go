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

import (
	"k8s.io/client-go/1.5/pkg/runtime"
)

// ServiceInstance defines a single instance of a service.
type ServiceInstance struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	DashboardURL string `json:"dashboard_url"`
	InternalID   string `json:"internal_id, omitempty"`
	// Either use Service/Plan (which are names) or ServiceID /
	// PlanID which are GUIDs. Up to you
	Service string `json:"service"`
	Plan    string `json:"plan"`
	// Either use Service/Plan (which are names) or ServiceID /
	// PlanID which are GUIDs. Up to you
	ServiceID        string `json:"service_id"`
	PlanID           string `json:"plan_id"`
	OrganizationGUID string `json:"organization_guid"`
	SpaceGUID        string `json:"space_guid"`

	LastOperation *LastOperation `json:"last_operation, omitempty"`

	Parameters map[string]interface{} `json:"parameters, omitempty"`

	// For k8s object completeness
	runtime.TypeMeta `json:",inline"`
}

// LastOperation defines the most recent operation performed by the
// service broker, as requested by a client.
type LastOperation struct {
	State                    string `json:"state"`
	Description              string `json:"description"`
	AsyncPollIntervalSeconds int    `json:"async_poll_interval_seconds, omitempty"`
}

// CreateServiceInstanceRequest is the paylaod for the HTTP request to
// create a new service instance.
type CreateServiceInstanceRequest struct {
	Name              string                 `json:"name"`
	OrgID             string                 `json:"organization_guid"`
	Service           string                 `json:"service"`
	Plan              string                 `json:"plan"`
	SpaceID           string                 `json:"space_guid"`
	Parameters        map[string]interface{} `json:"parameters"`
	AcceptsIncomplete bool                   `json:"accepts_incomplete"`
}

// CreateServiceInstanceResponse is the payload of the HTTP response to
// create a new service instance.
type CreateServiceInstanceResponse struct {
	DashboardURL  string         `json:"dashboard_url, omitempty"`
	LastOperation *LastOperation `json:"last_operation, omitempty"`
}
