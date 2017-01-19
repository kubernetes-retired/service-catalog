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
	"errors"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

type userProvidedServiceInstance struct {
	Name       string
	Credential *brokerapi.Credential
}

type userProvidedController struct {
	instanceMap map[string]*userProvidedServiceInstance
}

// CreateController creates an instance of a User Provided service broker controller.
func CreateController() controller.Controller {
	var instanceMap = make(map[string]*userProvidedServiceInstance)
	return &userProvidedController{
		instanceMap: instanceMap,
	}
}

func (c *userProvidedController) Catalog() (*brokerapi.Catalog, error) {
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "user-provided-service",
				ID:          "4F6E6CF6-FFDD-425F-A2C7-3C9258AD2468",
				Description: "User Provided Service",
				Plans: []brokerapi.ServicePlan{{
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

func (c *userProvidedController) CreateServiceInstance(id string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	credString, ok := req.Parameters["credentials"]
	if !ok {
		glog.Errorf("Didn't find creds\n %+v\n", req)
		return nil, errors.New("Credentials not found")
	}

	jsonCred, err := json.Marshal(credString)
	if err != nil {
		glog.Errorf("Failed to marshal credentials: %v", err)
		return nil, err
	}
	var cred brokerapi.Credential
	err = json.Unmarshal(jsonCred, &cred)

	c.instanceMap[id] = &userProvidedServiceInstance{
		Name:       id,
		Credential: &cred,
	}

	glog.Infof("Created User Provided Service Instance:\n%v\n", c.instanceMap[id])
	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *userProvidedController) GetServiceInstance(id string) (string, error) {
	return "", errors.New("Unimplemented")
}

func (c *userProvidedController) RemoveServiceInstance(id string) error {
	_, ok := c.instanceMap[id]
	if ok {
		delete(c.instanceMap, id)
		return nil
	}

	return errors.New("Not found")
}

func (c *userProvidedController) Bind(instanceID string, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	cred := c.instanceMap[instanceID].Credential
	return &brokerapi.CreateServiceBindingResponse{Credentials: *cred}, nil
}

func (c *userProvidedController) UnBind(instanceID string, bindingID string) error {
	// Since we don't persist the binding, there's nothing to do here.
	return nil
}
