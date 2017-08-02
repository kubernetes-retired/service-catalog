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

package userbroker

import (
	"fmt"
	"sync"
	"errors"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokers/broker"
)

// errNoSuchInstance implements the Error interface.
// This struct handles the common error of an unrecogonzied instanceID
// and should be used as a returned error value.
// e.g. return errNoSuchInstance{instanceID: <id>}
type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("No such instance with ID %s", e.instanceID)
}

// userProvidedServiceInstance contains identifying data for each existing service instance.
type userProvidedServiceInstance struct {
	// Id is the instanceID
	Id         string                   `json:"id"`
	// Namespace is the k8s namespace provided in the CreateServiceInstanceReqeust.ContextProfile.Namespace
	Namespace  string                   `json:"namespace"`
	// ServiceID is the service's associated id.
	ServiceID  string                   `json:"serviceid"`
	// Credential is the binding credential created during Bind()
	Credential *brokerapi.Credential    `json:"credential"`
}

// userProvidedBroker implements the OSB API and represents the actual Broker.
type userProvidedBroker struct {
	// rwMutex controls concurrent R and RW access.
	rwMutex     sync.RWMutex
	// instanceMap should take instanceIDs as the key and maps to that ID's userProvidedServiceInstance
	instanceMap map[string]*userProvidedServiceInstance
}

const (
	// Service IDs should always be constants.  The variable names should be prefixed with "serviceid"
	// serviceidUserProvided is the basic demo. It provides no actual service
	serviceidUserProvided string = "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468"
	// serviceidDatabasePod  provides an instance of a mongo db
	serviceidDatabasePod  string = "database-1"
)

// CreateBroker initializes the service broker.  This function is called by server.Start()
func CreateBroker() broker.Broker {
	var instanceMap = make(map[string]*userProvidedServiceInstance)
	return &userProvidedBroker{
		instanceMap: instanceMap,
	}
}

// Catalog is an OSB method.  It returns a slice of services.
// New services should be specified here.
func (b *userProvidedBroker) Catalog() (*brokerapi.Catalog, error) {
	glog.Info("Catalog()")
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "user-provided-service",
				ID:          serviceidUserProvided,
				Description: "A user provided service",
				Plans: []brokerapi.ServicePlan{{
					Name:        "default",
					ID:          "86064792-7ea2-467b-af93-ac9694d96d52",
					Description: "Sample plan description",
					Free:        true,
				},
				},
				Bindable: true,
			},
			{
				Name:        "database-service",
				ID:          serviceidDatabasePod,
				Description: "A Hacky little pod service.",
				Plans: []brokerapi.ServicePlan{
					{
						Name:        "default",
						ID:          "default",
						Description: "There is only one, and this is it.",
						Free:        true,
					},
				},
				Bindable: true,
			},
		},
	}, nil
}

// CreateServiceInstance is an OSB method.  It handles provisioning of service instances
// as determined by the instance's serviceID.
// New services should be added as a new case in the switch.
func (b *userProvidedBroker) CreateServiceInstance(
	id string,
	req *brokerapi.CreateServiceInstanceRequest,
) (*brokerapi.CreateServiceInstanceResponse, error) {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()

	glog.Info("CreateServiceInstance", id)

	if _, ok := b.instanceMap[id]; ok {
		return nil, fmt.Errorf("Instance %q already exists", id)
	}
	// Create New Instance
	newInstance := &userProvidedServiceInstance{
		Id:        id,
		ServiceID: req.ServiceID,
		Namespace: req.ContextProfile.Namespace,
	}
	// Do provisioning logic based on service id
	switch newInstance.ServiceID {
	case serviceidUserProvided:
	case serviceidDatabasePod:
		err := doDBProvision(id, newInstance.Namespace)
		if err != nil {
			return nil, err
		}
	}
	glog.Infof("Provisioned Instance %q in Namespace %q", newInstance.Id, newInstance.Namespace)
	b.instanceMap[id] = newInstance
	return nil, nil
}

func (c *userProvidedBroker) GetServiceInstanceLastOperation(
	instanceID,
	serviceID,
	planID,
	operation string,
) (*brokerapi.LastOperationResponse, error) {
	glog.Info("GetServiceInstanceLastOperation()")
	return nil, errors.New("Unimplemented")
}

// RemoveServiceInstance is an OSB method.  It handles deprovisioning determined by the serviceID.
// New services should be added as a new case in the switch.
func (b *userProvidedBroker) RemoveServiceInstance(
	instanceID,
	serviceID,
	planID string,
	acceptsIncomplete bool,
) (*brokerapi.DeleteServiceInstanceResponse, error) {
	glog.Info("RemoveServiceInstance()")
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()

	// DEBUG
	glog.Infof("[DEBUG] Remove ServiceInstance Request (ID: %q)", instanceID)

	if _, ok := b.instanceMap[instanceID]; ! ok {
		return nil, errNoSuchInstance{instanceID: instanceID}
	}
	switch b.instanceMap[instanceID].ServiceID {
	case serviceidUserProvided:
		// Do nothing.
	case serviceidDatabasePod:
		if err := doDBDeprovision(instanceID, b.instanceMap[instanceID].Namespace); err != nil {
			err = fmt.Errorf("Error deprovisioning instance %q, %v", instanceID, err)
			glog.Error(err)
			return nil, err
		}
	}
	glog.Infof("Deprovisioned Instance: %q", b.instanceMap[instanceID].Id)
	delete(b.instanceMap, instanceID)
	return nil, nil
}

// Bind is an OSB method.  It handles bindings as determined by the serviceID.
// New services should be added as a new case in the switch.
// TODO implment bindMap to track db bindings (user, bindId, etc.)
func (b *userProvidedBroker) Bind(
	instanceID,
	bindingID string,
	req *brokerapi.BindingRequest,
) (*brokerapi.CreateServiceBindingResponse, error) {
	glog.Info("Bind()")
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	instance, ok := b.instanceMap[instanceID]
	if !ok {
		return nil, errNoSuchInstance{instanceID: instanceID}
	}
	var newCredential *brokerapi.Credential
	switch b.instanceMap[instanceID].ServiceID {
	case serviceidUserProvided:
		// Extract credentials from request or generate dummy
		newCredential = &brokerapi.Credential{
			"special-key-1": "special-value-1",
			"special-key-2": "special-value-2",
		}
	case serviceidDatabasePod:
		ip, port, err := doDBBind(instanceID, instance.Namespace)
		if err != nil {
			return nil, err
		}
		newCredential = &brokerapi.Credential{
			"mongoInstanceIp": ip,
			"mongoInstancePort": port,
		}
	}
	b.instanceMap[instanceID].Credential = newCredential
	glog.Infof("Bound Instance: %q", instanceID)
	return &brokerapi.CreateServiceBindingResponse{Credentials: *newCredential}, nil
}

// UnBind is an OSB method.  It handles credentials deletion relative to each service.
// New services should be added as a new case in the switch.
//TODO implement DB unbinding (delete user, etc)
func (b *userProvidedBroker) UnBind(instanceID, bindingID, serviceID, planID string) error {
	glog.Info("UnBind()")
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	// DEBUG
	glog.Infof("[DEBUG] Unind ServiceInstance Request (ID: %q)", instanceID)

	instance, ok := b.instanceMap[instanceID]
	if !ok {
		return errNoSuchInstance{instanceID: instanceID}
	}
	switch instance.ServiceID {
	case serviceidUserProvided:
		// Do nothing
	case serviceidDatabasePod:
		doDBUnbind()
	}
	glog.Infof("Unbound Instance: %q", instanceID)
	return nil
}
