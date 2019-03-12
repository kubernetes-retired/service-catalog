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

package tester

import "github.com/appscode/jsonpatch"

// FilterOutStatusPatch filters out the status operation
func FilterOutStatusPatch(operations []jsonpatch.JsonPatchOperation) []jsonpatch.JsonPatchOperation {
	var filtered []jsonpatch.JsonPatchOperation
	for _, op := range operations {
		if op.Path != "/status" {
			filtered = append(filtered, op)
		}
	}

	return filtered
}
