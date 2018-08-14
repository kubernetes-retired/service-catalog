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

// Scope is an enum that represents filtering resources by their scope (cluster vs. namespace).
type Scope string

const (
	// ClusterScope filters resources to those defined at the cluster-scope.
	ClusterScope = "cluster"

	// NamespaceScope filters resources to those defined within a namespace.
	NamespaceScope = "namespace"

	// AllScope combines all resources at both the cluster and namespace scopes.
	AllScope = "all"
)

// Matches determines if a particular value is included in the scope.
func (s Scope) Matches(value Scope) bool {
	if s == AllScope {
		return true
	}

	return s == value
}

// ScopeOptions allows for filtering results based on it's namespace and scope (cluster vs. namespaced).
type ScopeOptions struct {
	Namespace string
	Scope     Scope
}
