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

package serviceclass

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// Match determines whether an Instance matches a field and label
// selector.
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// ToSelectableFields returns a field set that represents the object for matching purposes.
func ToSelectableFields(serviceClass *servicecatalog.ServiceClass) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&serviceClass.ObjectMeta, true)
	return generic.MergeFieldsSets(objectMetaFieldsSet, nil)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	serviceclass, ok := obj.(*servicecatalog.ServiceClass)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a ServiceClass")
	}
	return labels.Set(serviceclass.ObjectMeta.Labels), ToSelectableFields(serviceclass), nil
}

// NewStorage creates a new rest.Storage responsible for accessing ServiceClass
// resources
func NewStorage(opts generic.RESTOptions) rest.Storage {
	prefix := "/" + opts.ResourcePrefix

	newListFunc := func() runtime.Object { return &servicecatalog.ServiceClassList{} }
	storageInterface, dFunc := opts.Decorator(
		opts.StorageConfig,
		1000,
		&servicecatalog.ServiceClass{},
		prefix,
		serviceclassRESTStrategies,
		newListFunc,
		nil,
		storage.NoTriggerPublisher,
	)

	store := registry.Store{
		NewFunc: func() runtime.Object {
			return &servicecatalog.ServiceClass{}
		},
		// NewListFunc returns an object capable of storing results of an etcd list.
		NewListFunc: func() runtime.Object {
			return &servicecatalog.ServiceClassList{}
		},
		// Produces a path that etcd understands, to the root of the resource
		// by combining the namespace in the context with the given prefix
		KeyRootFunc: func(ctx api.Context) string {
			return registry.NamespaceKeyRootFunc(ctx, prefix)
		},
		// Produces a path that etcd understands, to the resource by combining
		// the namespace in the context with the given prefix
		KeyFunc: func(ctx api.Context, name string) (string, error) {
			return registry.NamespaceKeyFunc(ctx, prefix, name)
		},
		// Retrieve the name field of the resource.
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*servicecatalog.ServiceClass).Name, nil
		},
		// Used to match objects based on labels/fields for list.
		PredicateFunc: Match,
		// QualifiedResource should always be plural
		QualifiedResource: api.Resource("serviceclasses"),

		CreateStrategy: serviceclassRESTStrategies,
		UpdateStrategy: serviceclassRESTStrategies,
		DeleteStrategy: serviceclassRESTStrategies,
		Storage:        storageInterface,
		DestroyFunc:    dFunc,
	}

	return &store
}
