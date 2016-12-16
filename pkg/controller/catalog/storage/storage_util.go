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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
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
func GetServicePlanInfo(storage ServiceClassStorage, service string, plan string) (string, string, string, error) {
	s, err := storage.Get(service)
	if err != nil {
		return "", "", "", err
	}
	// No plan specified and only one plan, use it.
	if plan == "" && len(s.Plans) == 1 {
		planID := s.Plans[0].CFGUID
		planName := s.Plans[0].Name
		log.Printf("Found Service Plan GUID as %s for %s : %s", planID, service, planName)
		return s.CFGUID, planID, planName, nil
	}
	for _, p := range s.Plans {
		if p.Name == plan {
			planID := p.CFGUID
			log.Printf("Found Service Plan GUID as %s for %s : %s", planID, service, plan)
			return s.CFGUID, planID, p.Name, nil
		}
	}
	return "", "", "", servicePlanNotFound{service, plan}
}

// GetBrokerByServiceClass returns the broker which serves a particular service
// class.
func GetBrokerByServiceClass(
	brokerStorage BrokerStorage,
	svcClassStorage ServiceClassStorage,
	id string,
) (*servicecatalog.Broker, error) {

	log.Printf("Getting broker by service id %s\n", id)

	svcList, err := svcClassStorage.List()
	if err != nil {
		return nil, err
	}
	for _, service := range svcList {
		if service.CFGUID == id {
			log.Printf("Found service type %s\n", service.Name)
			return brokerStorage.Get(service.BrokerName)
		}
	}
	return nil, serviceNotFound{id}
}
