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
	"log"

	"github.com/kubernetes-incubator/service-catalog/controller/storage"
	"github.com/kubernetes-incubator/service-catalog/controller/util"
	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"

	"github.com/satori/go.uuid"
)

const (
	catalogURLFormatString      = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"
	defaultNamespace            = "default"
)

type controller struct {
	storage storage.Storage
}

func createController(s storage.Storage) *controller {
	return &controller{
		storage: s,
	}
}

func (c *controller) updateServiceInstance(in *scmodel.ServiceInstance) error {
	// Currently there's no difference between create / update,
	// but for prepping for future, split these into two different
	// methods for now.
	return c.createServiceInstance(in)
}

func (c *controller) createServiceInstance(in *scmodel.ServiceInstance) error {
	params := in.Parameters

	// Inject all the bindings that are from this service instance.
	fromBindings := make(map[string]*scmodel.Credential)
	if err := c.getBindingsFrom(in.Name, fromBindings); err != nil {
		return err
	}

	if len(fromBindings) > 0 {
		if params == nil {
			params = make(map[string]interface{})
		}
		params["bindings"] = fromBindings
	}

	broker, err := storage.GetBrokerByServiceClass(c.storage, in.ServiceID)
	if err != nil {
		return err
	}
	client := util.CreateOpenServiceBrokerClient(broker)

	// Make the request to instantiate.
	createReq := &sbmodel.ServiceInstanceRequest{
		ServiceID:  in.ServiceID,
		PlanID:     in.PlanID,
		Parameters: params,
	}
	_, err = client.CreateServiceInstance(in.ID, createReq)
	return err
}

// getBindingsFrom returns the set of bindings for a consuming service instance.
//
// Binding data is passed to the service broker right now as part of the
// parameters in the form:
//
// parameters:
//   bindings:
//     <service-name>:
//       <credential>
func (c *controller) getBindingsFrom(sName string, fromBindings map[string]*scmodel.Credential) error {
	bindings, err := storage.GetBindingsForService(c.storage, sName, storage.From)
	if err != nil {
		log.Printf("Failed to fetch bindings for %s : %v", sName, err)
		return err
	}
	for _, b := range bindings {
		log.Printf("Found binding %s for service %s", b.Name, sName)
		fromBindings[b.Name] = &b.Credentials
	}
	return nil
}

// injectBindingIntoInstance causes a consuming service instance to be updated
// with new binding information. The actual binding injection happens during the
// instance update process.
func (c *controller) injectBindingIntoInstance(ID string) error {
	fromSI, err := c.storage.GetServiceInstance(defaultNamespace, ID)
	if err == nil && fromSI != nil {
		// Update the Service Instance with the new bindings
		log.Printf("Found existing FROM Service: %s, should update it", fromSI.Name)
		err = c.updateServiceInstance(fromSI)
		if err != nil {
			log.Printf("Failed to update existing FROM service %s : %v", fromSI.Name, err)
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// All the methods implementing ServiceController API go here for clarity sake.
///////////////////////////////////////////////////////////////////////////////
func (c *controller) CreateServiceInstance(in *scmodel.ServiceInstance) (*scmodel.ServiceInstance, error) {
	serviceID, planID, planName, err := storage.GetServicePlanInfo(c.storage, in.Service, in.Plan)
	if err != nil {
		log.Printf("Error fetching service ID: %v", err)
		return nil, err
	}
	in.ServiceID = serviceID
	in.PlanID = planID
	in.Plan = planName
	if in.ID == "" {
		in.ID = uuid.NewV4().String()
	}

	log.Printf("Instantiating service %s using service/plan %s : %s", in.Name, serviceID, planID)

	err = c.createServiceInstance(in)
	op := scmodel.LastOperation{}
	if err != nil {
		op.State = "FAILED"
		op.Description = err.Error()
		log.Printf("Failed to create service instance: %v", err)
	} else {
		op.State = "CREATED"
	}
	in.LastOperation = &op

	log.Printf("Updating Service %s with State\n%v", in.Name, in.LastOperation)
	return in, c.storage.UpdateServiceInstance(in)
}

func (c *controller) CreateServiceBinding(in *scmodel.ServiceBinding) (*scmodel.Credential, error) {
	log.Printf("Creating Service Binding: %v", in)

	// Get instance information for service being bound to.
	to, err := c.storage.GetServiceInstance(defaultNamespace, in.To)
	if err != nil {
		log.Printf("To service does not exist %s: %v", in.To, err)
		return nil, err
	}

	// Get broker associated with the service.
	st, err := c.storage.GetServiceClass(to.Service)
	if err != nil {
		log.Printf("Failed to fetch service type %s : %v", to.Service, err)
		return nil, err
	}
	broker, err := c.storage.GetBroker(st.Broker)
	if err != nil {
		log.Printf("Error fetching broker for service: %s : %v", to.Service, err)
		return nil, err
	}
	client := util.CreateOpenServiceBrokerClient(broker)

	// Assign UUID to binding.
	in.ID = uuid.NewV4().String()

	// Make the request to bind.
	createReq := &sbmodel.BindingRequest{
		ServiceID:  to.ServiceID,
		PlanID:     to.PlanID,
		Parameters: in.Parameters,
	}
	sbr, err := client.CreateServiceBinding(to.ID, in.ID, createReq)

	// Stash the credentials with the binding and update the binding.
	creds, err := util.ConvertCredential(&sbr.Credentials)
	if err != nil {
		log.Printf("Failed to convert creds: %v\n", err)
		return nil, err
	}
	in.Credentials = *creds

	err = c.storage.UpdateServiceBinding(in)
	if err != nil {
		log.Printf("Failed to update service binding %s : %v", in.Name, err)
		return nil, err
	}

	// NOTE: this is the plug-in point for changing binding injection. This
	// current function will inject binding information into a consuming service
	// instance.
	if err := c.injectBindingIntoInstance(in.From); err != nil {
		return nil, err
	}

	return &in.Credentials, nil
}

func (c *controller) CreateServiceBroker(in *scmodel.ServiceBroker) (*scmodel.ServiceBroker, error) {
	client := util.CreateOpenServiceBrokerClient(in)
	sbcat, err := client.GetCatalog()
	if err != nil {
		return nil, err
	}
	catalog, err := util.ConvertCatalog(sbcat)
	if err != nil {
		return nil, err
	}

	log.Printf("Adding a broker %s catalog:\n%v\n", in.Name, catalog)
	err = c.storage.AddBroker(in, catalog)
	if err != nil {
		return nil, err
	}
	return in, nil
}
