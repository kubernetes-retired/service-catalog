package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient/mem"
	"k8s.io/kubernetes/pkg/api"
)

const (
	namespace    = "testNS"
	brokerName   = "testBroker"
	svcClassName = "testSvcClass"
	instanceName = "testInstance"
	bindingName  = "testBinding"
)

var (
	binding = servicecatalog.Binding{
		ObjectMeta: api.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		Spec: servicecatalog.BindingSpec{
			InstanceRef: api.ObjectReference{
				Namespace: namespace,
				Name:      instanceName,
			},
		},
	}
)

func makeTraverableAPIClient() apiclient.APIClient {
	instances := map[string]apiclient.InstanceClient{
		namespace: mem.NewPopulatedInstanceClient(
			map[string]*servicecatalog.Instance{
				instanceName: {
					ObjectMeta: api.ObjectMeta{
						Namespace: namespace,
						Name:      instanceName,
					},
					Spec: servicecatalog.InstanceSpec{
						ServiceClassName: svcClassName,
					},
				},
			},
		),
	}
	bindings := map[string]apiclient.BindingClient{
		namespace: mem.NewPopulatedBindingClient(
			map[string]*servicecatalog.Binding{
				bindingName: &binding,
			},
		),
	}
	return mem.NewPopulatedAPIClient(
		map[string]*servicecatalog.Broker{
			brokerName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      brokerName,
				},
			},
		},
		map[string]*servicecatalog.ServiceClass{
			svcClassName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      svcClassName,
				},
				BrokerName: brokerName,
			},
		},
		instances,
		bindings,
	)
}

func TestAllTehThings(t *testing.T) {
	storage := makeTraverableAPIClient()
	inst, err := instanceForBinding(storage, &binding)
	if err != nil {
		t.Fatalf("error getting instance for binding (%s)", err)
	}
	svcClass, err := serviceClassForInstance(storage, inst)
	if err != nil {
		t.Fatalf("error getting service class for instance (%s)", err)
	}
	broker, err := brokerForServiceClass(storage, svcClass)
	if err != nil {
		t.Fatalf("error getting broker for service class (%s)", err)
	}
	if broker == nil {
		t.Fatalf("broker was nil")
	}
}
