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
	"context"
	"errors"
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
)

var (
	errNotAServiceClass = errors.New("not a ServiceClass")
)

// NewSingular returns a new shell of a service class, according to the
// given namespace and name.
func NewSingular(ns, name string) runtime.Object {
	return &servicecatalog.ServiceClass{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// EmptyObject returns an empty service class.
func EmptyObject() runtime.Object {
	return &servicecatalog.ServiceClass{}
}

// NewList returns a new shell of a service class list.
func NewList() runtime.Object {
	return &servicecatalog.ServiceClassList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ServiceClassList",
		},
		Items: []servicecatalog.ServiceClass{},
	}
}

// CheckObject returns a non-nil error if obj is not a service class
// object.
func CheckObject(obj runtime.Object) error {
	_, ok := obj.(*servicecatalog.ServiceClass)
	if !ok {
		return errNotAServiceClass
	}
	return nil
}

// Match determines whether an ServiceInstance matches a field and label
// selector.
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// toSelectableFields returns a field set that represents the object for
// matching purposes.
func toSelectableFields(serviceClass *servicecatalog.ServiceClass) fields.Set {
	// The purpose of allocation with a given number of elements is to reduce
	// amount of allocations needed to create the fields.Set. If you add any
	// field here or the number of object-meta related fields changes, this should
	// be adjusted.
	// You also need to modify
	// pkg/apis/servicecatalog/v1beta1/conversion[_test].go
	scSpecificFieldsSet := make(fields.Set, 3)
	scSpecificFieldsSet["spec.serviceBrokerName"] = serviceClass.Spec.ServiceBrokerName
	scSpecificFieldsSet["spec.externalName"] = serviceClass.Spec.ExternalName
	scSpecificFieldsSet["spec.externalID"] = serviceClass.Spec.ExternalID
	return generic.AddObjectMetaFieldsSet(scSpecificFieldsSet, &serviceClass.ObjectMeta, true)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	serviceclass, ok := obj.(*servicecatalog.ServiceClass)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a ServiceClass")
	}
	return labels.Set(serviceclass.ObjectMeta.Labels), toSelectableFields(serviceclass), serviceclass.Initializers != nil, nil
}

// NewStorage creates a new rest.Storage responsible for accessing
// ServiceClass resources.
func NewStorage(optsGetter generic.RESTOptionsGetter) (serviceClasses, serviceClassStatus rest.Storage, err error) {
	store := registry.Store{
		NewFunc:                  EmptyObject,
		NewListFunc:              NewList,
		PredicateFunc:            Match,
		DefaultQualifiedResource: servicecatalog.Resource("serviceclasses"),

		CreateStrategy:          serviceClassRESTStrategies,
		UpdateStrategy:          serviceClassRESTStrategies,
		DeleteStrategy:          serviceClassRESTStrategies,
		EnableGarbageCollection: true,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, nil, err
	}

	statusStore := store
	statusStore.UpdateStrategy = serviceClassStatusUpdateStrategy

	return &store, &StatusREST{&statusStore}, nil
}

// StatusREST defines the REST operations for the status subresource via
// implementation of various rest interfaces.  It supports the http verbs GET,
// PATCH, and PUT.
type StatusREST struct {
	store *registry.Store
}

var (
	_ rest.Storage = &StatusREST{}
	_ rest.Getter  = &StatusREST{}
	_ rest.Updater = &StatusREST{}
)

// New returns a new ServiceClass.
func (r *StatusREST) New() runtime.Object {
	return &servicecatalog.ServiceClass{}
}

// Get retrieves the object from the storage. It is required to support Patch
// and to implement the rest.Getter interface.
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object and it
// implements rest.Updater interface
func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation)
}
