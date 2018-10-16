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

package server

import (
	"context"

	"github.com/kubernetes-incubator/service-catalog/pkg/storage/etcd"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
)

// Options is the extension of a generic.RESTOptions struct, complete with service-catalog
// specific things
type Options struct {
	EtcdOptions etcd.Options
}

// NewOptions returns a new Options with the given parameters
func NewOptions(
	etcdOpts etcd.Options,
) *Options {
	return &Options{
		EtcdOptions: etcdOpts,
	}
}

// ResourcePrefix gets the resource prefix of all etcd keys
func (o Options) ResourcePrefix() string {
	return o.EtcdOptions.RESTOptions.ResourcePrefix
}

// KeyRootFunc returns the appropriate key root function for the storage type in o.
// This function produces a path that etcd or TPR storage understands, to the root of the resource
// by combining the namespace in the context with the given prefix
func (o Options) KeyRootFunc() func(context.Context) string {
	prefix := o.ResourcePrefix()
	return func(ctx context.Context) string {
		return registry.NamespaceKeyRootFunc(ctx, prefix)
	}
}

// KeyFunc returns the appropriate key function for the storage type in o.
// This function should produce a path that etcd or TPR storage understands, to the resource
// by combining the namespace in the context with the given prefix
func (o Options) KeyFunc(namespaced bool) func(context.Context, string) (string, error) {
	prefix := o.ResourcePrefix()
	return func(ctx context.Context, name string) (string, error) {
		if namespaced {
			return registry.NamespaceKeyFunc(ctx, prefix, name)
		}
		return registry.NoNamespaceKeyFunc(ctx, prefix, name)
	}
}

// GetStorage returns the storage from the given parameters
func (o Options) GetStorage(
	objectType runtime.Object,
	resourcePrefix string,
	scopeStrategy rest.NamespaceScopedStrategy,
	newListFunc func() runtime.Object,
	getAttrsFunc storage.AttrFunc,
	trigger storage.TriggerPublisherFunc,
) (registry.DryRunnableStorage, factory.DestroyFunc) {
	etcdRESTOpts := o.EtcdOptions.RESTOptions
	storageInterface, dFunc := etcdRESTOpts.Decorator(
		etcdRESTOpts.StorageConfig,
		objectType,
		resourcePrefix,
		nil, /* keyFunc for decorator -- looks to be unused everywhere */
		newListFunc,
		getAttrsFunc,
		trigger,
	)
	dryRunnableStorage := registry.DryRunnableStorage{Storage: storageInterface, Codec: etcdRESTOpts.StorageConfig.Codec}
	return dryRunnableStorage, dFunc
}
