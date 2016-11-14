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

	"github.com/kubernetes-incubator/service-catalog/controller/utility"
	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
	"github.com/satori/go.uuid"
)

const (
	CATALOG_URL_FMT_STR      = "%s/v2/catalog"
	SERVICE_INSTANCE_FMT_STR = "%s/v2/service_instances/%s"
	BIND_FMT_STR             = "%s/v2/service_instances/%s/service_bindings/%s"
	DEFAULT_NAMESPACE        = "default"
)

type Controller struct {
	k8sStorage ServiceStorage
}

func CreateController(k8sStorage ServiceStorage) *Controller {
	return &Controller{
		k8sStorage: k8sStorage,
	}
}

//
// Inventory.
//

func (c *Controller) Inventory(w http.ResponseWriter, r *http.Request) {
	log.Println("Inventory")

	i, err := c.k8sStorage.GetInventory()
	if err != nil {
		log.Printf("Got Error: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, i)
}

//
// Service Broker.
//

func (c *Controller) ListServiceBrokers(w http.ResponseWriter, r *http.Request) {
	l, err := c.k8sStorage.ListBrokers()
	if err != nil {
		log.Printf("Got Error: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, l)
}

func (c *Controller) GetServiceBroker(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractVarFromRequest(r, "broker")
	log.Printf("GetServiceBroker: %s\n", id)

	b, err := c.k8sStorage.GetBroker(id)
	if err != nil {
		log.Printf("Got Error: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, b)
}

func (c *Controller) DeleteServiceBroker(w http.ResponseWriter, r *http.Request) {
	utils.WriteResponse(w, 400, "IMPLEMENT ME")
}

func (c *Controller) RefreshServiceBroker(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractVarFromRequest(r, "broker")

	sb, err := c.k8sStorage.GetBroker(id)
	if err != nil {
		log.Printf("Broker not found: %s\n", id)
		log.Printf("err: %#v\n", err)
		utils.WriteResponse(w, 404, "Not Found")
		return
	}

	// Fetch the catalog from the broker
	u := fmt.Sprintf(CATALOG_URL_FMT_STR, sb.BrokerURL)
	req, err := http.NewRequest("GET", u, nil)
	req.SetBasicAuth(sb.AuthUsername, sb.AuthPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch catalog from %s\n%v\n", u, resp)
		log.Printf("err: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	// TODO: the model from SB is fetched and stored directly as the one in the SC model (which the
	// storage operates on). We should convert it from the SB model to SC model before storing.
	var catalog scmodel.Catalog
	err = utils.ResponseBodyToObject(resp, &catalog)
	if err != nil {
		log.Printf("Failed to unmarshal catalog: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	err = c.k8sStorage.UpdateBroker(sb, &catalog)
	utils.WriteResponse(w, 200, "{}")
}

//
// Service Instances.
//

func (c *Controller) ListServiceInstances(w http.ResponseWriter, r *http.Request) {
	ns := utils.ExtractVarFromRequest(r, "namespace")
	if ns == "" {
		ns = DEFAULT_NAMESPACE
	}
	instances, err := c.k8sStorage.ListServices(ns)
	if err != nil {
		log.Printf("Couldn't list services: %v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, instances)
}

func (c *Controller) GetServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Printf("Getting Service Instance\n")
	id := utils.ExtractVarFromRequest(r, "service")
	ns := utils.ExtractVarFromRequest(r, "namespace")
	if ns == "" {
		ns = DEFAULT_NAMESPACE
	}

	si, err := c.k8sStorage.GetService(ns, id)
	if err != nil {
		log.Printf("Couldn't fetch the service: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	utils.WriteResponse(w, 200, si)
}

func (c *Controller) updateServiceInstance(in *scmodel.ServiceInstance) error {
	// Currently there's no difference between create / update,
	// but for prepping for future, split these into two different
	// methods for now.
	return c.createServiceInstance(in)
}

func (c *Controller) createServiceInstance(in *scmodel.ServiceInstance) error {
	params := in.Parameters

	// Inject all the bindings that are supposed to be injected (IE, these are direction 'FROM').
	fromBindings := make(map[string]*scmodel.Credential)
	bindings, err := c.k8sStorage.GetBindingsForService(in.Name, From)
	if err != nil {
		log.Printf("Failed to fetch bindings for %s : %v\n", in.Name, err)
		return err
	}
	for _, b := range bindings {
		log.Printf("Found binding %s for service %s\n", b.Name, in.Name)
		fromBindings[b.Name] = &b.Credentials
	}

	// Binding data is passed to the service broker right now as part of the
	// parameters in the form:
	//
	// parameters:
	//   bindings:
	//     <service-name>:
	//       <credential>
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

	url := fmt.Sprintf(SERVICE_INSTANCE_FMT_STR, broker.BrokerURL, in.ID)

	// TODO: Handle the auth
	createHttpReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	client := &http.Client{}
	log.Printf("Doing a request to: %s\n", url)
	resp, err := client.Do(createHttpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: Align this with the actual response model.
	si := scmodel.ServiceInstance{}
	err = utils.ResponseBodyToObject(resp, &si)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) DeleteServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Printf("Deleting Service Instance\n")
	id := utils.ExtractVarFromRequest(r, "service")
	ns := utils.ExtractVarFromRequest(r, "namespace")
	if ns == "" {
		ns = DEFAULT_NAMESPACE
	}

	si, err := c.k8sStorage.GetService(ns, id)
	if err != nil {
		log.Printf("Service id doesn't exist: %s\n", id)
		err := fmt.Errorf("Service id %s doesn't exist\n", id)
		utils.WriteResponse(w, 400, err)
		return
	}

	// Fetch the bindings this Service Instance is bound as Target, so we warn
	// the user if there are Service Instances using this before deleting.
	bindings, err := c.k8sStorage.GetBindingsForService(id, To)
	if err != nil {
		log.Printf("Failed to get bindings for %s\n", id)
		utils.WriteResponse(w, 400, err)
		return
	}

	if len(bindings) > 0 {
		log.Printf("There are %d active findings to this service, cowardly refusing to delete it\n", len(bindings))
		err = fmt.Errorf("There are %d active findings to this service, cowardly refusing to delete it\n", len(bindings))
		utils.WriteResponse(w, 400, err)
		return
	}

	// Delete the service from the broker
	broker, err := c.getBroker(si.ServiceID)
	if err != nil {
		log.Printf("Error fetching service: %v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	url := fmt.Sprintf(SERVICE_INSTANCE_FMT_STR, broker.BrokerURL, si.ID)

	// TODO: Handle the auth
	deleteHttpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("Failed to create new HTTP request: %v", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	client := &http.Client{}
	log.Printf("Doing a request to: %s\n", url)
	resp, err := client.Do(deleteHttpReq)
	if err != nil {
		log.Printf("Failed to DELETE: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	defer resp.Body.Close()

	err = c.k8sStorage.DeleteService(id)
	if err != nil {
		log.Printf("Failed to delete: %v", err)
		utils.WriteResponse(w, 400, err)
		return
	}
}

func (c *Controller) ListServiceBindings(w http.ResponseWriter, r *http.Request) {
	l, err := c.k8sStorage.ListServiceBindings()
	if err != nil {
		log.Printf("Got Error: %#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}
	utils.WriteResponse(w, 200, l)
}

func (c *Controller) GetServiceBinding(w http.ResponseWriter, r *http.Request) {
	log.Printf("Getting Service Binding\n")
	id := utils.ExtractVarFromRequest(r, "binding")

	b, err := c.k8sStorage.GetServiceBinding(id)
	if err != nil {
		log.Printf("%#v\n", err)
		utils.WriteResponse(w, 400, err)
		return
	}

	utils.WriteResponse(w, 200, b)
}

func (c *Controller) DeleteServiceBinding(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractVarFromRequest(r, "binding")

	// TODO: Update user of this binding...
	c.k8sStorage.DeleteServiceBinding(id)
}

func (c *Controller) getBroker(serviceID string) (*scmodel.ServiceBroker, error) {
	broker, err := c.k8sStorage.GetBrokerByService(serviceID)
	if err != nil {
		return nil, err
	}

	return broker, nil
}

// fetchServicePlanGUID fetches the GUIDs for Service and Plan, also
// returns the name of the plan since it might get defaulted.
// If Plan is not given and there's only one plan for a given service, we'll choose that.
func (c *Controller) fetchServicePlanGUID(service string, plan string) (string, string, string, error) {
	s, err := c.k8sStorage.GetServiceType(service)
	if err != nil {
		return "", "", "", err
	}
	// No plan specified and only one plan, use it.
	if plan == "" && len(s.Plans) == 1 {
		log.Printf("Found Service Plan GUID as %s for %s : %s\n", s.Plans[0].ID, service, s.Plans[0].Name)
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

///////////////////////////////////////////////////////////////////////////////
// All the methods implementing ServiceController API go here for clarity sake.
///////////////////////////////////////////////////////////////////////////////
func (c *Controller) CreateServiceInstance(in *scmodel.ServiceInstance) (*scmodel.ServiceInstance, error) {
	serviceID, planID, planName, err := c.fetchServicePlanGUID(in.Service, in.Plan)
	if err != nil {
		log.Printf("Error fetching service ID: %v\n", err)
		return nil, err
	}
	in.ServiceID = serviceID
	in.PlanID = planID
	in.Plan = planName
	if in.ID == "" {
		in.ID = uuid.NewV4().String()
	}

	log.Printf("Instantiating service %s using service/plan %s : %s\n", in.Name, serviceID, planID)

	err = c.createServiceInstance(in)
	op := scmodel.LastOperation{}
	if err != nil {
		op.State = "FAILED"
		op.Description = err.Error()
		log.Printf("Failed to create service instance: %v\n", err)
	} else {
		op.State = "CREATED"
	}
	in.LastOperation = &op

	log.Printf("Updating Service %s with State\n%v\n", in.Name, in.LastOperation)
	return in, c.k8sStorage.SetService(in)
}

func (c *Controller) CreateServiceBinding(in *scmodel.ServiceBinding) (*scmodel.Credential, error) {
	log.Printf("Creating Service Binding: %v\n", in)

	// Get instance information for service being bound to.
	to, err := c.k8sStorage.GetService(DEFAULT_NAMESPACE, in.To)
	if err != nil {
		log.Printf("To service does not exist %s: %v\n", in.To, err)
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
		log.Printf("Failed to marshal: %#v\n", err)
		return nil, err
	}

	in.ID = uuid.NewV4().String()

	st, err := c.k8sStorage.GetServiceType(to.Service)
	if err != nil {
		log.Printf("Failed to fetch service type %s : %v\n", to.Service, err)
		return nil, err
	}
	broker, err := c.k8sStorage.GetBroker(st.Broker)
	if err != nil {
		log.Printf("Error fetching broker for service: %s : %v\n", to.Service, err)
		return nil, err
	}
	url := fmt.Sprintf(BIND_FMT_STR, broker.BrokerURL, to.ID, in.ID)

	// TODO: Handle the auth
	createHttpReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	client := &http.Client{}
	log.Printf("Doing a request to: %s\n", url)
	resp, err := client.Do(createHttpReq)
	if err != nil {
		log.Printf("Failed to PUT: %#v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	sbr := scmodel.CreateServiceBindingResponse{}
	err = utils.ResponseBodyToObject(resp, &sbr)
	if err != nil {
		log.Printf("Failed to unmarshal: %#v\n", err)
		return nil, err
	}

	// Stash the credentials with the binding and update the binding.
	in.Credentials = sbr.Credentials

	err = c.k8sStorage.UpdateServiceBinding(in)
	if err != nil {
		log.Printf("Failed to update service binding %s : %v\n", in.Name, err)
		return nil, err
	}

	// If FROM already exists, we need to update it here...
	fromSI, err := c.k8sStorage.GetService(DEFAULT_NAMESPACE, in.From)
	if err == nil && fromSI != nil {
		// Update the Service Instance with the new bindings
		log.Printf("Found existing FROM Service: %s, should update it\n", fromSI.Name)
		err = c.updateServiceInstance(fromSI)
		if err != nil {
			log.Printf("Failed to update existing FROM service %s : %v\n", fromSI.Name, err)
			return nil, err
		}
	}
	return &in.Credentials, nil
}

func (c *Controller) CreateServiceBroker(in *scmodel.ServiceBroker) (*scmodel.ServiceBroker, error) {
	// Fetch the catalog from the broker
	u := fmt.Sprintf(CATALOG_URL_FMT_STR, in.BrokerURL)
	req, err := http.NewRequest("GET", u, nil)
	req.SetBasicAuth(in.AuthUsername, in.AuthPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch catalog from %s\n%v\n", u, resp)
		log.Printf("err: %#v\n", err)
		return nil, err
	}

	// TODO: the model from SB is fetched and stored directly as the one in the SC model (which the
	// storage operates on). We should convert it from the SB model to SC model before storing.
	var catalog scmodel.Catalog
	err = utils.ResponseBodyToObject(resp, &catalog)
	if err != nil {
		log.Printf("Failed to unmarshal catalog: %#v\n", err)
		return nil, err
	}

	log.Printf("Adding a broker %s catalog:\n%v\n", in.Name, catalog)

	err = c.k8sStorage.AddBroker(in, &catalog)
	if err != nil {
		return nil, err
	}
	return in, nil
}
