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

package podpreset

import (
	"errors"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	settingsapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings"
)

var (
	errNotAPodPreset = errors.New("not a podpreset")
)

// EmptyObject returns an empty podpreset.
func EmptyObject() runtime.Object {
	return &settingsapi.PodPreset{}
}

// NewList returns a new shell of an podpreset list
func NewList() runtime.Object {
	return &settingsapi.PodPresetList{}
}

// NewStorage creates a new rest.Storage responsible for accessing PodPreset
// resources
func NewStorage(optsGetter generic.RESTOptionsGetter) (serviceBindings rest.Storage, err error) {
	store := registry.Store{
		NewFunc:                  EmptyObject,
		NewListFunc:              NewList,
		PredicateFunc:            Matcher,
		DefaultQualifiedResource: settingsapi.Resource("podpresets"),

		CreateStrategy:          podPresetRESTStrategy,
		UpdateStrategy:          podPresetRESTStrategy,
		DeleteStrategy:          podPresetRESTStrategy,
		EnableGarbageCollection: true,
	}

	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &store, nil
}
