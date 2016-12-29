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

// Service defines a single service offered by a service broker.
type Service struct {
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	Metadata        interface{}   `json:"metadata, omitempty"`
	Plans           []ServicePlan `json:"plans"`
	DashboardClient interface{}   `json:"dashboard_client"`

	// For k8s object completeness
	runtime.TypeMeta `json:",inline"`

	// Broker implementing this service
	Broker string `json:"broker"`
}
