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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/kubernetes-incubator/service-catalog/contrib/broker/controller"
	. "github.com/kubernetes-incubator/service-catalog/contrib/broker/server"
	brokerModel "github.com/kubernetes-incubator/service-catalog/model/service_broker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {

	Describe("/v2/catalog", func() {
		It("returns HTTP error on error", func() {
			handler := CreateHandler(&Controller{
				catalog: func() (*brokerModel.Catalog, error) {
					return nil, errors.New("Catalog retrieval error")
				},
			})

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest("GET", "/v2/catalog", nil))

			Ω(rr.Code).To(Equal(http.StatusBadRequest))
			Ω(rr.Header().Get("content-type")).To(Equal("application/json"))

			// TODO: This is a bug. We should be returning an error string.
			Ω(rr.Body.String()).To(Equal("{}"))
		})

		It("returns compliant JSON", func() {
			handler := CreateHandler(&Controller{
				catalog: func() (*brokerModel.Catalog, error) {
					return &brokerModel.Catalog{Services: []*brokerModel.Service{
						{
							Name: "foo",
						},
					}}, nil
				}})

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest("GET", "/v2/catalog", nil))
			Ω(rr.Code).To(Equal(http.StatusOK))
			Ω(rr.Header().Get("content-type")).To(Equal("application/json"))

			catalog := readJson(rr)

			Ω(catalog).Should(HaveLen(1))
			Ω(catalog).Should(HaveKey("services"))

			services := catalog["services"].([]interface{})
			Ω(services).ToNot(BeNil())

			var service map[string]interface{}
			service = services[0].(map[string]interface{})

			Ω(service).Should(HaveKey("name"))
			Ω(service["name"]).Should(Equal("foo"))
		})
	})
})

func readJson(rr *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &result)
	Ω(err).To(BeNil())
	return result
}

type Controller struct {
	catalog               func() (*brokerModel.Catalog, error)
	getServiceInstance    func(id string) (string, error)
	createServiceInstance func(id string, req *brokerModel.ServiceInstanceRequest) (*brokerModel.CreateServiceInstanceResponse, error)
	removeServiceInstance func(id string) error
	bind                  func(instanceId string, bindingId string, req *brokerModel.BindingRequest) (*brokerModel.CreateServiceBindingResponse, error)
	unBind                func(instanceId string, bindingId string) error
}

func (controller *Controller) Catalog() (*brokerModel.Catalog, error) {
	Ω(controller.catalog).ToNot(BeNil())
	return controller.catalog()
}

func (controller *Controller) GetServiceInstance(id string) (string, error) {
	Ω(controller.getServiceInstance).ToNot(BeNil())
	return controller.getServiceInstance(id)
}

func (controller *Controller) CreateServiceInstance(id string, req *brokerModel.ServiceInstanceRequest) (*brokerModel.CreateServiceInstanceResponse, error) {
	Ω(controller.createServiceInstance).ToNot(BeNil())
	return controller.createServiceInstance(id, req)
}

func (controller *Controller) RemoveServiceInstance(id string) error {
	Ω(controller.removeServiceInstance).ToNot(BeNil())
	return controller.removeServiceInstance(id)
}

func (controller *Controller) Bind(instanceId string, bindingId string, req *brokerModel.BindingRequest) (*brokerModel.CreateServiceBindingResponse, error) {
	Ω(controller.bind).ToNot(BeNil())
	return controller.bind(instanceId, bindingId, req)
}

func (controller *Controller) UnBind(instanceId string, bindingId string) error {
	Ω(controller.unBind).ToNot(BeNil())
	return controller.unBind(instanceId, bindingId)
}
