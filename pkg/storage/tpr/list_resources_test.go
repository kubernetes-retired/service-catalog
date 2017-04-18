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
	"testing"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStripNamespacesFromList(t *testing.T) {
	lst := sc.BrokerList{
		Items: []sc.Broker{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "testns1",
					Name:      "test1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "testns2",
					Name:      "test2",
				},
			},
		},
	}
	if err := stripNamespacesFromList(&lst); err != nil {
		t.Fatalf("removing namespaces from list (%s)", err)
	}
	for i, item := range lst.Items {
		if item.Namespace != "" {
			t.Errorf("item %d has a non-empty namespace %s", i, item.Namespace)
		}
	}
}

func TestGetAllNamespaces(t *testing.T) {
	const (
		ns1Name = "ns1"
	)
	cl := newFakeCoreRESTClient()
	nsList, err := getAllNamespaces(cl)
	if err != nil {
		t.Fatalf("getting all namespaces (%s)", err)
	}
	if len(nsList.Items) != 0 {
		t.Fatalf("expected 0 namespaces, got %d", len(nsList.Items))
	}
	cl.storage[ns1Name] = newTypedStorage()
	nsList, err = getAllNamespaces(cl)
	if err != nil {
		t.Fatalf("getting all namespaces (%s)", err)
	}
	if len(nsList.Items) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(nsList.Items))
	}
	if nsList.Items[0].Name != ns1Name {
		t.Fatalf("expected namespace with name %s, got %s instead", ns1Name, nsList.Items[0].Name)
	}
}

func TestListResource(t *testing.T) {
	const (
		ns   = "testns"
		kind = ServiceBrokerKind
	)

	cl := newFakeCoreRESTClient()
	listObj := sc.BrokerList{TypeMeta: newTypeMeta(kind)}
	codec, err := testapi.GetCodecForObject(&sc.BrokerList{TypeMeta: newTypeMeta(kind)})
	if err != nil {
		t.Fatalf("error getting codec (%s)", err)
	}
	objs, err := listResource(cl, ns, kind, &listObj, codec)
	if err != nil {
		t.Fatalf("error listing resource (%s)", err)
	}
	if len(objs) != 0 {
		t.Fatalf("expected 0 objects returned, got %d instead", len(objs))
	}
	objStorage := newObjStorage()
	objStorage["broker1"] = &sc.Broker{
		TypeMeta:   newTypeMeta(kind),
		ObjectMeta: metav1.ObjectMeta{Name: "broker1"},
	}
	objStorage["broker2"] = &sc.Broker{
		TypeMeta:   newTypeMeta(kind),
		ObjectMeta: metav1.ObjectMeta{Name: "broker2"},
	}
	cl.storage[ns] = newTypedStorage()
	cl.storage[ns][ServiceBrokerKind.URLName()] = objStorage
	objs, err = listResource(cl, ns, kind, &listObj, codec)
	if err != nil {
		t.Fatalf("error listing resource (%s)", err)
	}
	if len(objs) != len(objStorage) {
		t.Fatalf("expected %d objects returned, got %d instead", len(objStorage), len(objs))
	}
}
