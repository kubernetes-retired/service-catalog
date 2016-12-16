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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
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

func (c *controller) updateServiceInstance(in *servicecatalog.Instance) error {
	// Currently there's no difference between create / update,
	// but for prepping for future, split these into two different
	// methods for now.
	return c.createServiceInstance(in)
}

func (c *controller) createServiceInstance(in *servicecatalog.Instance) error {
	broker, err := storage.GetBrokerByServiceClass(c.storage.Brokers(), c.storage.ServiceClasses(), in.Spec.CFServiceID)
	if err != nil {
		return err
	}
	client := openservicebroker.NewClient(broker)

	// Make the request to instantiate.
	createReq := &sbmodel.ServiceInstanceRequest{
		ServiceID:  in.Spec.CFServiceID,
		PlanID:     in.Spec.CFPlanID,
		Parameters: in.Spec.Parameters,
	}
	_, err = client.CreateServiceInstance(in.Spec.CFGUID, createReq)
	return err
}

///////////////////////////////////////////////////////////////////////////////
// All the methods implementing ServiceController API go here for clarity sake.
///////////////////////////////////////////////////////////////////////////////
func (c *controller) CreateServiceInstance(in *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	serviceID, planID, planName, err := storage.GetServicePlanInfo(
		c.storage.ServiceClasses(),
		in.Spec.ServiceClassName,
		in.Spec.PlanName,
	)
	if err != nil {
		glog.Errorf("Error fetching service ID: %v", err)
		return nil, err
	}
	in.Spec.CFServiceID = serviceID
	in.Spec.CFPlanID = planID
	in.Spec.PlanName = planName
	if in.Spec.CFGUID == "" {
		in.Spec.CFGUID = uuid.NewV4().String()
	}

	glog.Infof("Instantiating service %s using service/plan %s : %s", in.Name, serviceID, planID)

	err = c.createServiceInstance(in)
	in.Status = servicecatalog.InstanceStatus{}
	if err != nil {
		in.Status.Conditions = []servicecatalog.InstanceCondition{
			{
				Type:   servicecatalog.InstanceConditionProvisionFailed,
				Status: servicecatalog.ConditionTrue,
				Reason: err.Error(),
			},
		}
		glog.Errorf("Failed to create service instance: %v", err)
	} else {
		in.Status.Conditions = []servicecatalog.InstanceCondition{
			{
				Type:   servicecatalog.InstanceConditionReady,
				Status: servicecatalog.ConditionTrue,
			},
		}
	}

	glog.Infof("Updating Service %s with State\n%v", in.Name, in.Status.Conditions[0].Type)
	return c.storage.Instances(in.ObjectMeta.Namespace).Update(in)
}

func (c *controller) CreateServiceBinding(in *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	glog.Infof("Creating Service Binding: %v", in)

	// Get instance information for service being bound to.
	instance, err := c.storage.Instances(in.Spec.InstanceRef.Namespace).Get(in.Spec.InstanceRef.Name)
	if err != nil {
		glog.Errorf("Service inatance does not exist %v: %v", in.Spec.InstanceRef, err)
		return nil, err
	}

	// Get the serviceclass for the instance.
	sc, err := c.storage.ServiceClasses().Get(instance.Spec.ServiceClassName)
	if err != nil {
		glog.Errorf("Failed to fetch service type %s : %v", instance.Spec.ServiceClassName, err)
		return nil, err
	}

	// Get the broker for the serviceclass.
	broker, err := c.storage.Brokers().Get(sc.BrokerName)
	if err != nil {
		glog.Errorf("Error fetching broker for service: %s : %v", sc.BrokerName, err)
		return nil, err
	}
	client := openservicebroker.NewClient(broker)

	// Assign UUID to binding.
	in.Spec.CFGUID = uuid.NewV4().String()

	// Make the request to bind.
	createReq := &sbmodel.BindingRequest{
		ServiceID:  instance.Spec.CFServiceID,
		PlanID:     instance.Spec.CFPlanID,
		Parameters: in.Spec.Parameters,
	}
	sbr, err := client.CreateServiceBinding(instance.Spec.CFGUID, in.Spec.CFGUID, createReq)

	in.Status = servicecatalog.BindingStatus{}
	if err != nil {
		in.Status.Conditions = []servicecatalog.BindingCondition{
			{
				Type:   servicecatalog.BindingConditionFailed,
				Status: servicecatalog.ConditionTrue,
				Reason: err.Error(),
			},
		}
		glog.Errorf("Failed to create service instance: %v", err)
	} else {
		// Now try injection
		err := c.injector.Inject(in, &sbr.Credentials)
		if err != nil {
			in.Status.Conditions = []servicecatalog.BindingCondition{
				{
					Type:   servicecatalog.BindingConditionFailed,
					Status: servicecatalog.ConditionTrue,
					Reason: err.Error(),
				},
			}
			glog.Errorf("Failed to create service instance: %v", err)
		} else {
			in.Status.Conditions = []servicecatalog.BindingCondition{
				{
					Type:   servicecatalog.BindingConditionReady,
					Status: servicecatalog.ConditionTrue,
				},
			}
		}
	}

	glog.Infof("Updating Service Binding %s with State\n%v", in.Name, in.Status.Conditions[0].Type)
	return c.storage.Bindings(in.ObjectMeta.Namespace).Update(in)
}

func (c *controller) CreateServiceBroker(in *servicecatalog.Broker) (*servicecatalog.Broker, error) {
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
	_, err = c.storage.Brokers().Create(in)
	if err != nil {
		return nil, err
	}

	for _, sc := range catalog {
		if _, err := c.storage.ServiceClasses().Create(sc); err != nil {
			return nil, err
		}
	}

	in.Status.Conditions = []servicecatalog.BrokerCondition{
		{
			Type:   servicecatalog.BrokerConditionReady,
			Status: servicecatalog.ConditionTrue,
		},
	}

	glog.Infof("Updating Service Broker %s with State\n%v", in.Name, in.Status.Conditions[0].Type)
	return c.storage.Brokers().Update(in)
}
