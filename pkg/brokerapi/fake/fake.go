package fake

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
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
}

func (i *InstanceClient) CreateServiceInstance(
	ID string,
	req *brokerapi.ServiceInstanceRequest,
) (*brokerapi.ServiceInstance, error) {
	return nil, nil
}

func (i *InstanceClient) UpdateServiceInstance(
	ID string,
	req *ServiceInstanceRequest,
) (*ServiceInstance, error) {
	return nil, nil
}

func (i *InstanceClient) DeleteServiceInstance(ID string) error {
	return nil
}

type BindingClient struct {
}

func (b *BindingClient) CreateServiceBinding(
	sID,
	bID string,
	req *brokerapi.BindingRequest,
) (*brokerapi.CreateServiceBindingResponse, error) {
	return nil, nil
}

func (b *BindingClient) DeleteServiceBinding(sID, bID string) error {
	return nil
}
