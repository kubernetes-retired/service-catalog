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

package binding

import (
	"testing"

	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/client-go/pkg/api"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/etcd"
	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"
)

func TestNewListNilItems(t *testing.T) {
	newList := NewList()
	realObj := newList.(*servicecatalog.BindingList)

	if realObj.Items == nil {
		t.Fatalf("nil incorrectly set on Items field")
	}
}

func TestNewStorageWithTPR(t *testing.T) {
	opts := server.NewOptions(
		etcd.Options{},
		tpr.Options{
			RESTOptions: generic.RESTOptions{
				StorageConfig: storagebackend.NewDefaultConfig("", api.Scheme, nil),
			},
		},
		server.StorageTypeTPR,
	)

	storage, statusStorage, err := NewStorage(*opts)

	if err != nil {
		t.Fatalf("unexpected error ocurred (%s)", err)
	}

	if storage.(*registry.Store).EnableGarbageCollection != false {
		t.Fatalf("'EnableGarbageCollection` in storage should be set to false for TPR type")
	}

	if statusStorage.(*registry.Store).EnableGarbageCollection != false {
		t.Fatalf("'EnableGarbageCollection` in statusStorage should be set to false for TPR type")
	}
}
