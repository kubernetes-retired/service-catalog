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
	"fmt"
	"log"
	"strings"

	model "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

type BindingPair struct {
	Binding    *model.ServiceBinding
	Credential *model.Credential
}

type InMemServiceStorage struct {
	brokers map[string]*model.ServiceBroker
	// This gets fetched when a SB is created (or possibly later when refetched).
	// It's static for now to keep compatibility, seems like this could be more dynamic.
	catalogs map[string]*model.Catalog
	// maps instance ID to instance
	services map[string]*model.ServiceInstance
	// maps binding ID to binding
	// TODO: support looking up all bindings for a service instance.
	bindings map[string]*BindingPair
}

var _ ServiceStorage = (*InMemServiceStorage)(nil)

func CreateInMemServiceStorage() ServiceStorage {
	return &InMemServiceStorage{
		brokers:  make(map[string]*model.ServiceBroker),
		catalogs: make(map[string]*model.Catalog),
		services: make(map[string]*model.ServiceInstance),
		bindings: make(map[string]*BindingPair),
	}
}

func (s *InMemServiceStorage) GetInventory() (*model.Catalog, error) {
	services := []*model.Service{}
	for _, v := range s.catalogs {
		services = append(services, v.Services...)
	}
	return &model.Catalog{Services: services}, nil
}

func (s *InMemServiceStorage) ListBrokers() ([]*model.ServiceBroker, error) {
	b := []*model.ServiceBroker{}
	for _, v := range s.brokers {
		b = append(b, v)
	}
	return b, nil
}

func (s *InMemServiceStorage) GetBroker(id string) (*model.ServiceBroker, error) {
	if b, ok := s.brokers[id]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("No such broker: %s", id)
}

func (s *InMemServiceStorage) GetBrokerByService(id string) (*model.ServiceBroker, error) {
	for k, v := range s.catalogs {
		for _, service := range v.Services {
			if service.ID == id {
				return s.brokers[k], nil
			}
		}
	}
	return nil, fmt.Errorf("No service matching ID %s", id)
}

func (s *InMemServiceStorage) AddBroker(broker *model.ServiceBroker, catalog *model.Catalog) error {
	if _, ok := s.brokers[broker.GUID]; ok {
		return fmt.Errorf("Broker %s already exists", broker.Name)
	}
	s.brokers[broker.GUID] = broker
	s.catalogs[broker.GUID] = catalog
	return nil
}

func (s *InMemServiceStorage) UpdateBroker(broker *model.ServiceBroker, catalog *model.Catalog) error {
	if _, ok := s.brokers[broker.GUID]; !ok {
		return fmt.Errorf("Broker %s does not exist", broker.Name)
	}
	s.brokers[broker.GUID] = broker
	s.catalogs[broker.GUID] = catalog
	return nil
}

func (s *InMemServiceStorage) DeleteBroker(id string) error {
	_, err := s.GetBroker(id)
	if err != nil {
		return fmt.Errorf("Broker %s does not exist", id)
	}
	delete(s.brokers, id)
	delete(s.catalogs, id)

	// TODO: Delete bindings too.
	return nil
}

func (s *InMemServiceStorage) GetServiceType(name string) (*model.Service, error) {
	c, err := s.GetInventory()
	if err != nil {
		return nil, err
	}
	for _, serviceType := range c.Services {
		if serviceType.Name == name {
			return serviceType, nil
		}
	}
	return nil, fmt.Errorf("ServiceType %s not found", name)
}

// ServiceExists returns true if service exists
// Only supports "default" namespace for now.
func (s *InMemServiceStorage) ServiceExists(ns string, id string) bool {
	_, err := s.GetService(ns, id)
	return err == nil
}

func (s *InMemServiceStorage) ListServices(ns string) ([]*model.ServiceInstance, error) {
	services := []*model.ServiceInstance{}
	for _, v := range s.services {
		services = append(services, v)
	}
	return services, nil
}

func (s *InMemServiceStorage) GetService(ns string, id string) (*model.ServiceInstance, error) {
	service, ok := s.services[id]
	if !ok {
		return &model.ServiceInstance{}, fmt.Errorf("Service %s does not exist", id)
	}

	return service, nil
}

func (s *InMemServiceStorage) AddService(si *model.ServiceInstance) error {
	if s.ServiceExists("default", si.ID) {
		return fmt.Errorf("Service %s already exists", si.ID)
	}

	s.services[si.ID] = si
	return nil
}

func (s *InMemServiceStorage) SetService(si *model.ServiceInstance) error {
	s.services[si.ID] = si
	return nil
}

func (s *InMemServiceStorage) DeleteService(id string) error {
	// First delete all the bindings where this ID is either to / from
	bindings, err := s.GetBindingsForService(id, Both)
	for _, b := range bindings {
		err = s.DeleteServiceBinding(b.ID)
		if err != nil {
			return err
		}
	}
	delete(s.services, id)
	return nil
}

func (s *InMemServiceStorage) ListServiceBindings() ([]*model.ServiceBinding, error) {
	bindings := []*model.ServiceBinding{}
	for _, v := range s.bindings {
		bindings = append(bindings, v.Binding)
	}
	return bindings, nil
}

func (s *InMemServiceStorage) GetServiceBinding(id string) (*model.ServiceBinding, error) {
	b, ok := s.bindings[id]
	if !ok {
		return &model.ServiceBinding{}, fmt.Errorf("Binding %s does not exist", id)
	}

	return b.Binding, nil
}

func (s *InMemServiceStorage) AddServiceBinding(binding *model.ServiceBinding, cred *model.Credential) error {
	_, err := s.GetServiceBinding(binding.ID)
	if err == nil {
		return fmt.Errorf("Binding %s already exists", binding.ID)
	}

	s.bindings[binding.ID] = &BindingPair{Binding: binding, Credential: cred}
	return nil
}

func (s *InMemServiceStorage) UpdateServiceBinding(binding *model.ServiceBinding) error {
	_, err := s.GetServiceBinding(binding.ID)
	if err != nil {
		return fmt.Errorf("Binding %s doesn't exist", binding.ID)
	}

	// TODO(vaikas): Fix
	s.bindings[binding.ID] = &BindingPair{Binding: binding, Credential: nil}
	return nil
}

func (s *InMemServiceStorage) DeleteServiceBinding(id string) error {
	log.Printf("Deleting binding: %s\n", id)
	delete(s.bindings, id)
	return nil
}

func (s *InMemServiceStorage) getServiceInstanceByName(name string) (*model.ServiceInstance, error) {
	siList, err := s.ListServices("default")
	if err != nil {
		return nil, err
	}

	for _, si := range siList {
		if strings.Compare(si.Name, name) == 0 {
			return si, nil
		}
	}

	return nil, fmt.Errorf("Service instance %s was not found", name)
}

// GetBindingsForService returns all the specific kinds of bindings (to, from, both).
func (s *InMemServiceStorage) GetBindingsForService(serviceID string, t BindingDirection) ([]*model.ServiceBinding, error) {
	var ret []*model.ServiceBinding
	bindings, err := s.ListServiceBindings()
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
