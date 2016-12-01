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

	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
)

type inMemoryStorage struct {
	services map[string]*model.Service
}

// CreateInMemoryStorage creates a storage backed by memory.
func CreateInMemoryStorage() Storage {
	return &inMemoryStorage{
		services: make(map[string]*model.Service),
	}
}

func (s *inMemoryStorage) ListServices() ([]*model.Service, error) {
	ret := make([]*model.Service, len(s.services))

	i := 0
	for _, v := range s.services {
		ret[i] = v
		i++
	}

	return ret, nil
}

func (s *inMemoryStorage) GetService(id string) (*model.Service, error) {
	ret, ok := s.services[id]
	if !ok {
		return &model.Service{}, fmt.Errorf("Service %s does not exist", id)
	}

	return ret, nil
}

func (s *inMemoryStorage) CreateService(service *model.Service) error {
	if _, ok := s.services[service.Name]; ok {
		return fmt.Errorf("Service '%s' already exists", service.Name)
	}

	s.services[service.ID] = service
	return nil
}

func (s *inMemoryStorage) DeleteService(id string) error {
	if _, ok := s.services[id]; !ok {
		return fmt.Errorf("Service '%s' not found", id)
	}

	delete(s.services, id)
	return nil
}
