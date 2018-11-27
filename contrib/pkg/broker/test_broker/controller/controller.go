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
	Name                           string
	Credential                     *brokerapi.Credential
	provisionedAt                  time.Time
	deprovisionedAt                time.Time
	remainingDeprovisionFailures   int
	remainingLastOperationFailures int
}

type testService struct {
	brokerapi.Service
	Asynchronous           bool
	ProvisionFailTimes     int
	DeprovisionFailTimes   int
	LastOperationFailTimes int
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
		newTestService(
			"test-service-single-plan",
			"4458dd64-8b63-4f84-9c1b-6a127614e122",
			"A test service that only has a single plan",
			"35b6030d-f81e-49cd-9d1f-2f5eaec57048",
			false, 0, 0, 0),
		newTestService(
			"test-service-async",
			"b4073486-4759-4055-840a-f5f8b07231ff",
			"A test service that is asynchronously provisioned & deprovisioned",
			"4f6741a8-2451-43c7-b473-a4f8e9f89a87",
			true, 0, 0, 0),
		newTestService(
			"test-service-async-last-operation-fail",
			"40b0dff1-2180-460e-98f1-890c127e3613",
			"A test service that is asynchronously provisioned, but lastOperation never succeeds",
			"624eea7a-4fb1-4e67-9ec8-379f0c855c3b",
			true, 0, 0, failAlways),
		newTestService(
			"test-service-async-last-operation-fail-5x",
			"c3bb2dda-8946-4f84-a66e-e4957d8f0e07",
			"A test service that is asynchronously provisioned, but lastOperation only succeeds on the 5th attempt.",
			"4254a380-4e3d-4cc1-b2b6-3c7e55b63ea2",
			true, 0, 0, 5),
		newTestService(
			"test-service-provision-fail",
			"15619930-5f4f-476a-87cd-7690901874c6",
			"Provisioning of this service always returns HTTP status 500 (provisioning never succeeds)",
			"525a787c-78d8-42af-8800-e9bf4bd71117",
			false, failAlways, 0, 0),
		newTestService(
			"test-service-provision-fail-5x",
			"226f24e0-def0-491d-a5b3-cd484bb6a4cf",
			"Provisioning of this service fails 5 times, then succeeds.",
			"21f83e68-0f4d-4377-bf5a-a5dddfaf7a5c",
			false, 5, 0, 0),
		newTestService(
			"test-service-deprovision-fail",
			"8207d20b-e428-44cd-bff4-20926aa19327",
			"Provisioning of this service always succeeds, but deprovisiong always fails.",
			"27ac655b-864e-4447-8bea-eb38a0e0cf79",
			false, 0, failAlways, 0),
		newTestService(
			"test-service-deprovision-fail-5x",
			"07668858-b210-4101-916e-2627165af174",
			"Provisioning of this service always succeeds, while deprovisioning fails 5 times, then succeeds.",
			"3dab1aa9-4004-4252-b1ff-3d0bff42b36b",
			false, 0, 5, 0),
		newTestService(
			"test-service-provision-fail-5x-deprovision-fail-5x",
			"38f9a4a1-c206-411b-ad33-71a1af979993",
			"Provisioning of this service fails 5 times, then succeeds; deprovisioning also fails 5 times, then succeeds.",
			"1179dfe7-9dbb-4d23-987f-2f722ca4f733",
			false, 5, 5, 0),
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

func newTestService(name string, id string, description string, planID string, async bool, provisionFailTimes int, deprovisionFailTimes int, lastOperationFailTimes int) *testService {
	return &testService{
		Service: brokerapi.Service{
			Name:        name,
			ID:          id,
			Description: description,
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default",
					ID:          planID,
					Description: "Default plan",
					Free:        true,
				},
			},
			Bindable:       true,
			PlanUpdateable: true,
		},
		Asynchronous:           async,
		ProvisionFailTimes:     provisionFailTimes,
		DeprovisionFailTimes:   deprovisionFailTimes,
		LastOperationFailTimes: lastOperationFailTimes,
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
			Name:                           id,
			Credential:                     &cred,
			remainingDeprovisionFailures:   service.DeprovisionFailTimes,
			remainingLastOperationFailures: service.LastOperationFailTimes,
		}
	} else {
		c.instanceMap[id] = &testServiceInstance{
			Name: id,
			Credential: &brokerapi.Credential{
				"special-key-1": "special-value-1",
				"special-key-2": "special-value-2",
			},
			remainingDeprovisionFailures:   service.DeprovisionFailTimes,
			remainingLastOperationFailures: service.LastOperationFailTimes,
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

	instance, exists := c.instanceMap[instanceID]
	if !exists {
		return nil, server.NewErrorWithHttpStatus("Instance not found", http.StatusGone)
	}

	if instance.remainingLastOperationFailures > 0 {
		instance.remainingLastOperationFailures--
		return nil, server.NewErrorWithHttpStatus("Service is configured to fail lastOperation", http.StatusInternalServerError)
	} else {
		// reset remainingLastOperationFalures
		service, ok := c.serviceMap[serviceID]
		if ok {
			instance.remainingLastOperationFailures = service.LastOperationFailTimes
		}
	}

	switch operation {
	case "provision":
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
	case "deprovision":
		if instance.deprovisionedAt.Before(time.Now()) {
			delete(c.instanceMap, instanceID)
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "Succeeded",
			}, nil
		} else {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateInProgress,
				Description: "Still deprovisioning...",
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
		service, ok := c.serviceMap[serviceID]
		if ok {
			if service.Asynchronous {
				instance.deprovisionedAt = time.Now().Add(1 * time.Minute)
				return &brokerapi.DeleteServiceInstanceResponse{
					Operation: "deprovision",
				}, nil
			} else {
				if instance.remainingDeprovisionFailures > 0 {
					instance.remainingDeprovisionFailures--
					return nil, server.NewErrorWithHttpStatus("Service is configured to fail deprovisioning", http.StatusInternalServerError)
				} else {
					delete(c.instanceMap, instanceID)
					return &brokerapi.DeleteServiceInstanceResponse{}, nil
				}
			}
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
