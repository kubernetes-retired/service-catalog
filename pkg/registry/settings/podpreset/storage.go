/*
Copyright 2015 The Kubernetes Authors.

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

package podpreset

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/client-go/pkg/api"
	// "k8s.io/kubernetes/pkg/registry/cachesize"

	scmeta "github.com/kubernetes-incubator/service-catalog/pkg/api/meta"
	settingsapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
)

var (
	errNotAPodPreset = errors.New("not a podpreset")
)

// NewSingular returns a new shell of a podpreset, according to the given namespace and
// name
func NewSingular(ns, name string) runtime.Object {
	return &settingsapi.PodPreset{
		TypeMeta: metav1.TypeMeta{
		// Kind: tpr.PodPresetKind.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// EmptyObject returns an empty podpreset.
func EmptyObject() runtime.Object {
	return &settingsapi.PodPreset{}
}

// NewList returns a new shell of an podpreset list
func NewList() runtime.Object {
	return &settingsapi.PodPresetList{
		TypeMeta: metav1.TypeMeta{
		// Kind: tpr.PodPresetListKind.String(),
		},
		Items: []settingsapi.PodPreset{},
	}
}

// CheckObject returns a non-nil error if obj is not an podpreset object
func CheckObject(obj runtime.Object) error {
	_, ok := obj.(*settingsapi.PodPreset)
	if !ok {
		return errNotAPodPreset
	}
	return nil
}

// NewStorage creates a new rest.Storage responsible for accessing PodPreset
// resources
func NewStorage(opts server.Options) (rest.Storage, error) {
	prefix := "/" + opts.ResourcePrefix()

	storageInterface, dFunc := opts.GetStorage(
		1000,
		&settingsapi.PodPreset{},
		prefix,
		podPresetRESTStrategy,
		NewList,
		nil,
		storage.NoTriggerPublisher,
	)

	store := genericregistry.Store{
		NewFunc:     EmptyObject,
		NewListFunc: NewList,
		KeyRootFunc: opts.KeyRootFunc(),
		KeyFunc:     opts.KeyFunc(true),
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return scmeta.GetAccessor().Name(obj)
		},
		PredicateFunc:     Matcher,
		Copier:            api.Scheme,
		QualifiedResource: settingsapi.Resource("podpresets"),
		// WatchCacheSize:    cachesize.GetWatchCacheSizeByResource("podpresets"),

		CreateStrategy:          podPresetRESTStrategy,
		UpdateStrategy:          podPresetRESTStrategy,
		DeleteStrategy:          podPresetRESTStrategy,
		EnableGarbageCollection: true,

		Storage:     storageInterface,
		DestroyFunc: dFunc,
	}

	return &store, nil
}
