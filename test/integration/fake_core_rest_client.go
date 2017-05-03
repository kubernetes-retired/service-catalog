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

package integration

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/pkg/api"
	fakerestclient "k8s.io/client-go/rest/fake"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

type objStorage map[string]runtime.Object
type typedStorage map[string]objStorage
type namespacedStorage map[string]typedStorage

func (s namespacedStorage) set(ns, tipe, name string, obj runtime.Object) {
	if _, ok := s[ns]; !ok {
		s[ns] = make(typedStorage)
	}
	if _, ok := s[ns][tipe]; !ok {
		s[ns][tipe] = make(objStorage)
	}
	s[ns][tipe][name] = obj
}

func (s namespacedStorage) getList(ns, tipe string) []runtime.Object {
	itemMap, ok := s[ns][tipe]
	if !ok {
		return []runtime.Object{}
	}
	items := make([]runtime.Object, 0, len(itemMap))
	for _, item := range itemMap {
		items = append(items, item)
	}
	return items
}

func (s namespacedStorage) get(ns, tipe, name string) runtime.Object {
	item, ok := s[ns][tipe][name]
	if !ok {
		return nil
	}
	return item
}

func (s namespacedStorage) delete(ns, tipe, name string) {
	delete(s[ns][tipe], name)
}

var (
	accessor = meta.NewAccessor()
	storage  = make(namespacedStorage)
)

func getFakeCoreRESTClient() *fakerestclient.RESTClient {
	return &fakerestclient.RESTClient{
		Client: fakerestclient.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
			r := getRouter()
			rw := newResponseWriter()
			r.ServeHTTP(rw, request)
			return rw.getResponse(), nil
		}),
		NegotiatedSerializer: serializer.DirectCodecFactory{
			CodecFactory: api.Codecs,
		},
		APIRegistry: api.Registry,
	}
}

type responseWriter struct {
	header    http.Header
	headerSet bool
	body      []byte
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		header: make(http.Header),
	}
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(bytes []byte) (int, error) {
	if !rw.headerSet {
		rw.WriteHeader(http.StatusOK)
	}
	rw.body = append(rw.body, bytes...)
	return len(bytes), nil
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.headerSet = true
	rw.header.Set("status", strconv.Itoa(status))
}

func (rw *responseWriter) getResponse() *http.Response {
	status, err := strconv.ParseInt(rw.header.Get("status"), 10, 16)
	if err != nil {
		panic(err)
	}
	return &http.Response{
		StatusCode: int(status),
		Header:     rw.header,
		Body:       ioutil.NopCloser(bytes.NewBuffer(rw.body)),
	}
}

func getRouter() http.Handler {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/{namespace}/{type}", getItems).Methods("GET")
	r.HandleFunc("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/{namespace}/{type}", createItem).Methods("POST")
	r.HandleFunc("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/{namespace}/{type}/{name}", getItem).Methods("GET")
	r.HandleFunc("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/{namespace}/{type}/{name}", updateItem).Methods("PUT")
	r.HandleFunc("/apis/servicecatalog.k8s.io/v1alpha1/namespaces/{namespace}/{type}/{name}", deleteItem).Methods("DELETE")
	r.HandleFunc("/apis/servicecatalog.k8s.i0/v1alpha1/watch/namespaces/{namespace}/{type}/{name}", watchItem).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return r
}

func getItems(rw http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	tipe := mux.Vars(r)["type"]
	objs := storage.getList(ns, tipe)
	items := make([]runtime.Object, 0, len(objs))
	for _, obj := range objs {
		// We need to strip away typemeta, but we don't want to tamper with what's
		// in memory, so we're going to make a deep copy first. We can actually
		// convert from a *runtime.Object to a *v1alpha1.Broker at the same
		// time!
		objCopy, err := conversion.NewCloner().DeepCopy(obj)
		if err != nil {
			log.Fatalf("error performing deep copy: %s", err)
		}
		item, ok := objCopy.(runtime.Object)
		if !ok {
			log.Fatalf("error performing type assertion: %s", err)
		}
		accessor.SetKind(item, "")
		accessor.SetAPIVersion(item, "")
		items = append(items, item)
	}
	list := &api.List{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "servicecatalog.k8s.io/v1alpha1",
		},
		Items: items,
	}
	// list is actually an *api.List, but we're going to encode it as if it were
	// a list for a specific type. This cheat sends the right bytes over the wire.
	var codec runtime.Codec
	var err error
	switch tipe {
	case "brokers":
		accessor.SetKind(list, "BrokerList")
		codec, err = testapi.GetCodecForObject(&v1alpha1.BrokerList{})
	case "serviceclasses":
		accessor.SetKind(list, "ServiceClassList")
		codec, err = testapi.GetCodecForObject(&v1alpha1.ServiceClassList{})
	case "instances":
		accessor.SetKind(list, "InstanceList")
		codec, err = testapi.GetCodecForObject(&v1alpha1.InstanceList{})
	case "bindings":
		accessor.SetKind(list, "BindingList")
		codec, err = testapi.GetCodecForObject(&v1alpha1.BindingList{})
	default:
		log.Fatalf("unrecognized resource type: %s", tipe)
	}
	if err != nil {
		log.Fatalf("error getting codec: %s", err)
	}
	listBytes, err := runtime.Encode(codec, list)
	if err != nil {
		log.Fatalf("error encoding list: %s", err)
	}
	rw.Write(listBytes)
}

