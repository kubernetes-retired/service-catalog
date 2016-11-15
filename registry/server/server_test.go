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

	"github.com/kubernetes-incubator/service-catalog/model/service_broker"
	"github.com/kubernetes-incubator/service-catalog/registry/server"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {
	Describe("/services", func() {
		It("returns HTTP error on error", func() {
			handler := server.CreateHandler(&testController{
				listServices: func() ([]*model.Service, error) {
					return nil, errors.New("Cannot list services")
				},
			})

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest("GET", "/services", nil))
			Ω(rr.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns list of services", func() {
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
			Ω(rr.Code).To(Equal(http.StatusOK))

			// TODO: Validate more values
		})
	})

	// TODO: Add tests for all other API methods.
})

type testController struct {
	listServices  func() ([]*model.Service, error)
	getService    func(serviceID string) (*model.Service, error)
	createService func(s *model.Service) error
	deleteService func(serviceID string) error
}

func (c *testController) ListServices() ([]*model.Service, error) {
	Ω(c.listServices).ToNot(BeNil())
	return c.listServices()
}

func (c *testController) GetService(serviceID string) (*model.Service, error) {
	Ω(c.getService).ToNot(BeNil())
	return c.getService(serviceID)
}

func (c *testController) CreateService(s *model.Service) error {
	Ω(c.createService).ToNot(BeNil())
	return c.createService(s)
}

func (c *testController) DeleteService(serviceID string) error {
	Ω(c.deleteService).ToNot(BeNil())
	return c.deleteService(serviceID)
}
