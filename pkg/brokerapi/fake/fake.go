package fake

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	uuid "github.com/satori/go.uuid"
)

type Client struct {
	CatalogClient
	InstanceClient
	BindingClient
}

type CatalogClient struct {
	RetCatalog *brokerapi.Catalog
	RetErr     error
}

func (c *CatalogClient) GetCatalog() (*brokerapi.Catalog, error) {
	return c.RetCatalog, c.RetErr
}

type InstanceClient struct {
	Instances map[string]*brokerapi.ServiceInstance
	CreateErr error
	UpdateErr error
	DeleteErr error
}

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
	return inst, nil
}

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
	return inst, nil
}

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

type BindingClient struct {
	Bindings    map[string]struct{}
	CreateCreds brokerapi.Credential
	CreateErr   error
	DeleteErr   error
}

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

func (b *BindingClient) DeleteServiceBinding(sID, bID string) error {
	if b.DeleteErr != nil {
		return b.DeleteErr
	}
	return nil
}

func (b *BindingClient) exists(sID, bID string) bool {
	_, ok := b.Bindings[bindingsMapKey(sID, bID)]
	return ok
}

func bindingsMapKey(sID, bID string) string {
	return fmt.Sprintf("%s:%s", sID, bID)
}
