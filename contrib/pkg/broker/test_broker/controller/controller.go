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
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/server"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokerapi"
)

const failAlways = math.MaxInt32

type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("no such instance with ID %s", e.instanceID)
}

type testServiceInstance struct {
	Name                         string
	Credential                   *brokerapi.Credential
	provisionedAt                time.Time
	remainingDeprovisionFailures int
}

type testService struct {
	brokerapi.Service
	Asynchronous         bool
	ProvisionFailTimes   int
	DeprovisionFailTimes int
}

type testController struct {
	rwMutex           sync.RWMutex
	serviceMap        map[string]*testService
	instanceMap       map[string]*testServiceInstance
	provisionCountMap map[string]int
}

// CreateController creates an instance of a Test service broker controller.
func CreateController() controller.Controller {
	var instanceMap = make(map[string]*testServiceInstance)
	services := []*testService{
		{
			Service: brokerapi.Service{
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
		},
		{
			Service: brokerapi.Service{
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
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-async",
				ID:          "b4073486-4759-4055-840a-f5f8b07231ff",
				Description: "A test service that is asynchronously provisioned",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "4f6741a8-2451-43c7-b473-a4f8e9f89a87",
						Description: "Sample plan description",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			Asynchronous: true,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-provision-fail",
				ID:          "15619930-5f4f-476a-87cd-7690901874c6",
				Description: "Provisioning of this service always returns HTTP status 500 (provisioning never succeeds)",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "525a787c-78d8-42af-8800-e9bf4bd71117",
						Description: "Default plan",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			ProvisionFailTimes: failAlways,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-provision-fail-5x",
				ID:          "226f24e0-def0-491d-a5b3-cd484bb6a4cf",
				Description: "Provisioning of this service fails 5 times, then succeeds.",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "21f83e68-0f4d-4377-bf5a-a5dddfaf7a5c",
						Description: "Default plan",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			ProvisionFailTimes: 5,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-deprovision-fail",
				ID:          "8207d20b-e428-44cd-bff4-20926aa19327",
				Description: "Provisioning of this service always succeeds, but deprovisiong always fails.",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "27ac655b-864e-4447-8bea-eb38a0e0cf79",
						Description: "Default plan",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			DeprovisionFailTimes: failAlways,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-deprovision-fail-5x",
				ID:          "07668858-b210-4101-916e-2627165af174",
				Description: "Provisioning of this service always succeeds, while deprovisioning fails 5 times, then succeeds.",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "3dab1aa9-4004-4252-b1ff-3d0bff42b36b",
						Description: "Default plan",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			DeprovisionFailTimes: 5,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-provision-fail-5x-deprovision-fail-5x",
				ID:          "38f9a4a1-c206-411b-ad33-71a1af979993",
				Description: "Provisioning of this service fails 5 times, then succeeds; deprovisioning also fails 5 times, then succeeds.",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "1179dfe7-9dbb-4d23-987f-2f722ca4f733",
						Description: "Default plan",
						Free:        true,
					},
				},
				Bindable:       true,
				PlanUpdateable: true,
			},
			ProvisionFailTimes:   5,
			DeprovisionFailTimes: 5,
		},
		{
			Service: brokerapi.Service{
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
	}

	var serviceMap = make(map[string]*testService)
	for _, s := range services {
		serviceMap[s.ID] = s
	}

	return &testController{
		instanceMap:       instanceMap,
		serviceMap:        serviceMap,
		provisionCountMap: make(map[string]int),
	}
}

func (c *testController) Catalog() (*brokerapi.Catalog, error) {
	glog.Info("Catalog()")
	services := []*brokerapi.Service{}
	for _, s := range c.serviceMap {
		services = append(services, &s.Service)
	}
	return &brokerapi.Catalog{
		Services: services,
	}, nil
}

func (c *testController) CreateServiceInstance(
	id string,
	req *brokerapi.CreateServiceInstanceRequest,
) (*brokerapi.CreateServiceInstanceResponse, error) {

	service, ok := c.serviceMap[req.ServiceID]
	if !ok {
		return nil, fmt.Errorf("Service %q does not exist", req.ServiceID)
	}

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
			Name:                         id,
			Credential:                   &cred,
			remainingDeprovisionFailures: service.DeprovisionFailTimes,
		}
	} else {
		c.instanceMap[id] = &testServiceInstance{
			Name: id,
			Credential: &brokerapi.Credential{
				"special-key-1": "special-value-1",
				"special-key-2": "special-value-2",
			},
			remainingDeprovisionFailures: service.DeprovisionFailTimes,
		}
	}

	c.provisionCountMap[id]++

	async := false
	if service.Asynchronous {
		async = true
		c.instanceMap[id].provisionedAt = time.Now().Add(1 * time.Minute)
	}

	glog.Infof("Created Test Service Instance:\n%v\n", c.instanceMap[id])
	if async {
		return &brokerapi.CreateServiceInstanceResponse{
			Operation: "provision",
		}, nil
	} else {
		provisionCount, _ := c.provisionCountMap[id]
		if provisionCount <= service.ProvisionFailTimes {
			return nil, server.NewErrorWithHttpStatus("Service is configured to fail provisioning", http.StatusInternalServerError)
		} else {
			return &brokerapi.CreateServiceInstanceResponse{}, nil
		}
	}
}

func (c *testController) GetServiceInstanceLastOperation(
	instanceID,
	serviceID,
	planID,
	operation string,
) (*brokerapi.LastOperationResponse, error) {
	glog.Info("GetServiceInstanceLastOperation()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	instance, ok := c.instanceMap[instanceID]

	switch operation {
	case "provision":
		if !ok {
			return nil, errors.New("Not found")
		}
		if instance.provisionedAt.Before(time.Now()) {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "Succeeded",
			}, nil
		} else {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateInProgress,
				Description: "Still provisioning...",
			}, nil
		}
	}

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
	instance, ok := c.instanceMap[instanceID]
	if ok {
		if instance.remainingDeprovisionFailures > 0 {
			instance.remainingDeprovisionFailures--
			return nil, server.NewErrorWithHttpStatus("Service is configured to fail deprovisioning", http.StatusInternalServerError)
		} else {
			delete(c.instanceMap, instanceID)
			return &brokerapi.DeleteServiceInstanceResponse{}, nil
		}
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
