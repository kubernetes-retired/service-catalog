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

package fake

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

const (
	ns1   = "ns1"
	ns2   = "ns2"
	tipe1 = "tipe1"
	name1 = "name1"
	name2 = "name2"
	name3 = "name3"
)

//Helpers
func createSingleItemStorage() NamespacedStorage {
	storage := make(NamespacedStorage)
	storage.Set(ns1, tipe1, name1, &servicecatalog.Broker{})
	return storage
}

func createMultipleItemStorage() NamespacedStorage {
	storage := make(NamespacedStorage)
	storage.Set(ns1, tipe1, name1, &servicecatalog.Broker{})
	storage.Set(ns1, tipe1, name2, &servicecatalog.Broker{})
	storage.Set(ns1, tipe1, name3, &servicecatalog.Broker{})

	storage.Set(ns2, tipe1, name1, &servicecatalog.Broker{})
	storage.Set(ns2, tipe1, name2, &servicecatalog.Broker{})

	return storage

}

func createURL(namespace, tipe, name string) string {
	if name == "" {
		return fmt.Sprintf("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/%s/%s", namespace, tipe)
	}
	return fmt.Sprintf("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/%s/%s/%s", namespace, tipe, name)
}

func TestNamespacedStorageSetDelete(t *testing.T) {
	storage := make(NamespacedStorage)

	if nil != storage.Get(ns1, tipe1, name1) {
		t.Fatal("Expected no results from an empty storage")
	}

	storage = createSingleItemStorage()

	if storage.Get(ns1, tipe1, name1) == nil {
		t.Fatal("Expected a object to be stored")
	}

	storage.Delete(ns1, tipe1, name1)
	if nil != storage.Get(ns1, tipe1, name1) {
		t.Fatal("Expected the object to be deleted")
	}
}

func TestNamespacedStorageGetList(t *testing.T) {
	storage := createMultipleItemStorage()

	objects := storage.GetList(ns1, tipe1)
	if count := len(objects); count != 3 {
		t.Fatal("Expected ", 3, "got", count)
	}

	objects = storage.GetList(ns2, tipe1)
	if count := len(objects); count != 2 {
		t.Fatal("Expected ", 2, "got", count)
	}
}

func TestResponseWriter(t *testing.T) {
	rw := newResponseWriter()
	length, err := rw.Write([]byte{0, 0, 0})
	if err != nil {
		t.Fatal("Error writing response", err)
	}

	if length != 3 {
		t.Fatal("Expected length", 3, "got", length)
	}
}

func TestResponseWriterGetResponsePanic(t *testing.T) {
	rw := newResponseWriter()
	// no hearder status set should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected a panic")
		}
	}()
	rw.getResponse()
}

func TestResponseWriterGetResponse(t *testing.T) {
	rw := newResponseWriter()
	rw.WriteHeader(http.StatusFound)

	if response := rw.getResponse(); response.StatusCode != http.StatusFound {
		t.Fatal("Expected", 0, "got", response.StatusCode)
	}
}

func TestGetItemEmptyStorage(t *testing.T) {
	storage := make(NamespacedStorage)
	rw := newResponseWriter()
	url := createURL(ns1, tipe1, name1)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		t.Fatal(err)
	}

	f := getItem(storage)
	f(rw, request)

	if rw.getResponse().StatusCode != http.StatusNotFound && rw.headerSet {
		t.Fatal("Expected NotFoundStatus for empty storage")
	}

}

func TestGetItem(t *testing.T) {
	storage := createSingleItemStorage()
	rw := newResponseWriter()
	url := createURL(ns1, tipe1, name1)
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		t.Fatal(err)
	}

	f := getItem(storage)
	f(rw, request)

	if rw.getResponse().StatusCode != http.StatusOK && rw.headerSet {
		t.Fatal("Expected Status OK", http.StatusOK, "got", rw.getResponse().StatusCode)
	}
	_, err = ioutil.ReadAll(rw.getResponse().Body)

	if err != nil {
		t.Fatal(err)
	}

}
