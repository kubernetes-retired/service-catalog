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

package controller

import (
	"encoding/json"
	"log"

	"github.com/kubernetes-incubator/service-catalog/broker/controller"

	"errors"
	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
)

type UserProvidedServiceInstance struct {
	Name       string
	Credential *sbmodel.Credential
}

type Controller struct {
	instanceMap map[string]*UserProvidedServiceInstance
}

// Verify that Controller implements the broker Controller interface.
var _ controller.Controller = (*Controller)(nil)

func CreateController() *Controller {
	var instanceMap map[string]*UserProvidedServiceInstance = make(map[string]*UserProvidedServiceInstance)
	return &Controller{
		instanceMap: instanceMap,
	}
}

func (c *Controller) Catalog() (*sbmodel.Catalog, error) {
	return &sbmodel.Catalog{
		Services: []*sbmodel.Service{
			{
				Name:        "user-provided-service",
				ID:          "4F6E6CF6-FFDD-425F-A2C7-3C9258AD2468",
				Description: "User Provided Service",
				Plans: []sbmodel.ServicePlan{{
					Name:        "default",
					ID:          "86064792-7ea2-467b-af93-ac9694d96d52",
					Description: "User Provided Service",
					Free:        true,
				},
				},
			},
		},
	}, nil
}

func (c *Controller) CreateServiceInstance(id string, req *sbmodel.ServiceInstanceRequest) (*sbmodel.CreateServiceInstanceResponse, error) {
	credString, ok := req.Parameters["credentials"]
	if !ok {
		log.Printf("Didn't find creds\n %+v\n", req)
		return nil, errors.New("Credentials not found")
	}

	jsonCred, err := json.Marshal(credString)
	if err != nil {
		log.Printf("Failed to marshal credentials: %v", err)
		return nil, err
	}
	var cred sbmodel.Credential
	err = json.Unmarshal(jsonCred, &cred)

	c.instanceMap[id] = &UserProvidedServiceInstance{
		Name:       id,
		Credential: &cred,
	}

	log.Printf("Created User Provided Service Instance:\n%v\n", c.instanceMap[id])
	return &sbmodel.CreateServiceInstanceResponse{}, nil
}

func (c *Controller) GetServiceInstance(id string) (string, error) {
	return "", errors.New("Unimplemented")
}

func (c *Controller) RemoveServiceInstance(id string) error {
	_, ok := c.instanceMap[id]
	if ok {
		delete(c.instanceMap, id)
		return nil
	} else {
		return errors.New("Not found")
	}
}

func (c *Controller) Bind(instanceId string, bindingId string, req *sbmodel.BindingRequest) (*sbmodel.CreateServiceBindingResponse, error) {
	cred := c.instanceMap[instanceId].Credential
	return &sbmodel.CreateServiceBindingResponse{Credentials: *cred}, nil
}

func (c *Controller) UnBind(instanceId string, bindingId string) error {
	// Since we don't persist the binding, there's nothing to do here.
	return nil
}
