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
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fakeserver"
)

const (
	testBrokerName            = "test-broker"
	bindingSuffixFormatString = "/v2/service_instances/%s/service_bindings/%s"
	testServiceInstanceID     = "1"
	testServiceBindingID      = "2"
)

func setup() (*fakeserver.FakeBrokerServer, *servicecatalog.Broker) {
	fbs := &fakeserver.FakeBrokerServer{}
	url := fbs.Start()
	fakeBroker := &servicecatalog.Broker{
		Spec: servicecatalog.BrokerSpec{
			URL: url,
		},
	}

	return fbs, fakeBroker
}

// Catalog tests

func TestCatalog(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	fbs.Catalog = &brokerapi.Catalog{}

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")
	if _, err := c.GetCatalog(); err != nil {
		t.Fatal(err.Error())
	}
}

// Provision tests

func TestProvisionInstanceCreated(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusCreated)
	if _, err := c.CreateServiceInstance(testServiceInstanceID, &brokerapi.CreateServiceInstanceRequest{}); err != nil {
		t.Fatal(err.Error())
	}
}

func TestProvisionInstanceOK(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusOK)
	if _, err := c.CreateServiceInstance(testServiceInstanceID, &brokerapi.CreateServiceInstanceRequest{}); err != nil {
		t.Fatal(err.Error())
	}
}

func TestProvisionInstanceConflict(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusConflict)
	_, err := c.CreateServiceInstance(testServiceInstanceID, &brokerapi.CreateServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Fatalf("Expected '%v'", errConflict)
	case err != errConflict:
		t.Fatalf("Expected '%v', got '%v'", errConflict, err)
	}
}

func TestProvisionInstanceUnprocessableEntity(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusUnprocessableEntity)
	_, err := c.CreateServiceInstance(testServiceInstanceID, &brokerapi.CreateServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Fatalf("Expected '%v'", errAsynchronous)
	case err != errAsynchronous:
		t.Fatalf("Expected '%v', got '%v'", errAsynchronous, err)
	}
}

func TestProvisionInstanceAcceptedSuccessAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetAsynchronous(2, true, "succeed_async")
	req := brokerapi.CreateServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	if _, err := c.CreateServiceInstance(testServiceInstanceID, &req); err != nil {
		t.Fatal(err.Error())
	}
}

func TestProvisionInstanceAcceptedFailureAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetAsynchronous(2, false, "fail_async")
	req := brokerapi.CreateServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	_, err := c.CreateServiceInstance(testServiceInstanceID, &req)
	switch {
	case err == nil:
		t.Fatalf("Expected '%v'", errFailedState)
	case err != errFailedState:
		t.Fatalf("Expected '%v', got '%v'", errFailedState, err)
	}
}

// Deprovision

func TestDeprovisionInstanceOK(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusOK)
	if err := c.DeleteServiceInstance(testServiceInstanceID, &brokerapi.DeleteServiceInstanceRequest{}); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDeprovisionInstanceGone(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusGone)
	if err := c.DeleteServiceInstance(testServiceInstanceID, &brokerapi.DeleteServiceInstanceRequest{}); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDeprovisionInstanceUnprocessableEntity(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusUnprocessableEntity)
	err := c.DeleteServiceInstance(testServiceInstanceID, &brokerapi.DeleteServiceInstanceRequest{})
	switch {
	case err == nil:
		t.Fatalf("Expected '%v'", errAsynchronous)
	case err != errAsynchronous:
		t.Fatalf("Expected '%v', got '%v'", errAsynchronous, err)
	}
}

func TestDeprovisionInstanceAcceptedSuccessAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetAsynchronous(2, true, "succeed_async")
	req := brokerapi.DeleteServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	if err := c.DeleteServiceInstance(testServiceInstanceID, &req); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDeprovisionInstanceAcceptedFailureAsynchronous(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetAsynchronous(2, false, "fail_async")
	req := brokerapi.DeleteServiceInstanceRequest{
		AcceptsIncomplete: true,
	}

	err := c.DeleteServiceInstance(testServiceInstanceID, &req)
	switch {
	case err == nil:
		t.Fatalf("Expected '%v'", errFailedState)
	case err != errFailedState:
		t.Fatalf("Expected '%v', got '%v'", errFailedState, err)
	}
}

