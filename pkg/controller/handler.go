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

package controller

import (
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
	"github.com/satori/go.uuid"
)

const (
	catalogURLFormatString      = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"
	defaultNamespace            = "default"
)

// Handler defines an interface used as a facade for data access operations.
// The controller uses the functions of this interface as callbacks for various
// events.
type Handler interface {
	// CreateServiceInstance takes in a (possibly incomplete)
	// ServiceInstance and will either create or update an
	// existing one.
	CreateServiceInstance(*servicecatalog.Instance) (*servicecatalog.Instance, error)

	// CreateServiceBinding takes in a (possibly incomplete)
	// ServiceBinding and will either create or update an
	// existing one.
	CreateServiceBinding(*servicecatalog.Binding) (*servicecatalog.Binding, error)

	// CreateServiceBroker takes in a (possibly incomplete)
	// ServiceBroker and will either create or update an
	// existing one.
	CreateServiceBroker(*servicecatalog.Broker) (*servicecatalog.Broker, error)
}

type handler struct {
	apiClient     apiclient.APIClient
	injector      injector.BindingInjector
	newClientFunc func(*servicecatalog.Broker) brokerapi.BrokerClient
}

func createHandler(client apiclient.APIClient, injector injector.BindingInjector, newClientFn brokerapi.CreateFunc) *handler {
	return &handler{
		apiClient:     client,
		injector:      injector,
		newClientFunc: newClientFn,
	}
}

func (h *handler) updateServiceInstance(in *servicecatalog.Instance) error {
	// Currently there's no difference between create / update,
	// but for prepping for future, split these into two different
	// methods for now.
	return h.createServiceInstance(in)
}

func (h *handler) createServiceInstance(in *servicecatalog.Instance) error {
	broker, err := apiclient.GetBrokerByServiceClassName(h.apiClient.Brokers(), h.apiClient.ServiceClasses(), in.Spec.ServiceClassName)
	if err != nil {
		return err
	}
	client := h.newClientFunc(broker)

	// TODO: uncomment parameters line once parameters types are refactored.
	// Make the request to instantiate.
	createReq := &brokerapi.CreateServiceInstanceRequest{
		ServiceID: in.Spec.OSBServiceID,
		PlanID:    in.Spec.OSBPlanID,
		// Parameters: in.Spec.Parameters,
	}
	_, err = client.CreateServiceInstance(in.Spec.OSBGUID, createReq)
	return err
}

///////////////////////////////////////////////////////////////////////////////
// All the methods implementing the Handler interface go here for clarity sake.
///////////////////////////////////////////////////////////////////////////////
func (h *handler) CreateServiceInstance(in *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	serviceID, planID, planName, err := apiclient.GetServicePlanInfo(
		h.apiClient.ServiceClasses(),
		in.Spec.ServiceClassName,
		in.Spec.PlanName,
	)
	if err != nil {
		glog.Errorf("Error fetching service ID: %v", err)
		return nil, err
	}
	in.Spec.OSBServiceID = serviceID
	in.Spec.OSBPlanID = planID
	in.Spec.PlanName = planName
	if in.Spec.OSBGUID == "" {
		in.Spec.OSBGUID = uuid.NewV4().String()
	}

	glog.Infof("Instantiating service %s using service/plan %s : %s", in.Name, serviceID, planID)

	err = h.createServiceInstance(in)
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
	return h.apiClient.Instances(in.ObjectMeta.Namespace).Update(in)
}

func (h *handler) CreateServiceBinding(in *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	glog.Infof("Creating Service Binding: %v", in)

	// Get instance information for service being bound to.
	instance, err := h.apiClient.Instances(in.Spec.InstanceRef.Namespace).Get(in.Spec.InstanceRef.Name)
	if err != nil {
		glog.Errorf("Service instance does not exist %v: %v", in.Spec.InstanceRef, err)
		return nil, err
	}

	// Get the serviceclass for the instance.
	sc, err := h.apiClient.ServiceClasses().Get(instance.Spec.ServiceClassName)
	if err != nil {
		glog.Errorf("Failed to fetch service type %s : %v", instance.Spec.ServiceClassName, err)
		return nil, err
	}

	// Get the broker for the serviceclass.
	broker, err := h.apiClient.Brokers().Get(sc.BrokerName)
	if err != nil {
		glog.Errorf("Error fetching broker for service: %s : %v", sc.BrokerName, err)
		return nil, err
	}
	client := h.newClientFunc(broker)

	// Assign UUID to binding.
	in.Spec.OSBGUID = uuid.NewV4().String()

	// TODO: uncomment parameters line once parameters types are refactored.
	// Make the request to bind.
	createReq := &brokerapi.BindingRequest{
		ServiceID: instance.Spec.OSBServiceID,
		PlanID:    instance.Spec.OSBPlanID,
		// Parameters: in.Spec.Parameters,
	}
	sbr, err := client.CreateServiceBinding(instance.Spec.OSBGUID, in.Spec.OSBGUID, createReq)

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
		err := h.injector.Inject(in, &sbr.Credentials)
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
	return h.apiClient.Bindings(in.ObjectMeta.Namespace).Update(in)
}

func (h *handler) CreateServiceBroker(in *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	client := h.newClientFunc(in)
	sbcat, err := client.GetCatalog()
	if err != nil {
		return nil, err
	}
	catalog, err := convertCatalog(sbcat)
	if err != nil {
		return nil, err
	}

	glog.Infof("Adding a broker %s catalog:\n%v\n", in.Name, catalog)
	_, err = h.apiClient.Brokers().Create(in)
	if err != nil {
		return nil, err
	}

	for _, sc := range catalog {
		sc.BrokerName = in.Name
		if _, err := h.apiClient.ServiceClasses().Create(sc); err != nil {
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
	return h.apiClient.Brokers().Update(in)
}

// convertCatalog converts a service broker catalog into an array of ServiceClasses
func convertCatalog(in *brokerapi.Catalog) ([]*servicecatalog.ServiceClass, error) {
	ret := make([]*servicecatalog.ServiceClass, len(in.Services))
	for i, svc := range in.Services {
		plans := convertServicePlans(svc.Plans)
		ret[i] = &servicecatalog.ServiceClass{
			Bindable:      svc.Bindable,
			Plans:         plans,
			PlanUpdatable: svc.PlanUpdateable,
			OSBGUID:       svc.ID,
			OSBTags:       svc.Tags,
			OSBRequires:   svc.Requires,
			// OSBMetadata:   svc.Metadata,
		}
		ret[i].SetName(svc.Name)
	}
	return ret, nil
}

func convertServicePlans(plans []brokerapi.ServicePlan) []servicecatalog.ServicePlan {
	ret := make([]servicecatalog.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = servicecatalog.ServicePlan{
			Name:    plan.Name,
			OSBGUID: plan.ID,
			// OSBMetadata: plan.Metadata,
			OSBFree: plan.Free,
		}
	}
	return ret
}
