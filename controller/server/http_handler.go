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
	"log"
	"net/http"
	"time"

	"github.com/kubernetes-incubator/service-catalog/controller/utility"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"

	"github.com/satori/go.uuid"
)

// httpHandler handles service controller HTTP requests.
type httpHandler struct {
	controller ServiceController
	k8sStorage ServiceStorage
}

// NewHTTPHandler creates an instance of the HttpHandler object.
func newHTTPHandler(c ServiceController, k8sStorage ServiceStorage) *httpHandler {
	return &httpHandler{
		controller: c,
		k8sStorage: k8sStorage,
	}
}

// CreateServiceInstance handles the 'create service instance' API request.
func (h *httpHandler) CreateServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating Service Instance")

	var req scmodel.CreateServiceInstanceRequest
	err := utils.BodyToObject(r, &req)
	if err != nil {
		log.Printf("Error unmarshaling: %v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	instance := scmodel.ServiceInstance{
		Name:             req.Name,
		Service:          req.Service,
		Plan:             req.Plan,
		OrganizationGuid: req.OrgID,
		Parameters:       req.Parameters,
	}

	err = h.k8sStorage.AddService(&instance)
	if err != nil {
		log.Printf("Error creating a service instance: %v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, instance)
}

// CreateServiceBinding handles the 'create service binding' API request.
func (h *httpHandler) CreateServiceBinding(w http.ResponseWriter, r *http.Request) {
	var req scmodel.CreateServiceBindingRequest
	err := utils.BodyToObject(r, &req)
	if err != nil {
		log.Printf("Error unmarshaling: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	binding := scmodel.ServiceBinding{
		Name:       req.Name,
		From:       req.From,
		To:         req.To,
		Parameters: req.Parameters,
	}

	err = h.k8sStorage.AddServiceBinding(&binding, nil)
	if err != nil {
		log.Printf("Error creating a service binding %s: %v\n", req.Name, err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, binding)
}

// CreateServiceBroker handles the 'create service broker' API request.
func (h *httpHandler) CreateServiceBroker(w http.ResponseWriter, r *http.Request) {
	var sbReq scmodel.CreateServiceBrokerRequest
	err := utils.BodyToObject(r, &sbReq)
	if err != nil {
		log.Printf("Error unmarshaling: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	sb := scmodel.ServiceBroker{
		GUID:         uuid.NewV4().String(),
		Name:         sbReq.Name,
		BrokerURL:    sbReq.BrokerURL,
		AuthUsername: sbReq.AuthUsername,
		AuthPassword: sbReq.AuthPassword,

		Created: time.Now().Unix(),
		Updated: 0,
	}
	sb.SelfURL = "/v2/service_brokers/" + sb.GUID

	// TODO: This should just store the record in k8s storage. FIX ME.
	created, err := h.controller.CreateServiceBroker(&sb)
	if err != nil {
		log.Printf("Error creating a service broker: %v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	sbRes := scmodel.CreateServiceBrokerResponse{
		Metadata: scmodel.ServiceBrokerMetadata{
			GUID:      created.GUID,
			CreatedAt: time.Unix(created.Created, 0).Format(time.RFC3339),
			URL:       created.SelfURL,
		},
		Entity: scmodel.ServiceBrokerEntity{
			Name:         created.Name,
			BrokerURL:    created.BrokerURL,
			AuthUsername: created.AuthUsername,
		},
	}
	utils.WriteResponse(w, 200, sbRes)
}
