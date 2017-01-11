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

package openservicebroker

import (
	"net/http"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/openservicebroker/util"
)

func setup() (*util.FakeBrokerServer, *servicecatalog.Broker) {
	fbs := &util.FakeBrokerServer{}
	url := fbs.Start()
	fakeBroker := &servicecatalog.Broker{
		Spec: servicecatalog.BrokerSpec{
			URL: url,
		},
	}

	return fbs, fakeBroker
}

// Provision

func TestProvisionInstanceCreated(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusCreated)
	if _, err := c.CreateServiceInstance("1", &brokerapi.CreateServiceInstanceRequest{}); err != nil {
		t.Error(err.Error())
	}
}

func TestProvisionInstanceOK(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusOK)
	if _, err := c.CreateServiceInstance("1", &brokerapi.CreateServiceInstanceRequest{}); err != nil {
		t.Error(err.Error())
	}
}

func TestProvisionInstanceConflict(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusConflict)
	_, err := c.CreateServiceInstance("1", &brokerapi.CreateServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Errorf("Expected '%v'", errConflict)
	case err != errConflict:
		t.Errorf("Expected '%v', got '%v'", errConflict, err)
	}
}

func TestProvisionInstanceUnprocessableEntity(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusUnprocessableEntity)
	_, err := c.CreateServiceInstance("1", &brokerapi.CreateServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Errorf("Expected '%v'", errAsynchronous)
	case err != errAsynchronous:
		t.Errorf("Expected '%v', got '%v'", errAsynchronous, err)
	}
}

func TestProvisionInstanceAcceptedSuccessAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetAsynchronous(2, true, "succeed_async")
	req := brokerapi.CreateServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	if _, err := c.CreateServiceInstance("1", &req); err != nil {
		t.Error(err.Error())
	}
}

func TestProvisionInstanceAcceptedFailureAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetAsynchronous(2, false, "fail_async")
	req := brokerapi.CreateServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	_, err := c.CreateServiceInstance("1", &req)
	switch {
	case err == nil:
		t.Errorf("Expected '%v'", errFailedState)
	case err != errFailedState:
		t.Errorf("Expected '%v', got '%v'", errFailedState, err)
	}
}

// Deprovision

func TestDeprovisionInstanceOK(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusOK)
	if err := c.DeleteServiceInstance("1", &brokerapi.DeleteServiceInstanceRequest{}); err != nil {
		t.Error(err.Error())
	}
}

func TestDeprovisionInstanceGone(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusGone)
	if err := c.DeleteServiceInstance("1", &brokerapi.DeleteServiceInstanceRequest{}); err != nil {
		t.Error(err.Error())
	}
}

func TestDeprovisionInstanceUnprocessableEntity(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetResponseStatus(http.StatusUnprocessableEntity)
	err := c.DeleteServiceInstance("1", &brokerapi.DeleteServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Errorf("Expected '%v'", errAsynchronous)
	case err != errAsynchronous:
		t.Errorf("Expected '%v', got '%v'", errAsynchronous, err)
	}
}

func TestDeprovisionInstanceAcceptedSuccessAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetAsynchronous(2, true, "succeed_async")
	req := brokerapi.DeleteServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	if err := c.DeleteServiceInstance("1", &req); err != nil {
		t.Error(err.Error())
	}
}

func TestDeprovisionInstanceAcceptedFailureAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(fakeBroker)

	fbs.SetAsynchronous(2, false, "fail_async")
	req := brokerapi.DeleteServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	err := c.DeleteServiceInstance("1", &req)
	switch {
	case err == nil:
		t.Errorf("Expected '%v'", errFailedState)
	case err != errFailedState:
		t.Errorf("Expected '%v', got '%v'", errFailedState, err)
	}
}