func createItem(rw http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	tipe := mux.Vars(r)["type"]
	// TODO: Is there some type-agnostic way to get the codec?
	codec, err := testapi.GetCodecForObject(&v1alpha1.Broker{})
	if err != nil {
		log.Fatalf("error getting codec: %s", err)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("error getting body bytes: %s", err)
	}
	item, err := runtime.Decode(codec, bodyBytes)
	if err != nil {
		log.Fatalf("error decoding body bytes: %s", err)
	}
	name, err := accessor.Name(item)
	if err != nil {
		log.Fatalf("couldn't get object name: %s", err)
	}
	accessor.SetResourceVersion(item, "1")
	storage.set(ns, tipe, name, item)
	rw.WriteHeader(http.StatusCreated)
	bytes, err := runtime.Encode(codec, item)
	if err != nil {
		log.Fatalf("error encoding item: %s", err)
	}
	rw.Write(bytes)
}

func getItem(rw http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	tipe := mux.Vars(r)["type"]
	name := mux.Vars(r)["name"]
	item := storage.get(ns, tipe, name)
	if item == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	codec, err := testapi.GetCodecForObject(item)
	if err != nil {
		log.Fatalf("error getting codec: %s", err)
	}
	bytes, err := runtime.Encode(codec, item)
	if err != nil {
		log.Fatalf("error encoding item: %s", err)
	}
	rw.Write(bytes)
}

func updateItem(rw http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	tipe := mux.Vars(r)["type"]
	name := mux.Vars(r)["name"]
	origItem := storage.get(ns, tipe, name)
	if origItem == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	// TODO: Is there some type-agnostic way to get the codec?
	codec, err := testapi.GetCodecForObject(&v1alpha1.Broker{})
	if err != nil {
		log.Fatalf("error getting codec: %s", err)
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("error getting body bytes: %s", err)
	}
	item, err := runtime.Decode(codec, bodyBytes)
	if err != nil {
		log.Fatalf("error decoding body bytes: %s", err)
	}
	origResourceVersionStr, err := accessor.ResourceVersion(origItem)
	if err != nil {
		log.Fatalf("error getting resource version")
	}
	resourceVersionStr, err := accessor.ResourceVersion(item)
	if err != nil {
		log.Fatalf("error getting resource version")
	}
	// As with the actual core apiserver, "0" is a special resource version that
	// forces an update as if the current / most up-to-date resource version had
	// been passed in.
	if resourceVersionStr != "0" && resourceVersionStr != origResourceVersionStr {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	resourceVersion, err := strconv.Atoi(origResourceVersionStr)
	resourceVersion++
	accessor.SetResourceVersion(item, strconv.Itoa(resourceVersion))
	storage.set(ns, tipe, name, item)
	rw.WriteHeader(http.StatusCreated)
	bytes, err := runtime.Encode(codec, item)
	if err != nil {
		log.Fatalf("error encoding item: %s", err)
	}
	rw.Write(bytes)
}

func deleteItem(rw http.ResponseWriter, r *http.Request) {
	ns := mux.Vars(r)["namespace"]
	tipe := mux.Vars(r)["type"]
	name := mux.Vars(r)["name"]
	item := storage.get(ns, tipe, name)
	if item == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	storage.delete(ns, tipe, name)
	rw.WriteHeader(http.StatusOK)
}

func watchItem(rw http.ResponseWriter, r *http.Request) {

}

func notFoundHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusNotFound)
}
