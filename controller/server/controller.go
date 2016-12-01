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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	k8sStorage ServiceStorage
}

func createController(k8sStorage ServiceStorage) *controller {
	return &controller{
		k8sStorage: k8sStorage,
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

	// Then actually make the request to reify the service instance
	createReq := &sbmodel.ServiceInstanceRequest{
		ServiceID:  in.ServiceID,
		PlanID:     in.PlanID,
		Parameters: params,
	}

	jsonBytes, err := json.Marshal(createReq)
	if err != nil {
		return err
	}

	broker, err := c.getBroker(in.ServiceID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(serviceInstanceFormatString, broker.BrokerURL, in.ID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	client := &http.Client{}
	log.Printf("Doing a request to: %s", url)
	resp, err := client.Do(createHTTPReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: Align this with the actual response model.
	si := scmodel.ServiceInstance{}
	err = util.ResponseBodyToObject(resp, &si)
	if err != nil {
		return err
	}
	return nil
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
	bindings, err := c.k8sStorage.GetBindingsForService(sName, From)
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

func (c *controller) getBroker(serviceID string) (*scmodel.ServiceBroker, error) {
	broker, err := c.k8sStorage.GetBrokerByService(serviceID)
	if err != nil {
		return nil, err
	}

	return broker, nil
}

// fetchServicePlanGUID fetches the GUIDs for Service and Plan, also
// returns the name of the plan since it might get defaulted.
// If Plan is not given and there's only one plan for a given service, we'll choose that.
func (c *controller) fetchServicePlanGUID(service string, plan string) (string, string, string, error) {
	s, err := c.k8sStorage.GetServiceType(service)
	if err != nil {
		return "", "", "", err
	}
	// No plan specified and only one plan, use it.
	if plan == "" && len(s.Plans) == 1 {
		log.Printf("Found Service Plan GUID as %s for %s : %s", s.Plans[0].ID, service, s.Plans[0].Name)
		return s.ID, s.Plans[0].ID, s.Plans[0].Name, nil
	}
	for _, p := range s.Plans {
		if p.Name == plan {
			fmt.Printf("Found Service Plan GUID as %s for %s : %s", p.ID, service, plan)
			return s.ID, p.ID, p.Name, nil
		}
	}
	return "", "", "", fmt.Errorf("Can't find a service / plan : %s/%s", service, plan)
}

// injectBindingIntoInstance causes a consuming service instance to be updated
// with new binding information. The actual binding injection happens during the
// instance update process.
func (c *controller) injectBindingIntoInstance(ID string) error {
	fromSI, err := c.k8sStorage.GetService(defaultNamespace, ID)
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
	serviceID, planID, planName, err := c.fetchServicePlanGUID(in.Service, in.Plan)
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
	return in, c.k8sStorage.SetService(in)
}

func (c *controller) CreateServiceBinding(in *scmodel.ServiceBinding) (*scmodel.Credential, error) {
	log.Printf("Creating Service Binding: %v", in)

	// Get instance information for service being bound to.
	to, err := c.k8sStorage.GetService(defaultNamespace, in.To)
	if err != nil {
		log.Printf("To service does not exist %s: %v", in.To, err)
		return nil, err
	}

	// Then actually make the request to create the binding
	createReq := &sbmodel.BindingRequest{
		ServiceID:  to.ServiceID,
		PlanID:     to.PlanID,
		Parameters: in.Parameters,
	}

	jsonBytes, err := json.Marshal(createReq)
	if err != nil {
		log.Printf("Failed to marshal: %#v", err)
		return nil, err
	}

	in.ID = uuid.NewV4().String()

	st, err := c.k8sStorage.GetServiceType(to.Service)
	if err != nil {
		log.Printf("Failed to fetch service type %s : %v", to.Service, err)
		return nil, err
	}
	broker, err := c.k8sStorage.GetBroker(st.Broker)
	if err != nil {
		log.Printf("Error fetching broker for service: %s : %v", to.Service, err)
		return nil, err
	}
	url := fmt.Sprintf(bindingFormatString, broker.BrokerURL, to.ID, in.ID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	client := &http.Client{}
	log.Printf("Doing a request to: %s", url)
	resp, err := client.Do(createHTTPReq)
	if err != nil {
		log.Printf("Failed to PUT: %#v", err)
		return nil, err
	}
	defer resp.Body.Close()

	sbr := scmodel.CreateServiceBindingResponse{}
	err = util.ResponseBodyToObject(resp, &sbr)
	if err != nil {
		log.Printf("Failed to unmarshal: %#v", err)
		return nil, err
	}

	// Stash the credentials with the binding and update the binding.
	in.Credentials = sbr.Credentials

	err = c.k8sStorage.UpdateServiceBinding(in)
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
	// Fetch the catalog from the broker
	u := fmt.Sprintf(catalogURLFormatString, in.BrokerURL)
	req, err := http.NewRequest("GET", u, nil)
	req.SetBasicAuth(in.AuthUsername, in.AuthPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch catalog from %s\n%v", u, resp)
		log.Printf("err: %#v", err)
		return nil, err
	}

	// TODO: the model from SB is fetched and stored directly as the one in the SC model (which the
	// storage operates on). We should convert it from the SB model to SC model before storing.
	var catalog scmodel.Catalog
	err = util.ResponseBodyToObject(resp, &catalog)
	if err != nil {
		log.Printf("Failed to unmarshal catalog: %#v", err)
		return nil, err
	}

	log.Printf("Adding a broker %s catalog:\n%v", in.Name, catalog)

	err = c.k8sStorage.AddBroker(in, &catalog)
	if err != nil {
		return nil, err
	}
	return in, nil
}
