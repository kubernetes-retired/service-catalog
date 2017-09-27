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

package instance

import (
	"errors"
	"fmt"

	scmeta "github.com/kubernetes-incubator/service-catalog/pkg/api/meta"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/client-go/pkg/api"
)

var (
	errNotAnServiceInstance = errors.New("not an instance")
)

// NewSingular returns a new shell of a service instance, according to the given namespace and
// name
func NewSingular(ns, name string) runtime.Object {
	return &servicecatalog.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind: tpr.ServiceInstanceKind.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// EmptyObject returns an empty instance
func EmptyObject() runtime.Object {
	return &servicecatalog.ServiceInstance{}
}

// NewList returns a new shell of an instance list
func NewList() runtime.Object {
	return &servicecatalog.ServiceInstanceList{
		TypeMeta: metav1.TypeMeta{
			Kind: tpr.ServiceInstanceListKind.String(),
		},
		Items: []servicecatalog.ServiceInstance{},
	}
}

// CheckObject returns a non-nil error if obj is not an instance object
func CheckObject(obj runtime.Object) error {
	_, ok := obj.(*servicecatalog.ServiceInstance)
	if !ok {
		return errNotAnServiceInstance
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
func toSelectableFields(instance *servicecatalog.ServiceInstance) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&instance.ObjectMeta, true)
	return generic.MergeFieldsSets(objectMetaFieldsSet, nil)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	instance, ok := obj.(*servicecatalog.ServiceInstance)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not an ServiceInstance")
	}
	return labels.Set(instance.ObjectMeta.Labels), toSelectableFields(instance), instance.Initializers != nil, nil
}

// NewStorage creates a new rest.Storage responsible for accessing ServiceInstance
// resources
func NewStorage(opts server.Options) (rest.Storage, rest.Storage, rest.Storage) {
	prefix := "/" + opts.ResourcePrefix()

	storageInterface, dFunc := opts.GetStorage(
		1000,
		&servicecatalog.ServiceInstance{},
		prefix,
		instanceRESTStrategies,
		NewList,
		nil,
		storage.NoTriggerPublisher,
	)

	store := registry.Store{
		NewFunc:     EmptyObject,
		NewListFunc: NewList,
		KeyRootFunc: opts.KeyRootFunc(),
		KeyFunc:     opts.KeyFunc(true),
		// Retrieve the name field of the resource.
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return scmeta.GetAccessor().Name(obj)
		},
		// Used to match objects based on labels/fields for list.
		PredicateFunc: Match,
		// QualifiedResource should always be plural
		QualifiedResource: api.Resource("serviceinstances"),

		CreateStrategy:          instanceRESTStrategies,
		UpdateStrategy:          instanceRESTStrategies,
		DeleteStrategy:          instanceRESTStrategies,
		EnableGarbageCollection: true,

		Storage:     storageInterface,
		DestroyFunc: dFunc,
	}

	statusStore := store
	statusStore.UpdateStrategy = instanceStatusUpdateStrategy

	referencesStore := store
	referencesStore.UpdateStrategy = instanceReferencesUpdateStrategy

	return &store, &StatusREST{&statusStore}, &referencesStore

}

// StatusREST defines the REST operations for the status subresource.
type StatusREST struct {
	store *registry.Store
}

// New returns a new ServiceInstance
func (r *StatusREST) New() runtime.Object {
	return &servicecatalog.ServiceInstance{}
}

// Get retrieves the object from the storage. It is required to support Patch
// and to implement the rest.Getter interface.
func (r *StatusREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object and it
// implements rest.Updater interface
func (r *StatusREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}
