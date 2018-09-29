/*
Copyright 2018 The Kubernetes Authors.

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

package controller_test

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"testing"
)

func TestBrokerClientManager_CreateBrokerClient(t *testing.T) {
	// GIVEN
	osbCl1, _ := osb.NewClient(testOsbConfig("osb-1"))
	osbCl2, _ := osb.NewClient(testOsbConfig("osb-2"))
	brokerClientFunc := clientFunc(osbCl1, osbCl2)
	manager := controller.NewBrokerClientManager(brokerClientFunc)

	// WHEN
	createdClient1, _ := manager.UpdateBrokerClient(controller.NewClusterServiceBrokerKey("broker1"), testOsbConfig("osb-1"))
	createdClient2, _ := manager.UpdateBrokerClient(controller.NewServiceBrokerKey("prod", "broker1"), testOsbConfig("osb-2"))
	gotClient1, exists1 := manager.BrokerClient(controller.NewClusterServiceBrokerKey("broker1"))
	gotClient2, exists2 := manager.BrokerClient(controller.NewServiceBrokerKey("prod", "broker1"))
	_, exists3 := manager.BrokerClient(controller.NewServiceBrokerKey("stage", "broker1"))

	// THEN
	if !exists1 {
		t.Fatal("Broker client osb-1 does not exists")
	}
	if !exists2 {
		t.Fatal("Broker client osb-2 does not exists")
	}
	if exists3 {
		t.Fatal("Broker client for namespace 'stage' must not exists")
	}

	if osbCl1 != createdClient1 {
		t.Fatalf("Wrong client from broker1")
	}
	if osbCl2 != createdClient2 {
		t.Fatalf("Wrong client from broker2")
	}

	if osbCl1 != gotClient1 {
		t.Fatalf("Wrong client from broker1")
	}
	if osbCl2 != gotClient2 {
		t.Fatalf("Wrong client from broker2")
	}
}

func TestBrokerClientManager_RemoveBrokerClient(t *testing.T) {
	// GIVEN
	osbCl1, _ := osb.NewClient(testOsbConfig("osb-1"))
	osbCl2, _ := osb.NewClient(testOsbConfig("osb-2"))
	brokerClientFunc := clientFunc(osbCl1, osbCl2)
	manager := controller.NewBrokerClientManager(brokerClientFunc)

	// WHEN
	manager.UpdateBrokerClient(controller.NewClusterServiceBrokerKey("broker1"), testOsbConfig("osb-1"))
	manager.UpdateBrokerClient(controller.NewServiceBrokerKey("prod", "broker1"), testOsbConfig("osb-2"))
	manager.RemoveBrokerClient(controller.NewClusterServiceBrokerKey("broker1"))
	_, exists1 := manager.BrokerClient(controller.NewClusterServiceBrokerKey("broker1"))
	_, exists2 := manager.BrokerClient(controller.NewServiceBrokerKey("prod", "broker1"))

	// THEN
	if exists1 {
		t.Fatal("Broker client for 'broker1' must not exists")
	}
	if !exists2 {
		t.Fatal("Broker client osb-2 does not exists")
	}
}

func TestBrokerClientManager_UpdateBrokerClient(t *testing.T) {
	// GIVEN
	osbCl1, _ := osb.NewClient(testOsbConfig("osb-1"))
	osbCl2, _ := osb.NewClient(testOsbConfig("osb-2"))
	osbCl3, _ := osb.NewClient(testOsbConfig("osb-3"))
	brokerClientFunc := clientFunc(osbCl1, osbCl2, osbCl3)
	manager := controller.NewBrokerClientManager(brokerClientFunc)

	osbCfg := testOsbConfig("osb-1")
	osbCfg.AuthConfig = &osb.AuthConfig{
		BasicAuthConfig: &osb.BasicAuthConfig{
			Username: "user-1",
			Password: "password-1",
		},
	}
	osbCfgWithPasswordChange := testOsbConfig("osb-1")
	osbCfgWithPasswordChange.AuthConfig = &osb.AuthConfig{
		BasicAuthConfig: &osb.BasicAuthConfig{
			Username: "user-1",
			Password: "password-changed",
		},
	}
	manager.UpdateBrokerClient(controller.NewClusterServiceBrokerKey("broker1"), osbCfg)
	manager.UpdateBrokerClient(controller.NewServiceBrokerKey("prod", "broker1"), testOsbConfig("osb-2"))

	// WHEN
	manager.UpdateBrokerClient(controller.NewClusterServiceBrokerKey("broker1"), osbCfgWithPasswordChange)

	// THEN
	gotClient, exists := manager.BrokerClient(controller.NewClusterServiceBrokerKey("broker1"))
	if !exists {
		t.Fatal("Broker client osb-2 does not exists")
	}
	if gotClient != osbCl3 {
		t.Fatalf("Broker client must have updated auth config")
	}
}

func clientFunc(clients ...osb.Client) osb.CreateFunc {
	var i = 0
	return func(_ *osb.ClientConfiguration) (osb.Client, error) {
		client := clients[i]
		i++
		return client, nil
	}
}

func testOsbConfig(name string) *osb.ClientConfiguration {
	return &osb.ClientConfiguration{
		Name: name,
	}
}
