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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "testns"
	name      = "testthing"
)

func TestCreate(t *testing.T) {
	broker := &v1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       ServiceBrokerKind.String(),
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	outBroker := &v1alpha1.Broker{}
	keyer := Keyer{
		DefaultNamespace: namespace,
		ResourceName:     ServiceBrokerKind.String(),
		Separator:        "/",
	}
	codec, err := testapi.GetCodecForObject(broker)
	fakeCl := newFakeCoreRESTClient()
	if err != nil {
		t.Fatalf("error getting codec (%s)", err)
	}
	iface := &store{
		decodeKey:    keyer.NamespaceAndNameFromKey,
		codec:        codec,
		cl:           fakeCl,
		singularKind: ServiceBrokerKind,
	}
	if err := iface.Create(
		context.Background(),
		name,
		broker,
		outBroker,
		uint64(0),
	); err != nil {
		t.Fatalf("error on create (%s)", err)
	}
	// compare basic attributes of Create's output broker to the input
	if outBroker.Name != broker.Name {
		t.Fatalf(
			"name of output broker (%s) didn't match input (%s)",
			outBroker.Name,
			broker.Name,
		)
	}
	if outBroker.Namespace != broker.Namespace {
		t.Fatalf(
			"namespace of output broker (%s) didn't match input (%s)",
			outBroker.Namespace,
			broker.Namespace,
		)
	}
	// compare what's in storage to what was passed in
	obj := fakeCl.storage.get(name, ServiceBrokerKind.URLName(), name)
	if obj == nil {
		t.Fatal("no broker was in storage")
	}
	name, err := fakeCl.accessor.Name(obj)
	if err != nil {
		t.Fatalf("couldn't get name from obj (%s)", err)
	}
	ns, err := fakeCl.accessor.Namespace(obj)
	if err != nil {
		t.Fatalf("couldn't get namespace from obj (%s)", err)
	}
	if name != broker.Name {
		t.Fatalf("name of broker-in-storage (%s) didn't match expected (%s)", name, broker.Name)
	}
	if ns != broker.Namespace {
		t.Fatalf("namespace of broker-in-storage (%s) didn't match expected (%s)", ns, broker.Namespace)
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
		t.Fatalf("couldn't remove namespace from object. it is still %s", obj.Namespace)
	}
}
