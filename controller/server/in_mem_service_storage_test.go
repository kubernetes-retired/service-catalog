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

package server

import (
	"strings"
	"testing"

	model "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

const (
	brokerOneUUID = "126b8154-a24a-4e79-9185-3df2eb4d18a8"
	brokerTwoUUID = "2b0c42ed-c43a-4724-b883-e5ba878a8bfd"
)

func TestNoBrokers(t *testing.T) {
	s := CreateInMemServiceStorage()
	l, err := s.ListBrokers()
	if err != nil {
		t.Fatalf("ListBrokers failed with: %#v", err)
	}
	if len(l) != 0 {
		t.Fatalf("Expected 0 brokers, got %d", len(l))
	}
	b, err := s.GetBroker("NOT THERE")
	if err == nil {
		t.Fatal("GetBroker did not fail")
	}
	if b != nil {
		t.Fatalf("Got back a broker: %#v", b)
	}
}

func TestAddBroker(t *testing.T) {
	s := CreateInMemServiceStorage()
	b := &model.ServiceBroker{GUID: "Test"}
	cat := model.Catalog{
		Services: []*model.Service{},
	}
	err := s.AddBroker(b, &cat)
	if err != nil {
		t.Fatalf("AddBroker failed with: %#v", err)
	}
	l, err := s.ListBrokers()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := s.GetBroker("Test")
	if err != nil {
		t.Fatalf("GetBroker failed: %#v", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(b2.Name, b.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b.Name, b2.Name)
	}
}

func TestAddDuplicateBroker(t *testing.T) {
	s := CreateInMemServiceStorage()
	b := &model.ServiceBroker{GUID: "Test"}
	cat := model.Catalog{
		Services: []*model.Service{},
	}
	err := s.AddBroker(b, &cat)
	if err != nil {
		t.Fatalf("AddBroker failed with: %#v", err)
	}
	l, err := s.ListBrokers()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := s.GetBroker("Test")
	if err != nil {
		t.Fatalf("GetBroker failed: %#v", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(b2.Name, b.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b.Name, b2.Name)
	}
	err = s.AddBroker(b, &cat)
	if err == nil {
		t.Fatal("AddBroker did not fail with duplicate")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Unexpected error, wanted 'already exists' but got %#v", err)
	}
}

func TestUpdateBroker(t *testing.T) {
	s := CreateInMemServiceStorage()
	b1 := &model.ServiceBroker{GUID: "Test", Name: "Old"}
	cat1 := model.Catalog{
		Services: []*model.Service{
			{
				ID:   "s1",
				Name: "same service",
			},
			{
				ID:   "s2",
				Name: "old service",
			},
		},
	}
	s.AddBroker(b1, &cat1)

	b2 := &model.ServiceBroker{GUID: "Test", Name: "New"}
	cat2 := model.Catalog{
		Services: []*model.Service{
			{
				ID:   "s1",
				Name: "same service",
			},
			{
				ID:   "s3",
				Name: "new service",
			},
			{
				ID:   "s4",
				Name: "extra service",
			},
		},
	}
	s.UpdateBroker(b2, &cat2)

	l, err := s.ListBrokers()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b3, err := s.GetBroker("Test")
	if err != nil {
		t.Fatalf("GetBroker failed: %#v", err)
	}
	if b3 == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(b3.Name, b2.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b2.Name, b3.Name)
	}
	cat3, err := s.GetInventory()
	if err != nil {
		t.Fatalf("GetInventory failed: %#v", err)
	}
	if len(cat3.Services) != len(cat2.Services) {
		t.Fatalf("Catalog sizes do not match, expected: '%+v', got '%+v'", len(cat2.Services), len(cat3.Services))
	}
	for i, s3 := range cat3.Services {
		if strings.Compare(s3.Name, cat2.Services[i].Name) != 0 {
			t.Fatalf("Catalogs entries do not match, expected: '%+v', got '%+v'", cat2, cat3)
		}
	}
}

func TestUpdateNonExistentBroker(t *testing.T) {
	s := CreateInMemServiceStorage()
	b := &model.ServiceBroker{GUID: "Test"}
	cat := model.Catalog{
		Services: []*model.Service{},
	}

	err := s.UpdateBroker(b, &cat)

	if err == nil {
		t.Fatal("UpdateBroker did not fail with duplicate")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("Unexpected error, wanted 'does not exist' but got %#v", err)
	}
}

func TestDeleteBroker(t *testing.T) {
	s := CreateInMemServiceStorage()
	b := &model.ServiceBroker{GUID: brokerOneUUID}
	cat := model.Catalog{
		Services: []*model.Service{},
	}
	err := s.AddBroker(b, &cat)
	if err != nil {
		t.Fatalf("AddBroker failed with: %#v", err)
	}
	l, err := s.ListBrokers()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	b2, err := s.GetBroker(brokerOneUUID)
	if err != nil {
		t.Fatalf("GetBroker failed: %#v", err)
	}
	if b2 == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(b2.Name, b.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b.Name, b2.Name)
	}
	err = s.DeleteBroker(brokerOneUUID)
	if err != nil {
		t.Fatalf("Failed to delete broker: %s : %#v", brokerOneUUID, err)
	}
	l, err = s.ListBrokers()
	if len(l) != 0 {
		t.Fatalf("Expected 0 broker, got %d", len(l))
	}
	b2, err = s.GetBroker(brokerOneUUID)
	if err == nil {
		t.Fatal("GetBroker returned a broker when there should be none")
	}
}

func TestDeleteBrokerMultiple(t *testing.T) {
	s := CreateInMemServiceStorage()
	b := &model.ServiceBroker{GUID: brokerOneUUID}
	b2 := &model.ServiceBroker{GUID: brokerTwoUUID}
	cat := model.Catalog{
		Services: []*model.Service{{Name: "first"}},
	}
	cat2 := model.Catalog{
		Services: []*model.Service{{Name: "second"}},
	}
	err := s.AddBroker(b, &cat)
	if err != nil {
		t.Fatalf("AddBroker failed with: %#v", err)
	}
	err = s.AddBroker(b2, &cat2)
	if err != nil {
		t.Fatalf("AddBroker failed with: %#v", err)
	}
	l, err := s.ListBrokers()
	if len(l) != 2 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	bRet, err := s.GetBroker(brokerOneUUID)
	if err != nil {
		t.Fatalf("GetBroker failed: %#v", err)
	}
	if bRet == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(bRet.Name, b.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b.Name, bRet.Name)
	}
	catRet, err := s.GetInventory()
	if err != nil {
		t.Fatalf("Failed to get inventory: %#v", err)
	}
	if len(catRet.Services) != 2 {
		t.Fatalf("Expected 2 services from GetInventory, got %d ", len(catRet.Services))
	}

	err = s.DeleteBroker(brokerOneUUID)
	if err != nil {
		t.Fatalf("Failed to delete broker: %s : %#v", brokerOneUUID, err)
	}
	l, err = s.ListBrokers()
	if len(l) != 1 {
		t.Fatalf("Expected 1 broker, got %d", len(l))
	}
	bRet, err = s.GetBroker(brokerOneUUID)
	if err == nil {
		t.Fatal("GetBroker returned a broker when there should be none")
	}
	bRet, err = s.GetBroker(brokerTwoUUID)
	if err != nil {
		t.Fatal("GetBroker failed for entry that should be there")
	}

	if bRet == nil {
		t.Fatal("Did not get back a broker")
	}
	if strings.Compare(bRet.Name, b2.Name) != 0 {
		t.Fatalf("Names don't match, expected: '%s', got '%s'", b2.Name, bRet.Name)
	}
	catRet, err = s.GetInventory()
	if err != nil {
		t.Fatalf("Failed to get inventory: %#v", err)
	}
	if len(catRet.Services) != 1 {
		t.Fatalf("Expected 1 service from GetInventory, got %d ", len(catRet.Services))
	}
}
