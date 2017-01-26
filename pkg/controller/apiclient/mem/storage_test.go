/*
Copyright 2016 The Kubernetes Authors.

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

package mem

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	kapi "k8s.io/kubernetes/pkg/api"
)

const (
	brokerName1 = "Test1"
)

func TestNoBrokers(t *testing.T) {
	const bogusBrokerName = "NOT THERE"
	bs := NewAPIClient().Brokers()
	l, err := bs.List()
	if err != nil {
		t.Fatalf("List failed with: %s", err)
	}
	if len(l) != 0 {
		t.Fatalf("Expected 0 brokers, got %d", len(l))
	}
	b, err := bs.Get(bogusBrokerName)
	if err == nil {
		t.Fatal("Get did not fail")
	}
	if b != nil {
		t.Fatalf("Got back a broker: %#v", b)
	}
}

func TestAddBroker(t *testing.T) {
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Create(b)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err := bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b2 == b {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b2, b) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
}

func TestAddDuplicateBroker(t *testing.T) {
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Create(b)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err := bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b2 == b {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b2, b) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
	_, err = bs.Create(b)
	if err == nil {
		t.Fatal("Create did not fail with duplicate")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Unexpected error, wanted 'already exists' but got %s", err)
	}
}

func TestUpdateBroker(t *testing.T) {
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Create(b)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err := bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b2 == b {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b2, b) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
	b3 := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err = bs.Update(b3)
	if err != nil {
		t.Fatalf("Update failed with: %s", err)
	}
	l, err = bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b4, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b4 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b4 == b3 {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b4, b3) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
}

func TestUpdateNonExistentBroker(t *testing.T) {
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Update(b)
	if err == nil {
		t.Fatal("Update didn't fail for broker that does not exist")
	}
	if !strings.Contains(err.Error(), "no such broker") {
		t.Fatalf("Unexpected error, wanted 'no such broker' but got %s", err)
	}
}

func TestDeleteBroker(t *testing.T) {
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Create(b)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err := bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b2 == b {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b2, b) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
	err = bs.Delete(brokerName1)
	if err != nil {
		t.Fatalf("Failed to delete broker: %s : %s", brokerName1, err)
	}
	l, err = bs.List()
	if len(l) != 0 {
		t.Fatalf("Expected 0 broker, got %d", len(l))
	}
	b3, err := bs.Get(brokerName1)
	if err == nil {
		t.Fatal("Get returned a broker when there should be none")
	}
	if b3 != nil {
		t.Fatalf("Got back a broker: %#v", b3)
	}
}

func TestDeleteBrokerMultiple(t *testing.T) {
	const brokerName2 = "Test2"
	bs := NewAPIClient().Brokers()
	b := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName1}}
	_, err := bs.Create(b)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err := bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := bs.Get(brokerName1)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b2 == b {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b2, b) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
	b3 := &servicecatalog.Broker{ObjectMeta: kapi.ObjectMeta{Name: brokerName2}}
	_, err = bs.Create(b3)
	if err != nil {
		t.Fatalf("Create failed with: %s", err)
	}
	l, err = bs.List()
	if len(l) != 2 {
		t.Fatalf("Expected 2 brokers, got %d", len(l))
	}
	b4, err := bs.Get(brokerName2)
	if err != nil {
		t.Fatalf("Get failed: %s", err)
	}
	if b4 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b4 == b3 {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b4, b3) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
	err = bs.Delete(brokerName1)
	if err != nil {
		t.Fatalf("Failed to delete broker: %s : %s", brokerName1, err)
	}
	l, err = bs.List()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b5, err := bs.Get(brokerName1)
	if err == nil {
		t.Fatal("Get returned a broker when there should be none")
	}
	if b5 != nil {
		t.Fatalf("Got back a broker: %#v", b5)
	}
	b6, err := bs.Get(brokerName2)
	if err != nil {
		t.Fatal("Get failed for entry that should be there")
	}
	if b6 == nil {
		t.Fatal("Did not get back a broker")
	}
	if b6 == b3 {
		t.Fatal("Broker addresses match; expected them not to")
	}
	if !reflect.DeepEqual(b6, b3) {
		t.Fatal("Brokers are not deeply equal; expected them to be")
	}
}
