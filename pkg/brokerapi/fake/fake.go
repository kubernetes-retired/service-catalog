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

package fake

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	uuid "github.com/satori/go.uuid"
)

// Client implements a fake broker API client, useful for unit testing. None of the methods on
// the client are concurrency-safe
type Client struct {
	CatalogClient
	InstanceClient
	BindingClient
}

// NewClientFunc returns a function suitable for creating a new BrokerClient from a given
// Broker object. The returned function is suitable for passing as a callback to code that
// needs to create clients on-demand
func NewClientFunc(
	catCl CatalogClient,
	instCl InstanceClient,
	bindCl BindingClient,
) func(*servicecatalog.Broker) brokerapi.BrokerClient {
	return func(*servicecatalog.Broker) brokerapi.BrokerClient {
		return &Client{
			CatalogClient:  catCl,
			InstanceClient: instCl,
			BindingClient:  bindCl,
		}
	}
}

// CatalogClient implements a fake CF catalog API client
type CatalogClient struct {
	RetCatalog *brokerapi.Catalog
	RetErr     error
}

// GetCatalog just returns c.RetCatalog and c.RetErr
func (c *CatalogClient) GetCatalog() (*brokerapi.Catalog, error) {
	return c.RetCatalog, c.RetErr
}

// InstanceClient implements a fake CF instance API client
type InstanceClient struct {
	Instances map[string]*brokerapi.ServiceInstance
	CreateErr error
	UpdateErr error
	DeleteErr error
}

// CreateServiceInstance returns i.CreateErr if non-nil. If it is nil, checks if id already exists
// in i.Instances and returns ErrInstanceAlreadyExists if so. If not, converts req to a
// ServiceInstance, adds it to i.Instances and returns it
func (i *InstanceClient) CreateServiceInstance(
	id string,
	req *brokerapi.ServiceInstanceRequest,
) (*brokerapi.ServiceInstance, error) {

	if i.CreateErr != nil {
		return nil, i.CreateErr
	}
	if i.exists(id) {
		return nil, ErrInstanceAlreadyExists
	}

	i.Instances[id] = convertInstanceRequest(req)
	return i.Instances[id], nil
}

// UpdateServiceInstance returns i.UpdateErr if it was non-nil. Otherwise, returns
// ErrInstanceNotFound if id already exists in i.Instances. If it didn't exist, converts req into
// a ServiceInstance, adds it to i.Instances and returns it
func (i *InstanceClient) UpdateServiceInstance(
	id string,
	req *brokerapi.ServiceInstanceRequest,
) (*brokerapi.ServiceInstance, error) {

	if i.UpdateErr != nil {
		return nil, i.UpdateErr
	}
	if !i.exists(id) {
		return nil, ErrInstanceNotFound
	}

	i.Instances[id] = convertInstanceRequest(req)
	return i.Instances[id], nil
}

// DeleteServiceInstance returns i.DeleteErr if it was non-nil. Otherwise returns
// ErrInstanceNotFound if id didn't already exist in i.Instances. If it it did already exist,
// removes i.Instances[id] from the map and returns nil
func (i *InstanceClient) DeleteServiceInstance(id string) error {

	if i.DeleteErr != nil {
		return i.DeleteErr
	}
	if !i.exists(id) {
		return ErrInstanceNotFound
	}
	delete(i.Instances, id)
	return nil
}

func (i *InstanceClient) exists(id string) bool {
	_, ok := i.Instances[id]
	return ok
}

func convertInstanceRequest(req *brokerapi.ServiceInstanceRequest) *brokerapi.ServiceInstance {
	return &brokerapi.ServiceInstance{
		ID:               uuid.NewV4().String(),
		DashboardURL:     "https://github.com/kubernetes-incubator/service-catalog",
		InternalID:       uuid.NewV4().String(),
		ServiceID:        req.ServiceID,
		PlanID:           req.PlanID,
		OrganizationGUID: uuid.NewV4().String(),
		SpaceGUID:        req.SpaceID,
		LastOperation:    nil,
		Parameters:       req.Parameters,
	}
}

// BindingClient implements a fake CF binding API client
type BindingClient struct {
	Bindings    map[string]struct{}
	CreateCreds brokerapi.Credential
	CreateErr   error
	DeleteErr   error
}

// CreateServiceBinding returns b.CreateErr if it was non-nil. Otherwise, returns
// ErrBindingAlreadyExists if the IDs already existed in b.Bindings. If they didn't already exist,
// adds the IDs to b.Bindings and returns a new CreateServiceBindingResponse with b.CreateCreds in
// it
func (b *BindingClient) CreateServiceBinding(
	sID,
	bID string,
	req *brokerapi.BindingRequest,
) (*brokerapi.CreateServiceBindingResponse, error) {

	if b.CreateErr != nil {
		return nil, b.CreateErr
	}
	if b.exists(sID, bID) {
		return nil, ErrBindingAlreadyExists
	}

	b.Bindings[bindingsMapKey(sID, bID)] = struct{}{}
	return &brokerapi.CreateServiceBindingResponse{Credentials: b.CreateCreds}, nil
}

// DeleteServiceBinding returns b.DeleteErr if it was non-nil. Otherwise, if the binding associated
// with the given IDs didn't exist, returns ErrBindingNotFound. If it did exist, removes it and
// returns nil
func (b *BindingClient) DeleteServiceBinding(sID, bID string) error {
	if b.DeleteErr != nil {
		return b.DeleteErr
	}
	if !b.exists(sID, bID) {
		return ErrBindingNotFound
	}

	delete(b.Bindings, bindingsMapKey(sID, bID))
	return nil
}

func (b *BindingClient) exists(sID, bID string) bool {
	_, ok := b.Bindings[bindingsMapKey(sID, bID)]
	return ok
}

func bindingsMapKey(sID, bID string) string {
	return fmt.Sprintf("%s:%s", sID, bID)
}
