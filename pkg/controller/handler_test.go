package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/storage/mem"
)

func TestCreateServiceInstanceHelper(t *testing.T) {
	storage := mem.NewStorage()
	inj := injector.NewFake()
	catalogCl := fake.CatalogClient{}
	instanceCl := fake.InstanceClient{}
	bindingCl := fake.BindingClient{}
	brokerClFunc := fake.NewClientFunc(catalogCl, instanceCl, bindingCl)
	hdl := createHandler(storage, inj, brokerClFunc)

	inst := &servicecatalog.Instance{
		Spec:   servicecatalog.InstanceSpec{},
		Status: servicecatalog.InstanceStatus{},
	}
	if err := hdl.createServiceInstance(inst); err != nil {
		t.Fatalf("error creating service instance (%s)", err)
	}
}
