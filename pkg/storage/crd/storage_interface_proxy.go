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

package crd

import (
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
	"strings"
)

const (
	// TODO nilebox: make apiVersion prefixes configurable
	apiPrefix = "servicecatalog.k8s.io"
	crdPrefix = "crd.servicecatalog.k8s.io"
)

// Public API and CRD storage have different apiVersions for the same types
// This proxy modifies apiVersion in input and output objects to the correct one
type storageInterfaceProxy struct {
	storage      storage.Interface
	metaAccessor meta.MetadataAccessor
}

func newProxy(storage storage.Interface) storage.Interface {
	return &storageInterfaceProxy{
		storage:      storage,
		metaAccessor: meta.NewAccessor(),
	}
}

func (t *storageInterfaceProxy) Versioner() storage.Versioner {
	return t.storage.Versioner()
}

func (t *storageInterfaceProxy) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	if err := t.setCrdAPIVersion(obj); err != nil {
		return err
	}
	if err := t.storage.Create(ctx, key, obj, out, ttl); err != nil {
		return err
	}
	if err := t.unsetCrdAPIVersion(out); err != nil {
		return err
	}
	return nil
}

func (t *storageInterfaceProxy) Delete(ctx context.Context, key string, out runtime.Object, preconditions *storage.Preconditions) error {
	if err := t.storage.Delete(ctx, key, out, preconditions); err != nil {
		return err
	}
	if err := t.unsetCrdAPIVersion(out); err != nil {
		return err
	}
	return nil
}

func (t *storageInterfaceProxy) Watch(ctx context.Context, key string, resourceVersion string, p storage.SelectionPredicate) (watch.Interface, error) {
	watchIface, err := t.storage.Watch(ctx, key, resourceVersion, p)
	if err != nil {
		return nil, err
	}
	filteredIFace := watch.Filter(watchIface, watchProxyFilterer(t))
	return filteredIFace, nil
}

func (t *storageInterfaceProxy) WatchList(ctx context.Context, key string, resourceVersion string, p storage.SelectionPredicate) (watch.Interface, error) {
	watchIface, err := t.storage.WatchList(ctx, key, resourceVersion, p)
	if err != nil {
		return nil, err
	}
	filteredIFace := watch.Filter(watchIface, watchProxyFilterer(t))
	return filteredIFace, nil
}

func (t *storageInterfaceProxy) Get(ctx context.Context, key string, resourceVersion string, out runtime.Object, ignoreNotFound bool) error {
	err := t.storage.Get(ctx, key, resourceVersion, out, ignoreNotFound)
	if err != nil {
		return err
	}
	return t.unsetCrdAPIVersion(out)
}

func (t *storageInterfaceProxy) GetToList(ctx context.Context, key string, resourceVersion string, p storage.SelectionPredicate, listObj runtime.Object) error {
	err := t.storage.GetToList(ctx, key, resourceVersion, p, listObj)
	if err != nil {
		return err
	}
	list, err := meta.ExtractList(listObj)
	if err != nil {
		return err
	}
	for _, obj := range list {
		if err := t.unsetCrdAPIVersion(obj); err != nil {
			return err
		}
	}
	return nil
}

func (t *storageInterfaceProxy) List(ctx context.Context, key string, resourceVersion string, p storage.SelectionPredicate, listObj runtime.Object) error {
	err := t.storage.List(ctx, key, resourceVersion, p, listObj)
	if err != nil {
		return err
	}
	list, err := meta.ExtractList(listObj)
	if err != nil {
		return err
	}
	for _, obj := range list {
		if err := t.unsetCrdAPIVersion(obj); err != nil {
			return err
		}
	}
	return nil
}

func (t *storageInterfaceProxy) GuaranteedUpdate(
	ctx context.Context, key string, out runtime.Object, ignoreNotFound bool,
	precondtions *storage.Preconditions, tryUpdate storage.UpdateFunc, suggestion ...runtime.Object) error {
	if len(suggestion) == 1 && suggestion[0] != nil {
		if err := t.setCrdAPIVersion(suggestion[0]); err != nil {
			return err
		}
	}
	err := t.storage.GuaranteedUpdate(ctx, key, out, ignoreNotFound, precondtions, tryUpdate, suggestion...)
	if err != nil {
		return err
	}
	return t.unsetCrdAPIVersion(out)
}

func (t *storageInterfaceProxy) setCrdAPIVersion(in runtime.Object) error {
	apiVersion, err := t.metaAccessor.APIVersion(in)
	if err != nil {
		return err
	}
	if apiVersion == "" {
		return nil
	}
	if !strings.HasPrefix(apiVersion, crdPrefix) {
		newAPIVersion := strings.Replace(apiVersion, apiPrefix, crdPrefix, 1)
		err = t.metaAccessor.SetAPIVersion(in, newAPIVersion)
		return err
	}
	return nil
}

func (t *storageInterfaceProxy) unsetCrdAPIVersion(out runtime.Object) error {
	apiVersion, err := t.metaAccessor.APIVersion(out)
	if err != nil {
		return err
	}
	if apiVersion == "" {
		return nil
	}
	if !strings.HasPrefix(apiVersion, apiPrefix) {
		newAPIVersion := strings.Replace(apiVersion, crdPrefix, apiPrefix, 1)
		err = t.metaAccessor.SetAPIVersion(out, newAPIVersion)
		return err
	}
	return nil
}

func watchProxyFilterer(t *storageInterfaceProxy) func(watch.Event) (watch.Event, bool) {
	return func(in watch.Event) (watch.Event, bool) {
		if in.Object != nil {
			t.unsetCrdAPIVersion(in.Object)
		}
		return in, true
	}
}
