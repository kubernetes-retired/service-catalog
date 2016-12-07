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

package storage

import (
	"fmt"
	"log"

	model "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

type servicePlanNotFound struct {
	service string
	plan    string
}

func (e servicePlanNotFound) Error() string {
	return fmt.Sprintf("Can't find a service/plan: %s/%s", e.service, e.plan)
}

type serviceNotFound struct {
	service string
}

func (e serviceNotFound) Error() string {
	return fmt.Sprintf("Can't find the service with id: %s", e.service)
}

// GetServicePlanInfo fetches the GUIDs for Service and Plan, also returns the
// name of the plan since it might get defaulted.
//
// If Plan is not given and there's only one plan for a given service, we'll
// choose that.
func GetServicePlanInfo(storage Storage, service string, plan string) (string, string, string, error) {
	s, err := storage.GetServiceClass(service)
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
	return "", "", "", servicePlanNotFound{service, plan}
}

// GetBrokerByServiceClass returns the broker which serves a particular service
// class.
func GetBrokerByServiceClass(storage Storage, id string) (*model.ServiceBroker, error) {
	log.Printf("Getting broker by service id %s\n", id)

	c, err := storage.GetInventory()
	if err != nil {
		return nil, err
	}
	for _, service := range c.Services {
		if service.ID == id {
			log.Printf("Found service type %s\n", service.Name)
			return storage.GetBroker(service.Broker)
		}
	}
	return nil, serviceNotFound{id}
}

// GetBindingsForService returns all the specific kinds of bindings (to, from, both).
func GetBindingsForService(storage Storage, serviceID string, t BindingDirection) ([]*model.ServiceBinding, error) {
	var ret []*model.ServiceBinding
	bindings, err := storage.ListServiceBindings()
	if err != nil {
		return nil, err
	}

	for _, b := range bindings {
		switch t {
		case Both:
			if b.From == serviceID || b.To == serviceID {
				ret = append(ret, b)
			}
		case From:
			if b.From == serviceID {
				ret = append(ret, b)
			}
		case To:
			if b.To == serviceID {
				ret = append(ret, b)
			}
		}
	}
	return ret, nil
}
