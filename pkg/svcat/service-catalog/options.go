/*
Copyright 2018 The Kubernetes Authors.

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

package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FilterOptions allows for optional filtering fields to be passed to `Retrieve` methods.
type FilterOptions struct {
	ClassID string
}

// RegisterOptions allows for passing of optional fields to the broker Register method.
type RegisterOptions struct {
	BasicSecret       string
	BearerSecret      string
	CAFile            string
	ClassRestrictions []string
	Namespace         string
	PlanRestrictions  []string
	RelistBehavior    v1beta1.ServiceBrokerRelistBehavior
	RelistDuration    *metav1.Duration
	SkipTLS           bool
}
