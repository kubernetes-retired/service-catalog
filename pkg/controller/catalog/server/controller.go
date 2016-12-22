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

package server

import (
	"github.com/golang/glog"
	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/openservicebroker"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/injector"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/storage"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/util"

	"github.com/satori/go.uuid"
)

const (
	catalogURLFormatString      = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"
	defaultNamespace            = "default"
)

type controller struct {
	storage  storage.Storage
	injector injector.BindingInjector
}

func createController(s storage.Storage) (*controller, error) {
	bi, err := injector.CreateK8sBindingInjector()
	if err != nil {
		return nil, err
	}

	return &controller{
		storage:  s,
		injector: bi,
	}, nil
}

func (c *controller) updateServiceInstance(in *scmodel.ServiceInstance) error {
	// Currently there's no difference between create / update,
	// but for prepping for future, split these into two different
	// methods for now.
	return c.createServiceInstance(in)
}

func (c *controller) createServiceInstance(in *scmodel.ServiceInstance) error {
	broker, err := storage.GetBrokerByServiceClass(c.storage, in.ServiceID)
	if err != nil {
		return err
	}
	client := openservicebroker.NewClient(broker)

	// Make the request to instantiate.
	createReq := &sbmodel.ServiceInstanceRequest{
		ServiceID:  in.ServiceID,
		PlanID:     in.PlanID,
		Parameters: in.Parameters,
	}
	_, err = client.CreateServiceInstance(in.ID, createReq)
	return err
}

///////////////////////////////////////////////////////////////////////////////
// All the methods implementing ServiceController API go here for clarity sake.
///////////////////////////////////////////////////////////////////////////////
func (c *controller) CreateServiceInstance(in *scmodel.ServiceInstance) (*scmodel.ServiceInstance, error) {
	serviceID, planID, planName, err := storage.GetServicePlanInfo(c.storage, in.Service, in.Plan)
	if err != nil {
		glog.Errorf("Error fetching service ID: %v", err)
		return nil, err
	}
	in.ServiceID = serviceID
	in.PlanID = planID
	in.Plan = planName
	if in.ID == "" {
		in.ID = uuid.NewV4().String()
	}

	glog.Infof("Instantiating service %s using service/plan %s : %s", in.Name, serviceID, planID)

	err = c.createServiceInstance(in)
	op := scmodel.LastOperation{}
	if err != nil {
		op.State = "FAILED"
		op.Description = err.Error()
		glog.Errorf("Failed to create service instance: %v", err)
	} else {
		op.State = "CREATED"
	}
	in.LastOperation = &op

	glog.Infof("Updating Service %s with State\n%v", in.Name, in.LastOperation)
	return in, c.storage.UpdateServiceInstance(in)
}

func (c *controller) CreateServiceBinding(in *scmodel.ServiceBinding) (*scmodel.Credential, error) {
	glog.Infof("Creating Service Binding: %v", in)

	// Get instance information for service being bound to.
	to, err := c.storage.GetServiceInstance(defaultNamespace, in.To)
	if err != nil {
		glog.Errorf("To service does not exist %s: %v", in.To, err)
		return nil, err
	}

	// Get broker associated with the service.
	st, err := c.storage.GetServiceClass(to.Service)
	if err != nil {
		glog.Errorf("Failed to fetch service type %s : %v", to.Service, err)
		return nil, err
	}
	broker, err := c.storage.GetBroker(st.Broker)
	if err != nil {
		glog.Errorf("Error fetching broker for service: %s : %v", to.Service, err)
		return nil, err
	}
	client := openservicebroker.NewClient(broker)

	// Assign UUID to binding.
	in.ID = uuid.NewV4().String()

	// Make the request to bind.
	createReq := &sbmodel.BindingRequest{
		ServiceID:  to.ServiceID,
		PlanID:     to.PlanID,
		Parameters: in.Parameters,
	}
	sbr, err := client.CreateServiceBinding(to.ID, in.ID, createReq)
	if err != nil {
		glog.Errorf("Failed to create service binding: %v\n", err)
		return nil, err
	}

	// Stash the credentials with the binding and update the binding.
	creds, err := util.ConvertCredential(&sbr.Credentials)
	if err != nil {
		glog.Errorf("Failed to convert creds: %v\n", err)
		return nil, err
	}
	in.Credentials = *creds

	err = c.storage.UpdateServiceBinding(in)
	if err != nil {
		glog.Errorf("Failed to update service binding %s : %v", in.Name, err)
		return nil, err
	}

	if err := c.injector.Inject(in); err != nil {
		return nil, err
	}

	return &in.Credentials, nil
}

func (c *controller) CreateServiceBroker(in *scmodel.ServiceBroker) (*scmodel.ServiceBroker, error) {
	client := openservicebroker.NewClient(in)
	sbcat, err := client.GetCatalog()
	if err != nil {
		return nil, err
	}
	catalog, err := util.ConvertCatalog(sbcat)
	if err != nil {
		return nil, err
	}

	glog.Infof("Adding a broker %s catalog:\n%v\n", in.Name, catalog)
	err = c.storage.AddBroker(in, catalog)
	if err != nil {
		return nil, err
	}
	return in, nil
}
