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
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apiserver/pkg/endpoints/request"
	restclient "k8s.io/client-go/rest"
)

const (
	namespace = "testns"
	name      = "testthing"
)

func TestCreate(t *testing.T) {
	keyer := getKeyer()
	fakeCl := newFakeCoreRESTClient()
	iface := getTPRStorageIFace(t, keyer, fakeCl)
	broker := &v1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       ServiceBrokerKind.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	key, err := keyer.Key(request.NewContext(), name)
	if err != nil {
		t.Fatalf("error constructing key (%s)", err)
	}
	outBroker := &v1alpha1.Broker{}
	if err := iface.Create(
		context.Background(),
		key,
		broker,
		outBroker,
		uint64(0),
	); err != nil {
		t.Fatalf("error on create (%s)", err)
	}
	// Confirm resource version got set during the create operation
	if outBroker.ResourceVersion == "" {
		t.Fatalf("resource version was not set as expected")
	}
	// Confirm the output is identical to what is in storage (nothing funny
	// happened during encoding / decoding the response).
	obj := fakeCl.storage.get(namespace, ServiceBrokerKind.URLName(), name)
	if obj == nil {
		t.Fatal("no broker was in storage")
	}
	if !reflect.DeepEqual(outBroker, obj) {
		t.Fatalf(
			"output and object in storage are different: %s",
			diff.ObjectReflectDiff(outBroker, obj),
		)
	}
	// Output and what's in storage should be known to be deeply equal at this
	// point. Compare either of those to what was passed in. The only diff should
	// be resource version, so we will clear that first.
	outBroker.ResourceVersion = ""
	if !reflect.DeepEqual(broker, outBroker) {
		t.Fatalf(
			"input and output are different: %s",
			diff.ObjectReflectDiff(broker, outBroker),
		)
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
	codec, err := testapi.GetCodecForObject(&v1alpha1.Broker{})
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
