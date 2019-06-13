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

	"github.com/kubernetes-sigs/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-sigs/service-catalog/contrib/pkg/broker/server"
	"github.com/kubernetes-sigs/service-catalog/contrib/pkg/brokerapi"
	"k8s.io/klog"
)

const failAlways = math.MaxInt32
const noHTTPError = 0

type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("no such instance with ID %s", e.instanceID)
}

type testServiceInstance struct {
	Name                  string
	Credential            *brokerapi.Credential
	provisionedAt         time.Time
	updatedAt             time.Time
	deprovisionedAt       time.Time
	deprovisionAttempts   int
	lastOperationAttempts int
	updateAttempts        int
}

type testService struct {
	brokerapi.Service
	Asynchronous           bool
	ProvisionFailTimes     int
	UpdateFailTimes        int
	DeprovisionFailTimes   int
	LastOperationFailTimes int
	HTTPErrorStatus        int
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
		newTestService(
			"test-service",
			"2f2e85b5-030d-4776-ba7e-e26eb312f10f",
			"A test service that only has a single plan",
			"35b6030d-f81e-49cd-9d1f-2f5eaec57048",
			false, http.StatusBadRequest, 0, 0, 0, 0),
		newTestService(
			"test-service-provision-fail400",
			"308c0400-2edb-45d6-a63e-67f18226a404",
			"Provisioning of this service always returns HTTP status 400 (which is a terminal, non-retriable error))",
			"44443058-077e-43f3-9857-7ca7efedafd9",
			false, http.StatusBadRequest, failAlways, 0, 0, 0),
		newTestService(
			"test-service-provision-fail500",
			"308c0500-2edb-45d6-a63e-67f18226a404",
			"Provisioning of this service always returns HTTP status 500 (provisioning never succeeds)",
			"525a787c-78d8-42af-8800-e9bf4bd71117",
			false, http.StatusInternalServerError, failAlways, 0, 0, 0),
		newTestService(
			"test-service-provision-fail500-5x",
			"389e6500-93f9-49b4-bbe4-76e304cad22c",
			"Provisioning of this service fails 5 times, then succeeds.",
			"21f83e68-0f4d-4377-bf5a-a5dddfaf7a5c",
			false, http.StatusInternalServerError, 5, 0, 0, 0),
		newTestService(
			"test-service-provision-fail500-5x-deprovision-fail500-5x",
			"41f7f500-118c-4f22-a4e9-fc56c02046c0",
			"Provisioning of this service fails 5 times, then succeeds; deprovisioning also fails 5 times, then succeeds.",
			"1179dfe7-9dbb-4d23-987f-2f722ca4f733",
			false, http.StatusInternalServerError, 5, 0, 5, 0),
		newTestService(
			"test-service-deprovision-fail400",
			"43e24400-93ae-4c7d-bfd3-7cd03f051872",
			"Provisioning of this service always succeeds, but deprovisiong always fails with error 400 (a non-retriable error).",
			"b8e55ea4-05a7-43d6-a0f8-64fbee9e6cc6",
			false, http.StatusBadRequest, 0, 0, failAlways, 0),
		newTestService(
			"test-service-deprovision-fail500",
			"43e24500-93ae-4c7d-bfd3-7cd03f051872",
			"Provisioning of this service always succeeds, but deprovisiong always fails.",
			"27ac655b-864e-4447-8bea-eb38a0e0cf79",
			false, http.StatusInternalServerError, 0, 0, failAlways, 0),
		newTestService(
			"test-service-deprovision-fail500-5x",
			"4ed5a500-35ed-4748-be64-5007951373ab",
			"Provisioning of this service always succeeds, while deprovisioning fails 5 times, then succeeds.",
			"3dab1aa9-4004-4252-b1ff-3d0bff42b36b",
			false, http.StatusInternalServerError, 0, 0, 5, 0),
		newTestService(
			"test-service-update-fail400",
			"4efa9400-aafb-4738-94ab-e6e10a2f4af8",
			"Update of this service always returns HTTP status 400 (which is a terminal, non-retriable error)",
			"e3d738b6-8d5c-4f40-ba5b-2613e02af41d",
			false, http.StatusBadRequest, 0, failAlways, 0, 0),
		newTestService(
			"test-service-update-fail500",
			"4efa9500-aafb-4738-94ab-e6e10a2f4af8",
			"Update of this service always returns HTTP status 500 (update never succeeds)",
			"729c5f1f-aef4-4c38-81db-227993ec24c6",
			false, http.StatusInternalServerError, 0, failAlways, 0, 0),
		newTestService(
			"test-service-update-fail500-5x",
			"4f1eb500-6762-4605-917a-cfca0eaa9b01",
			"Update of this service fails 5 times, then succeeds.",
			"eb5a24ba-69ab-4acb-964a-dcad600ba4d3",
			false, http.StatusInternalServerError, 0, 5, 0, 0),
		newTestService(
			"test-service-async",
			"5a680caf-807e-4157-85af-552dc71b72d6",
			"A test service that is asynchronously provisioned & deprovisioned",
			"4f6741a8-2451-43c7-b473-a4f8e9f89a87",
			true, noHTTPError, 0, 0, 0, 0),
		newTestService(
			"test-service-async-provision-fail",
			"7aac9500-c42a-46f4-86d6-df21437d4c7f",
			"A test service that is asynchronously provisioned, but provisioning always returns state:failed",
			"9aca0b9a-192e-416a-a809-67e592bfa681",
			true, noHTTPError, failAlways, 0, 0, 0),
		newTestService(
			"test-service-async-provision-fail-5x",
			"7f73e500-1ba0-4882-94c7-7624b4219520",
			"A test service that is asynchronously provisioned; provisioning returns state:failed 5 times, then succeeds.",
			"a1027080-966d-4ec3-b4e1-abc3f52b7de2",
			true, noHTTPError, 5, 0, 0, 0),
		newTestService(
			"test-service-async-provision-fail-5x-deprovision-fail-5x",
			"86709500-1acb-473b-baa8-899e4dce12dc",
			"A test service that is asynchronously provisioned; provisioning returns state:failed 5 times, then succeeds; deprovisioning also returns state:failed 5 times, then succeeds.",
			"35234488-830f-4efe-ae16-a36bb0092cce",
			true, noHTTPError, 5, 0, 5, 0),
		newTestService(
			"test-service-async-deprovision-fail",
			"9bee1500-e5f7-4bd8-94de-eb65c811be83",
			"A test service that is asynchronously provisioned; provisioning always succeeds, deprovisiong always returns state:failed.",
			"6096a7e0-7ea6-4782-8246-c6e5d9eb97ca",
			true, noHTTPError, 0, 0, failAlways, 0),
		newTestService(
			"test-service-async-deprovision-fail-5x",
			"acddd500-97e5-4c69-99e2-d1a056b1ad25",
			"A test service that is asynchronously provisioned; provisioning always succeeds, deprovisioning returns state:failed 5 times, then succeeds.",
			"dce5da49-fc42-4490-a053-8415fd569461",
			true, noHTTPError, 0, 0, 5, 0),
		newTestService(
			"test-service-async-update-fail",
			"ad6ab500-c287-4090-a9ab-6d49b1204496",
			"Update of this service always returns state:failed in the last operation response",
			"94f9a5fd-6a99-440d-9315-ddb144755349",
			true, noHTTPError, 0, failAlways, 0, 0),
		newTestService(
			"test-service-async-update-fail-5x",
			"aec24500-f8a5-4c95-a02b-92b297bf7805",
			"Update of this service returns state:failed 5 times, then succeeds.",
			"e11860e1-f62f-4383-9eb4-30d8641fe2f0",
			true, noHTTPError, 0, 5, 0, 0),
		newTestService(
			"test-service-async-last-operation-fail400",
			"c594a400-ec7f-494b-a266-d540cf977382",
			"A test service that is asynchronously provisioned, but lastOperation always fails with error 400",
			"e937e0b6-ddd5-4565-82e2-1cda3d16ad32",
			true, http.StatusBadRequest, 0, 0, 0, failAlways),
		newTestService(
			"test-service-async-last-operation-fail500",
			"c594a500-ec7f-494b-a266-d540cf977382",
			"A test service that is asynchronously provisioned, but lastOperation always fails with error 500",
			"624eea7a-4fb1-4e67-9ec8-379f0c855c3b",
			true, http.StatusInternalServerError, 0, 0, 0, failAlways),
		newTestService(
			"test-service-async-last-operation-fail500-5x",
			"cce99500-3f6e-42f1-8100-5408a7b79e43",
			"A test service that is asynchronously provisioned, but lastOperation only succeeds on the 5th attempt.",
			"4254a380-4e3d-4cc1-b2b6-3c7e55b63ea2",
			true, http.StatusInternalServerError, 0, 0, 0, 5),
		{
			Service: brokerapi.Service{
				Name:        "test-service-multiple-plans",
				ID:          "f1b57a42-8035-4291-a555-51c461df6072",
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
			DeprovisionFailTimes: 0,
		},
		{
			Service: brokerapi.Service{
				Name:        "test-service-with-schemas",
				ID:          "f485442d-319b-43d4-80ef-bdf7ae200b09",
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
			DeprovisionFailTimes: 0,
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

func newTestService(name string, id string, description string, planID string, async bool, httpErrorStatus int, provisionFailTimes, updateFailTimes, deprovisionFailTimes, lastOperationFailTimes int) *testService {
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
		UpdateFailTimes:        updateFailTimes,
		DeprovisionFailTimes:   deprovisionFailTimes,
		LastOperationFailTimes: lastOperationFailTimes,
		HTTPErrorStatus:        httpErrorStatus,
	}
}

func (c *testController) Catalog() (*brokerapi.Catalog, error) {
	klog.Info("Catalog()")
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

	klog.Info("CreateServiceInstance()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	service, ok := c.serviceMap[req.ServiceID]
	if !ok {
		return nil, fmt.Errorf("Service %q does not exist", req.ServiceID)
	}

	cred, err := getCredentials(req.Parameters)
	if err != nil {
		return nil, err
	}

	instance := &testServiceInstance{
		Name:       id,
		Credential: &cred,
	}
	c.instanceMap[id] = instance

	c.provisionCountMap[id]++

	if service.Asynchronous {
		klog.Infof("Starting asynchronous creation of Service Instance:\n%v\n", instance)
		instance.provisionedAt = time.Now().Add(1 * time.Minute)
		return &brokerapi.CreateServiceInstanceResponse{
			Operation: "provision",
		}, nil
	}

	provisionCount, _ := c.provisionCountMap[id]
	if provisionCount <= service.ProvisionFailTimes {
		return nil, server.NewErrorWithHTTPStatus("Service is configured to fail provisioning", service.HTTPErrorStatus)
	}

	klog.Infof("Created Test Service Instance:\n%v\n", instance)
	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *testController) UpdateServiceInstance(
	id string,
	req *brokerapi.UpdateServiceInstanceRequest,
) (*brokerapi.UpdateServiceInstanceResponse, error) {
	klog.Info("UpdateServiceInstance()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	instance, exists := c.instanceMap[id]
	if !exists {
		return nil, server.NewErrorWithHTTPStatus("Instance not found", http.StatusGone)
	}

	service, ok := c.serviceMap[req.ServiceID]
	if !ok {
		return nil, fmt.Errorf("Service %q does not exist", req.ServiceID)
	}

	cred, err := getCredentials(req.Parameters)
	if err != nil {
		return nil, err
	}
	instance.Credential = &cred

	instance.updateAttempts++

	if service.Asynchronous {
		klog.Infof("Starting asynchronous update of Service Instance:\n%v\n", instance)
		instance.updatedAt = time.Now().Add(1 * time.Minute)
		return &brokerapi.UpdateServiceInstanceResponse{
			Operation: "update",
		}, nil
	}

	if instance.updateAttempts <= service.UpdateFailTimes {
		return nil, server.NewErrorWithHTTPStatus("Service is configured to fail update", service.HTTPErrorStatus)
	}

	klog.Infof("Updated Test Service Instance:\n%v\n", instance)
	return &brokerapi.UpdateServiceInstanceResponse{}, nil
}

func getCredentials(requestParameters map[string]interface{}) (brokerapi.Credential, error) {
	credString, found := requestParameters["credentials"]
	if !found {
		return brokerapi.Credential{
			"special-key-1": "special-value-1",
			"special-key-2": "special-value-2",
		}, nil
	}

	jsonCred, err := json.Marshal(credString)
	if err != nil {
		klog.Errorf("Failed to marshal credentials: %v", err)
		return nil, err
	}
	var cred brokerapi.Credential
	err = json.Unmarshal(jsonCred, &cred)
	if err != nil {
		klog.Errorf("Failed to unmarshal credentials: %v", err)
		return nil, err
	}
	return cred, nil
}

func (c *testController) GetServiceInstanceLastOperation(
	instanceID,
	serviceID,
	planID,
	operation string,
) (*brokerapi.LastOperationResponse, error) {
	klog.Info("GetServiceInstanceLastOperation()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	instance, exists := c.instanceMap[instanceID]
	if !exists {
		return nil, server.NewErrorWithHTTPStatus("Instance not found", http.StatusGone)
	}

	service, exists := c.serviceMap[serviceID]
	if !exists {
		return nil, errors.New("Service not found")
	}

	instance.lastOperationAttempts++

	if instance.lastOperationAttempts <= service.LastOperationFailTimes {
		return nil, server.NewErrorWithHTTPStatus("Service is configured to fail lastOperation", service.HTTPErrorStatus)
	}

	var completionTime time.Time
	var attempts int
	var attemptsToFail int
	var deleteInstance bool

	switch operation {
	case "provision":
		completionTime = instance.provisionedAt
		attempts, _ = c.provisionCountMap[instanceID]
		attemptsToFail = service.ProvisionFailTimes
	case "update":
		completionTime = instance.updatedAt
		attempts = instance.updateAttempts
		attemptsToFail = service.UpdateFailTimes
	case "deprovision":
		completionTime = instance.deprovisionedAt
		attempts = instance.deprovisionAttempts
		attemptsToFail = service.DeprovisionFailTimes
		deleteInstance = true
	default:
		return nil, errors.New("Unimplemented")
	}

	if completionTime.After(time.Now()) {
		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateInProgress,
			Description: "Operation still in progress...",
		}, nil
	}

	if attempts <= attemptsToFail {
		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "Failed",
		}, nil
	}

	if deleteInstance {
		delete(c.instanceMap, instanceID)
	}

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "Succeeded",
	}, nil
}

func (c *testController) RemoveServiceInstance(
	instanceID,
	serviceID,
	planID string,
	acceptsIncomplete bool,
) (*brokerapi.DeleteServiceInstanceResponse, error) {
	klog.Info("RemoveServiceInstance()")
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	instance, ok := c.instanceMap[instanceID]
	if ok {
		service, ok := c.serviceMap[serviceID]
		if ok {
			if service.Asynchronous {
				klog.Infof("Starting asynchronous deletion of Service Instance:\n%v\n", instance)
				instance.deprovisionedAt = time.Now().Add(1 * time.Minute)
				return &brokerapi.DeleteServiceInstanceResponse{
					Operation: "deprovision",
				}, nil
			}

			if service.DeprovisionFailTimes > 0 && instance.deprovisionAttempts < service.DeprovisionFailTimes {
				instance.deprovisionAttempts++
				return nil, server.NewErrorWithHTTPStatus("Service is configured to fail deprovisioning", service.HTTPErrorStatus)
			}

			delete(c.instanceMap, instanceID)
			klog.Infof("Deleted Test Service Instance:\n%v\n", instance)
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
	klog.Info("Bind()")
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
	klog.Info("UnBind()")
	// Since we don't persist the binding, there's nothing to do here.
	return nil
}
