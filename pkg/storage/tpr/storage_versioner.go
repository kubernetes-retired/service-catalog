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

package tpr

import (
	"strconv"

	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/runtime"
)

type storageVersioner struct {
	singularKind Kind
	listKind     Kind
	checkObject  func(runtime.Object) error
	cl           clientset.Interface
}

// UpdateObject sets storage metadata into an API object. Returns an error if the object
// cannot be updated correctly. May return nil if the requested object does not need metadata
// from database.
func (t *storageVersioner) UpdateObject(obj runtime.Object, resourceVersion uint64) error {
	// if err := t.checkObject(obj); err != nil {
	// 	return err
	// }
	// namespace, err := accessor.Namespace(obj)
	// if err != nil {
	// 	return err
	// }
	// accessor.SetResourceVersion(obj, fmt.Sprintf("%d", resourceVersion))
	// unstruc, err := ToUnstructured(obj)
	// if err != nil {
	// 	return err
	// }
	// cl, err := GetResourceClient(t.cl, t.singularKind, namespace)
	// if err != nil {
	// 	return err
	// }
	// if _, err := cl.Update(unstruc); err != nil {
	// 	return err
	// }
	return nil
}

// UpdateList sets the resource version into an API list object. Returns an error if the object
// cannot be updated correctly. May return nil if the requested object does not need metadata
// from database.
func (t *storageVersioner) UpdateList(obj runtime.Object, resourceVersion uint64) error {
	// ns, err := GetNamespace(obj)
	// if err != nil {
	// 	return err
	// }
	// if rvErr := GetAccessor().SetResourceVersion(
	// 	obj,
	// 	strconv.Itoa(int(resourceVersion)),
	// ); rvErr != nil {
	// 	return rvErr
	// }
	// unstruc, err := ToUnstructured(obj)
	// if err != nil {
	// 	return err
	// }
	// cl, err := GetResourceClient(t.cl, t.listKind, ns)
	// if err != nil {
	// 	return err
	// }
	// if _, err := cl.Update(unstruc); err != nil {
	// 	return err
	// }
	return nil
}

// ObjectResourceVersion returns the resource version (for persistence) of the specified object.
// Should return an error if the specified object does not have a persistable version.
func (t *storageVersioner) ObjectResourceVersion(obj runtime.Object) (uint64, error) {
	vsnStr, err := GetAccessor().ResourceVersion(obj)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseUint(vsnStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}
