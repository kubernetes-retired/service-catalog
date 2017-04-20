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

package tpr

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage"
	restclient "k8s.io/client-go/rest"
)

const (
	namespace = "testns"
	name      = "testthing"
)

func TestCreateAndRead(t *testing.T) {
	keyer := getKeyer()
	fakeCl := newFakeCoreRESTClient()
	iface := getTPRStorageIFace(t, keyer, fakeCl)
	origBroker := &sc.Broker{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: sc.BrokerSpec{
			URL: "http://my-awesome-broker.io",
		},
	}
	key, err := keyer.Key(request.NewContext(), name)
	if err != nil {
		t.Fatalf("error constructing key (%s)", err)
	}

	// Begin create test
	createdBroker := &sc.Broker{}
	if err := iface.Create(
		context.Background(),
		key,
		origBroker,
		createdBroker,
		uint64(0),
	); err != nil {
		t.Fatalf("error on create (%s)", err)
	}
	// Confirm resource version got set during the create operation
	if createdBroker.ResourceVersion == "" {
		t.Fatalf("resource version was not set as expected")
	}
	// Confirm the output is identical to what is in storage (nothing funny
	// happened during encoding / decoding the response).
	// the object will have been encoded by the test coder as an unversioned
	// servicecatalog.Broker type
	obj := fakeCl.storage.get(namespace, ServiceBrokerKind.URLName(), name)
	if obj == nil {
		t.Fatal("no broker was in storage")
	}
	err = deepCompare("output", createdBroker, "object in storage", obj)
	if err != nil {
		t.Fatal(err)
	}
	// Output and what's in storage should be known to be deeply equal at this
	// point. Compare either of those to what was passed in. The only diff should
	// be resource version, so we will set that first.
	origBroker.ResourceVersion = createdBroker.ResourceVersion
	err = deepCompare("input", origBroker, "output", createdBroker)
	if err != nil {
		t.Fatal(err)
	}
	// End create test

	// Begin create existing test
	origBroker2 := &sc.Broker{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	createdBroker2 := &sc.Broker{}
	err = iface.Create(
		context.Background(),
		key,
		origBroker2,
		createdBroker2,
		uint64(0),
	)
	if err = verifyStorageError(err, storage.ErrCodeKeyExists); err != nil {
		t.Fatal(err)
	}
	// End create existing test

	// Begin get test
	gottenBroker := &sc.Broker{}
	if err := iface.Get(
		context.Background(),
		key,
		"", // TODO: Current impl ignores resource version-- may be wrong
		gottenBroker,
		false, // Do not ignore if not found; error instead
	); err != nil {
		t.Fatalf("error getting object (%s)", err)
	}
	// Retrieved object should be deeply equal to what we created earlier
	if err = deepCompare(
		"retrieved object",
		gottenBroker,
		"expected object",
		createdBroker,
	); err != nil {
		t.Fatal(err)
	}
	// End get test

	// Begin list test
	gottenList := &sc.BrokerList{}
	if err := iface.List(
		context.Background(),
		keyer.KeyRoot(request.NewContext()),
		"", // TODO: Current impl ignores resource version-- may be wrong
		// TODO: Current impl ignores selection predicate-- may be wrong
		storage.SelectionPredicate{},
		gottenList,
	); err != nil {
		t.Fatalf("error listing objects (%s)", err)
	}
	// List should contain precisely one item
	if len(gottenList.Items) != 1 {
		t.Fatalf(
			"expected list to contain exactly one item, but got %d items",
			len(gottenList.Items),
		)
	}
	// That one list item should be deeply equal to the one we retrieved earlier
	if err = deepCompare(
		"retrieved list item",
		&gottenList.Items[0],
		"expected object",
		gottenBroker,
	); err != nil {
		t.Fatal(err)
	}
	// End list test
}

func TestGetEmptyList(t *testing.T) {
	keyer := getKeyer()
	fakeCl := newFakeCoreRESTClient()
	iface := getTPRStorageIFace(t, keyer, fakeCl)
	key := keyer.KeyRoot(request.NewContext())
	outBrokerList := &sc.BrokerList{}
	if err := iface.List(
		context.Background(),
		key,
		"", // TODO: Current impl ignores resource version-- may be wrong
		// TODO: Current impl ignores selection predicate-- may be wrong
		storage.SelectionPredicate{},
		outBrokerList,
	); err != nil {
		t.Fatalf("error listing objects (%s)", err)
	}
	if len(outBrokerList.Items) != 0 {
		t.Fatalf(
			"expected an empty list, but got %d items",
			len(outBrokerList.Items),
		)
	}
	// Repeat using GetToList
	if err := iface.GetToList(
		context.Background(),
		key,
		"", // TODO: Current impl ignores resource version-- may be wrong
		// TODO: Current impl ignores selection predicate-- may be wrong
		storage.SelectionPredicate{},
		outBrokerList,
	); err != nil {
		t.Fatalf("error listing objects (%s)", err)
	}
	if len(outBrokerList.Items) != 0 {
		t.Fatalf(
			"expected an empty list, but got %d items",
			len(outBrokerList.Items),
		)
	}
}

func TestGetNonExistent(t *testing.T) {
	keyer := getKeyer()
	fakeCl := newFakeCoreRESTClient()
	iface := getTPRStorageIFace(t, keyer, fakeCl)
	key, err := keyer.Key(request.NewContext(), name)
	if err != nil {
		t.Fatalf("error constructing key (%s)", err)
	}
	outBroker := &sc.Broker{}
	// Ignore not found
	if err := iface.Get(
		context.Background(),
		key,
		"", // TODO: Current impl ignores resource version-- may be wrong
		outBroker,
		true,
	); err != nil {
		t.Fatalf("expected no error, but received one (%s)", err)
	}
	// Object should remain unmodified-- i.e. deeply equal to a new broker
	err = deepCompare("output", outBroker, "new broker", &sc.Broker{})
	if err != nil {
		t.Fatal(err)
	}
	// Do not ignore not found
	err = iface.Get(
		context.Background(),
		key,
		"", // TODO: Current impl ignores resource version-- may be wrong
		outBroker,
		false,
	)
	if err = verifyStorageError(err, storage.ErrCodeKeyNotFound); err != nil {
		t.Fatal(err)
	}
	// Object should remain unmodified-- i.e. deeply equal to a new broker
	err = deepCompare("output", outBroker, "new broker", &sc.Broker{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteNonExistent(t *testing.T) {
	keyer := getKeyer()
	fakeCl := newFakeCoreRESTClient()
	iface := getTPRStorageIFace(t, keyer, fakeCl)
	key, err := keyer.Key(request.NewContext(), name)
	if err != nil {
		t.Fatalf("error constructing key (%s)", err)
	}
	outBroker := &sc.Broker{}
	// Ignore not found
	err = iface.Delete(
		context.Background(),
		key,
		outBroker,
		nil, // TODO: Current impl ignores preconditions-- may be wrong
	)
	if err = verifyStorageError(err, storage.ErrCodeKeyNotFound); err != nil {
		t.Fatal(err)
	}
	// Object should remain unmodified-- i.e. deeply equal to a new broker
	err = deepCompare("output", outBroker, "new broker", &sc.Broker{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveNamespace(t *testing.T) {
	obj := &servicecatalog.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testns",
		},
	}
	if err := removeNamespace(obj); err != nil {
		t.Fatalf("couldn't remove namespace (%s", err)
	}
	if obj.Namespace != "" {
		t.Fatalf(
			"couldn't remove namespace from object. it is still %s",
			obj.Namespace,
		)
	}
}

func getKeyer() Keyer {
	return Keyer{
		DefaultNamespace: namespace,
		ResourceName:     ServiceBrokerKind.String(),
		Separator:        "/",
	}
}

func getTPRStorageIFace(t *testing.T, keyer Keyer, restCl restclient.Interface) *store {
	codec, err := testapi.GetCodecForObject(&sc.Broker{})
	if err != nil {
		t.Fatalf("error getting codec (%s)", err)
	}
	return &store{
		decodeKey:    keyer.NamespaceAndNameFromKey,
		codec:        codec,
		cl:           restCl,
		singularKind: ServiceBrokerKind,
	}
}

func verifyStorageError(err error, errorCode int) error {
	if err == nil {
		return errors.New("expected an error, but did not receive one")
	}
	storageErr, ok := err.(*storage.StorageError)
	if !ok {
		return fmt.Errorf(
			"expected a storage.StorageError, but got a %s",
			reflect.TypeOf(err),
		)
	}
	if storageErr.Code != errorCode {
		return fmt.Errorf(
			"expected error code %d, but got %d",
			errorCode,
			storageErr.Code,
		)
	}
	return nil
}

func deepCompare(
	obj1Name string,
	obj1 runtime.Object,
	obj2Name string,
	obj2 runtime.Object,
) error {
	if !reflect.DeepEqual(obj1, obj2) {
		return fmt.Errorf(
			"%s and %s are different: %s",
			obj1Name,
			obj2Name,
			diff.ObjectReflectDiff(obj1, obj2),
		)
	}
	return nil
}
