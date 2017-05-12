/*
Copyright 2017 The Kubernetes Authors.

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

package meta

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// GetFinalizers gets the list of finalizers on obj
func GetFinalizers(obj runtime.Object) ([]string, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return accessor.GetFinalizers(), nil
}

// AddFinalizer adds value to the list of finalizers on obj
func AddFinalizer(obj runtime.Object, value string) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	finalizers := append(accessor.GetFinalizers(), value)
	accessor.SetFinalizers(finalizers)
	return nil
}

// RemoveFinalizer removes the given value from the list of finalizers in obj, then returns the list
// of finalizers after value has been removed
func RemoveFinalizer(obj runtime.Object, value string) ([]string, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	finalizers := accessor.GetFinalizers()
	newFinalizers := []string{}
	for _, finalizer := range finalizers {
		if finalizer == value {
			continue
		}
		newFinalizers = append(newFinalizers, finalizer)
	}
	accessor.SetFinalizers(newFinalizers)
	return newFinalizers, nil
}
