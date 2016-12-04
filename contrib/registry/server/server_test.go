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

package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/contrib/registry/server"
	"github.com/kubernetes-incubator/service-catalog/model/service_broker"
)

//
// Tests of the registry /services endpoint.
//

// Registry /services returns HTTP error on error.
func TestRegistryReturnsHTTPErrorOnError(t *testing.T) {

	handler := server.CreateHandler(&testController{
		listServices: func() ([]*model.Service, error) {
			return nil, errors.New("Cannot list services")
		},
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/services", nil))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected HTTP status http.StatusBadRequest (%d), got %d", http.StatusBadRequest, rr.Code)
	}
}

// Registry /services returns a valid list of services with success.
func TestRegistryReturnsListOfServices(t *testing.T) {
	handler := server.CreateHandler(&testController{
		listServices: func() ([]*model.Service, error) {
			return []*model.Service{
				{
					Name:        "backend",
					ID:          "backend-id",
					Description: "Backend Service",
				},
				{
					Name:        "frontend",
					ID:          "frontend-id",
					Description: "Frontend Service",
				},
			}, nil
		},
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/services", nil))

	if rr.Code != http.StatusOK {
		t.Errorf("Expected HTTP status http.StatusOK (%d), got %d", http.StatusOK, rr.Code)
	}

	// TODO: Validate more values
}

// TODO: Add tests for all other API methods.

type testController struct {
	t *testing.T
	listServices  func() ([]*model.Service, error)
	getService    func(serviceID string) (*model.Service, error)
	createService func(s *model.Service) error
	deleteService func(serviceID string) error
}

func (c *testController) ListServices() ([]*model.Service, error) {
	if c.listServices == nil {
		c.t.Error("Test failed to provide 'listServices' handler")
	}
	return c.listServices()
}

func (c *testController) GetService(serviceID string) (*model.Service, error) {
	if c.getService == nil {
		c.t.Error("Test failed to provide 'getService' handler")
	}
	return c.getService(serviceID)
}

func (c *testController) CreateService(s *model.Service) error {
	if c.createService == nil {
		c.t.Error("Test failed to provide 'createService' handler")
	}
	return c.createService(s)
}

func (c *testController) DeleteService(serviceID string) error {
	if c.deleteService == nil {
		c.t.Error("Test failed to provide 'deleteService' handler")
	}
	return c.deleteService(serviceID)
}
