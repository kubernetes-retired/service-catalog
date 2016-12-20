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

// Types - no idea what this is for - TODO(Dug)
type Types struct {
	Instance string `json:"instance"`
	Binding  string `json:"binding"`
}

// Defines the kinds of Types
const (
	InstanceType = "instanceType"
	BindingType  = "bindingType"
)
