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

package rest

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/crd"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

// ResourceMeta interface provides metadata for specific resource type.
type ResourceMeta interface {
	// HasNamespace defines whether the resource is Namespace-scoped.
	HasNamespace() bool
	// EtcdResource is a resource name for Etcd storage.
	EtcdResource() string
	// TprResource is a resource name for TPR storage.
	TprResource() string
	// TprKind is a resource kind for TPR Storage.
	TprKind() tpr.Kind
	// TprListKind is a resource list kind for TPR Storage.
	TprListKind() tpr.Kind
	// CrdResource is a resource name for CRD storage.
	CrdResource() crd.ResourcePlural
	// CrdKind is a resource kind for CRD storage.
	CrdKind() crd.Kind
	// CrdListKind is a resource list kind for CRD storage.
	CrdListKind() crd.Kind
	// HardDelete propagates delete request to the core API server.
	HardDelete() bool
	// EmptyObject returns an empty object of specific type.
	EmptyObject() runtime.Object
	// NewScopeStrategy returns a new NamespaceScopedStrategy for resource.
	NewScopeStrategy() rest.NamespaceScopedStrategy
	// NewList returns a new shell of a resource list.
	NewList() runtime.Object
	// NewSingular returns a new shell of a specific resource, according to the given
	// namespace and name.
	NewSingular(ns, name string) runtime.Object
	// GetAttrs returns labels and fields of a given object for filtering purposes.
	GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error)
	// CheckObject returns a non-nil error if obj is not an object of specific type.
	CheckObject(obj runtime.Object) error
}

type resourceMetaDelegate struct {
	hasNamespaceVal      bool
	etcdResourceVal      string
	tprResourceVal       string
	tprKindVal           tpr.Kind
	tprListKindVal       tpr.Kind
	crdResourceVal       crd.ResourcePlural
	crdKindVal           crd.Kind
	crdListKindVal       crd.Kind
	hardDeleteVal        bool
	emptyObjectFunc      func() runtime.Object
	newScopeStrategyFunc func() rest.NamespaceScopedStrategy
	newListFunc          func() runtime.Object
	newSingularFunc      func(ns, name string) runtime.Object
	getAttrsFunc         func(obj runtime.Object) (labels.Set, fields.Set, bool, error)
	checkObjectFunc      func(obj runtime.Object) error
}

// HasNamespace defines whether the resource is Namespace-scoped.
func (d *resourceMetaDelegate) HasNamespace() bool {
	return d.hasNamespaceVal
}

// EtcdResource is a resource name for Etcd storage.
func (d *resourceMetaDelegate) EtcdResource() string {
	return d.etcdResourceVal
}

// TprResource is a resource name for TPR storage.
func (d *resourceMetaDelegate) TprResource() string {
	return d.tprResourceVal
}

// TprKind is a resource kind for TPR storage.
func (d *resourceMetaDelegate) TprKind() tpr.Kind {
	return d.tprKindVal
}

// TprListKind is a resource list kind for TPR storage.
func (d *resourceMetaDelegate) TprListKind() tpr.Kind {
	return d.tprListKindVal
}

// CrdResource is a resource name for CRD storage.
func (d *resourceMetaDelegate) CrdResource() crd.ResourcePlural {
	return d.crdResourceVal
}

// CrdKind is a resource kind for CRD storage.
func (d *resourceMetaDelegate) CrdKind() crd.Kind {
	return d.crdKindVal
}

// CrdListKind is a resource list kind for CRD storage.
func (d *resourceMetaDelegate) CrdListKind() crd.Kind {
	return d.crdListKindVal
}

// HardDelete propagates delete request to the core API server.
func (d *resourceMetaDelegate) HardDelete() bool {
	return d.hardDeleteVal
}

// EmptyObject returns an empty object of specific type.
func (d *resourceMetaDelegate) EmptyObject() runtime.Object {
	return d.emptyObjectFunc()
}

// NewScopeStrategy returns a new NamespaceScopedStrategy for resource.
func (d *resourceMetaDelegate) NewScopeStrategy() rest.NamespaceScopedStrategy {
	return d.newScopeStrategyFunc()
}

// NewList returns a new shell of a resource list.
func (d *resourceMetaDelegate) NewList() runtime.Object {
	return d.newListFunc()
}

// NewSingular returns a new shell of a specific resource, according to the given
// namespace and name.
func (d *resourceMetaDelegate) NewSingular(ns, name string) runtime.Object {
	return d.newSingularFunc(ns, name)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func (d *resourceMetaDelegate) GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	return d.getAttrsFunc(obj)
}

// CheckObject returns a non-nil error if obj is not an object of specific type.
func (d *resourceMetaDelegate) CheckObject(obj runtime.Object) error {
	return d.checkObjectFunc(obj)
}