// Bind tests

func TestBindOk(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusOK)
	sent := &brokerapi.BindingRequest{}
	if _, err := c.CreateServiceBinding(testServiceInstanceID, testServiceBindingID, sent); err != nil {
		t.Fatal(err.Error())
	}

	verifyBindingMethodAndPath(http.MethodPut, testServiceInstanceID, testServiceBindingID, fbs.Request, t)

	if fbs.RequestObject == nil {
		t.Fatalf("BindingRequest was not received correctly")
	}
	actual := reflect.TypeOf(fbs.RequestObject)
	expected := reflect.TypeOf(&brokerapi.BindingRequest{})
	if actual != expected {
		t.Fatalf("Got the wrong type for the request, expected %v got %v", expected, actual)
	}
	received := fbs.RequestObject.(*brokerapi.BindingRequest)
	if !reflect.DeepEqual(*received, *sent) {
		t.Fatalf("Sent does not match received, sent: %+v received: %+v", sent, received)
	}
}

func TestBindConflict(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusConflict)
	sent := &brokerapi.BindingRequest{}
	if _, err := c.CreateServiceBinding(testServiceInstanceID, testServiceBindingID, sent); err == nil {
		t.Fatal("Expected create service binding to fail with conflict, but didn't")
	}

	verifyBindingMethodAndPath(http.MethodPut, testServiceInstanceID, testServiceBindingID, fbs.Request, t)

	if fbs.RequestObject == nil {
		t.Fatalf("BindingRequest was not received correctly")
	}
	actual := reflect.TypeOf(fbs.RequestObject)
	expected := reflect.TypeOf(&brokerapi.BindingRequest{})
	if actual != expected {
		t.Fatalf("Got the wrong type for the request, expected %v got %v", expected, actual)
	}
	received := fbs.RequestObject.(*brokerapi.BindingRequest)
	if !reflect.DeepEqual(*received, *sent) {
		t.Fatalf("Sent does not match received, sent: %+v received: %+v", sent, received)
	}
}

// Unbind tests

func TestUnbindOk(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusOK)
	if err := c.DeleteServiceBinding(testServiceInstanceID, testServiceBindingID); err != nil {
		t.Fatal(err.Error())
	}

	verifyBindingMethodAndPath(http.MethodDelete, testServiceInstanceID, testServiceBindingID, fbs.Request, t)

	if fbs.Request.ContentLength != 0 {
		t.Fatalf("not expecting a request body, but got one, size %d", fbs.Request.ContentLength)
	}
}

func TestUnbindGone(t *testing.T) {
	fbs, fakeBroker := setup()
	defer fbs.Stop()

	c := NewClient(testBrokerName, fakeBroker.Spec.URL, "", "")

	fbs.SetResponseStatus(http.StatusGone)
	err := c.DeleteServiceBinding(testServiceInstanceID, testServiceBindingID)
	if err == nil {
		t.Fatal("Expected delete service binding to fail with gone, but didn't")
	}
	if !strings.Contains(err.Error(), "There is no binding") {
		t.Fatalf("Did not find the expected error message 'There is no binding' in error: %s", err)
	}

	verifyBindingMethodAndPath(http.MethodDelete, testServiceInstanceID, testServiceBindingID, fbs.Request, t)
}

// verifyBindingMethodAndPath is a helper method that verifies that the request
// has the right method and the suffix URL for a binding request.
func verifyBindingMethodAndPath(method, serviceID, bindingID string, req *http.Request, t *testing.T) {
	if req.Method != method {
		t.Fatalf("Expected method to use %s but was %s", method, req.Method)
	}
	expectPath := fmt.Sprintf(bindingSuffixFormatString, serviceID, bindingID)
	if !strings.HasSuffix(req.URL.Path, expectPath) {
		t.Fatalf("Expected binding create path to have suffix %s but was: %s", expectPath, req.URL.Path)
	}

}
