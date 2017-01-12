package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/storage/mem"
	"k8s.io/kubernetes/pkg/api"
)

func TestCreateServiceInstanceHelper(t *testing.T) {
	const (
		brokerName   = "testBroker"
		svcClassName = "testSvcClass"
	)
	storage := mem.NewPopulatedStorage(
		map[string]*servicecatalog.Broker{
			brokerName: &servicecatalog.Broker{
				ObjectMeta: api.ObjectMeta{
					Name: brokerName,
				},
			},
		},
		map[string]*servicecatalog.ServiceClass{
			svcClassName: &servicecatalog.ServiceClass{
				ObjectMeta: api.ObjectMeta{
					Name: svcClassName,
				},
				BrokerName: brokerName,
			},
		},
	)
	inj := injector.NewFake()
	catalogCl := fake.CatalogClient{}
	instanceCl := fake.NewInstanceClient()
	bindingCl := fake.NewBindingClient()
	brokerClFunc := fake.NewClientFunc(catalogCl, *instanceCl, *bindingCl)
	hdl := createHandler(storage, inj, brokerClFunc)

	inst := &servicecatalog.Instance{
		Spec: servicecatalog.InstanceSpec{
			ServiceClassName: svcClassName,
		},
		Status: servicecatalog.InstanceStatus{},
	}
	if err := hdl.createServiceInstance(inst); err != nil {
		t.Fatalf("error creating service instance (%s)", err)
	}
}
