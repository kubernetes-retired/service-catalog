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

package clusterservicebroker

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
	errNotAClusterServiceBroker = errors.New("not a clusterservicebroker")
)

// NewSingular returns a new shell of a service broker, according to the given namespace and
// name
func NewSingular(ns, name string) runtime.Object {
	return &servicecatalog.ClusterServiceBroker{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterServiceBroker",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// EmptyObject returns an empty broker
func EmptyObject() runtime.Object {
	return &servicecatalog.ClusterServiceBroker{}
}

// NewList returns a new shell of a broker list
func NewList() runtime.Object {
	return &servicecatalog.ClusterServiceBrokerList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterServiceBrokerList",
		},
		Items: []servicecatalog.ClusterServiceBroker{},
	}
}

// CheckObject returns a non-nil error if obj is not a broker object
func CheckObject(obj runtime.Object) error {
	_, ok := obj.(*servicecatalog.ClusterServiceBroker)
	if !ok {
		return errNotAClusterServiceBroker
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

// toSelectableFields returns a field set that represents the object for matching purposes.
func toSelectableFields(broker *servicecatalog.ClusterServiceBroker) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&broker.ObjectMeta, true)
	return generic.MergeFieldsSets(objectMetaFieldsSet, nil)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	broker, ok := obj.(*servicecatalog.ClusterServiceBroker)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a ClusterServiceBroker")
	}
	return labels.Set(broker.ObjectMeta.Labels), toSelectableFields(broker), broker.Initializers != nil, nil
}

// NewStorage creates a new rest.Storage responsible for accessing
// ClusterServiceBroker resources
func NewStorage(optsGetter generic.RESTOptionsGetter) (clusterServiceBrokers, clusterServiceBrokerStatus rest.Storage, err error) {
	store := registry.Store{
		NewFunc:                  EmptyObject,
		NewListFunc:              NewList,
		PredicateFunc:            Match,
		DefaultQualifiedResource: servicecatalog.Resource("clusterservicebrokers"),

		CreateStrategy:          clusterServiceBrokerRESTStrategies,
		UpdateStrategy:          clusterServiceBrokerRESTStrategies,
		DeleteStrategy:          clusterServiceBrokerRESTStrategies,
		EnableGarbageCollection: true,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, nil, err
	}

	statusStore := store
	statusStore.UpdateStrategy = clusterServiceBrokerStatusUpdateStrategy

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

// New returns a new ClusterServiceBroker.
func (r *StatusREST) New() runtime.Object {
	return &servicecatalog.ClusterServiceBroker{}
}

// Get retrieves the object from the storage. It is required to support Patch
// and to implement the rest.Getter interface.
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object and implements the
// rest.Updater interface.
func (r *StatusREST) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation)
}
