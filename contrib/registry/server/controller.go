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
	"encoding/json"
	"errors"
	"io/ioutil"

	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
)

// Controller is an interface which implements the registry functionality.
type Controller interface {
	ListServices() ([]*model.Service, error)
	GetService(serviceID string) (*model.Service, error)
	CreateService(s *model.Service) error
	DeleteService(serviceID string) error
}

type registryController struct {
	storage Storage
}

// CreateController creates an instance of the Controller interface, given
// a storage implementation and a registry definition file.
func CreateController(storage Storage, defFile string) (Controller, error) {
	if defFile != "" {
		if err := loadDefinitions(storage, defFile); err != nil {
			return nil, err
		}
	}

	return &registryController{
		storage: storage,
	}, nil
}

func loadDefinitions(s Storage, f string) error {
	j, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	var services []*model.Service
	if err := json.Unmarshal([]byte(j), &services); err != nil {
		return err
	}

	for _, service := range services {
		if err := s.CreateService(service); err != nil {
			return err
		}
	}

	return nil
}

func (c *registryController) ListServices() ([]*model.Service, error) {
	return c.storage.ListServices()
}

func (c *registryController) GetService(serviceID string) (*model.Service, error) {
	return c.storage.GetService(serviceID)
}

func (c *registryController) CreateService(s *model.Service) error {
	// Check that service does not already exist.
	if _, err := c.storage.GetService(s.Name); err == nil {
		return errors.New("Service already exists")
	}

	return c.storage.CreateService(s)
}

func (c *registryController) DeleteService(serviceID string) error {
	return c.storage.DeleteService(serviceID)
}
