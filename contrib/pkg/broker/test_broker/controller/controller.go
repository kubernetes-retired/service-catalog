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

package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokerapi"
)

type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("no such instance with ID %s", e.instanceID)
}

type testServiceInstance struct {
	Name       string
	Credential *brokerapi.Credential
}

type testController struct {
	rwMutex     sync.RWMutex
	instanceMap map[string]*testServiceInstance
}

// CreateController creates an instance of a Test service broker controller.
func CreateController() controller.Controller {
	var instanceMap = make(map[string]*testServiceInstance)
	return &testController{
		instanceMap: instanceMap,
	}
}

func (c *testController) Catalog() (*brokerapi.Catalog, error) {
	glog.Info("Catalog()")
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "test-service",
				ID:          "fe43b7d8-f0d4-11e8-bdba-54ee754ec85f",
				Description: "A test service",
				Plans: []brokerapi.ServicePlan{{
					Name:        "default",
					ID:          "06576262-f0d5-11e8-83eb-54ee754ec85f",
					Description: "Sample plan description",
					Free:        true,
				}, {
					Name:        "premium",
					ID:          "e251a5bb-3266-4391-bdde-be9e87bffe2f",
					Description: "Premium plan",
					Free:        false,
				},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			{
				Name:        "test-service-single-plan",
				ID:          "4458dd64-8b63-4f84-9c1b-6a127614e122",
				Description: "A test service that only has a single plan",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "35b6030d-f81e-49cd-9d1f-2f5eaec57048",
						Description: "Sample plan description",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			{
				Name:        "test-service-with-schemas",
				ID:          "c57f5b14-804e-4a3b-9047-755a7f145961",
				Description: "A test service with parameter and response schemas",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "0b8e785e-9053-4acf-9eb8-c15f879ff485",
						Description: "Plan with parameter and response schemas",
						Free:        true,
						Schemas: &brokerapi.Schemas{
							ServiceInstance: &brokerapi.ServiceInstanceSchema{
								Create: &brokerapi.InputParametersSchema{
									Parameters: map[string]interface{}{ // TODO: use a JSON Schema library instead?
										"$schema": "http://json-schema.org/draft-04/schema#",
										"type":    "object",
										"properties": map[string]interface{}{
											"param-1": map[string]interface{}{
												"description": "First input parameter",
												"type":        "string",
											},
											"param-2": map[string]interface{}{
												"description": "Second input parameter",
												"type":        "string",
											},
										},
									},
								},
								Update: &brokerapi.InputParametersSchema{
									Parameters: map[string]interface{}{
										"$schema": "http://json-schema.org/draft-04/schema#",
										"type":    "object",
										"properties": map[string]interface{}{
											"param-1": map[string]interface{}{
												"description": "First input parameter",
												"type":        "string",
											},
											"param-2": map[string]interface{}{
												"description": "Second input parameter",
												"type":        "string",
											},
										},
									},
								},
							},
							ServiceBinding: &brokerapi.ServiceBindingSchema{
								Create: &brokerapi.RequestResponseSchema{
									InputParametersSchema: brokerapi.InputParametersSchema{
										Parameters: map[string]interface{}{
											"$schema": "http://json-schema.org/draft-04/schema#",
											"type":    "object",
											"properties": map[string]interface{}{
												"param-1": map[string]interface{}{
													"description": "First input parameter",
													"type":        "string",
												},
												"param-2": map[string]interface{}{
													"description": "Second input parameter",
													"type":        "string",
												},
											},
										},
									},
									Response: map[string]interface{}{
										"$schema": "http://json-schema.org/draft-04/schema#",
										"type":    "object",
										"properties": map[string]interface{}{
											"credentials": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"special-key-1": map[string]interface{}{
														"description": "Special key 1",
														"type":        "string",
													},
													"special-key-2": map[string]interface{}{
														"description": "Special key 2",
														"type":        "string",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
		},
	}, nil
}

func (c *testController) CreateServiceInstance(
	id string,
	req *brokerapi.CreateServiceInstanceRequest,
) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Info("CreateServiceInstance()")
	credString, ok := req.Parameters["credentials"]
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	if ok {
		jsonCred, err := json.Marshal(credString)
		if err != nil {
			glog.Errorf("Failed to marshal credentials: %v", err)
			return nil, err
		}
		var cred brokerapi.Credential
		err = json.Unmarshal(jsonCred, &cred)
		if err != nil {
			glog.Errorf("Failed to unmarshal credentials: %v", err)
			return nil, err
		}

		c.instanceMap[id] = &testServiceInstance{
			Name:       id,
			Credential: &cred,
		}
	} else {
		c.instanceMap[id] = &testServiceInstance{
			Name: id,
			Credential: &brokerapi.Credential{
				"special-key-1": "special-value-1",
				"special-key-2": "special-value-2",
			},
		}
	}

	glog.Infof("Created Test Service Instance:\n%v\n", c.instanceMap[id])
	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *testController) GetServiceInstanceLastOperation(
	instanceID,
	serviceID,
	planID,
	operation string,
) (*brokerapi.LastOperationResponse, error) {
	glog.Info("GetServiceInstanceLastOperation()")
	return nil, errors.New("Unimplemented")
}

func (c *testController) RemoveServiceInstance(
	instanceID,
	serviceID,
	planID string,
	acceptsIncomplete bool,
) (*brokerapi.DeleteServiceInstanceResponse, error) {
	glog.Info("RemoveServiceInstance()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	_, ok := c.instanceMap[instanceID]
	if ok {
		delete(c.instanceMap, instanceID)
		return &brokerapi.DeleteServiceInstanceResponse{}, nil
	}

	return &brokerapi.DeleteServiceInstanceResponse{}, nil
}

func (c *testController) Bind(
	instanceID,
	bindingID string,
	req *brokerapi.BindingRequest,
) (*brokerapi.CreateServiceBindingResponse, error) {
	glog.Info("Bind()")
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	instance, ok := c.instanceMap[instanceID]
	if !ok {
		return nil, errNoSuchInstance{instanceID: instanceID}
	}
	cred := instance.Credential
	return &brokerapi.CreateServiceBindingResponse{Credentials: *cred}, nil
}

func (c *testController) UnBind(instanceID, bindingID, serviceID, planID string) error {
	glog.Info("UnBind()")
	// Since we don't persist the binding, there's nothing to do here.
	return nil
}
