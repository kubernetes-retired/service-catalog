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

package generator

// generator holds the parameters for generated responses.
type Generator struct {
	Services        Services
	ClassPoolOffset int
	ClassPool       Pool
	DescriptionPool Pool
	PlanPool        Pool
	TagPool         Pool
	MetadataPool    Pool
	RequiresPool    Pool
}

type Pool []string

type Services []Service

type Service struct {
	FromPool Pull
	Plans    Plans
}

type Plans []Plan

type Plan struct {
	FromPool Pull
}

type Pull map[Property]int

type Property string

const (
	Tags                Property = "tags"
	Metadata            Property = "metadata"
	Requires            Property = "Requires"
	Bindable            Property = "bindable"
	BindingsRetrievable Property = "bindings_retrievable"
	Free                Property = "free"
)

type Parameters struct {
	Seed     int64
	Services ServiceRanges
	Plans    PlanRanges
}

type ServiceRanges struct {
	// Plans will default to 1. Range will be [1-Plans)
	Plans               int
	Tags                int
	Metadata            int
	Requires            int
	Bindable            int
	BindingsRetrievable int
}

type PlanRanges struct {
	Metadata int
	Bindable int
	Free     int
}

// Classes can have:
// - tags
// - metadata
// - bindable
// - requires
// - bindings retrievable

// Plans can have:
// - metadata
// - free
// - bindable
