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

package mem

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/storage"
)

type memStorage struct {
	brokers        *memStorageBroker
	serviceClasses *memStorageServiceClass
	instances      map[string]*memStorageInstance
	bindings       map[string]*memStorageBinding
}

// NewStorage creates an instance of Storage interface, backed by memory.
func NewStorage() storage.Storage {
	return &memStorage{
		brokers:        newMemStorageBroker(),
		serviceClasses: newMemStorageServiceClass(),
		instances:      make(map[string]*memStorageInstance),
		bindings:       make(map[string]*memStorageBinding),
	}
}

// NewPopulatedStorage is the equivalent of NewStorage, except pre-populataes the returned
// in-memory storage instance with brokers and service classes
func NewPopulatedStorage(
	brokers map[string]*servicecatalog.Broker,
	serviceClasses map[string]*servicecatalog.ServiceClass,
) storage.Storage {
	brokerStorage := newMemStorageBroker()
	brokerStorage.brokers = brokers
	serviceClassStorage := newMemStorageServiceClass()
	serviceClassStorage.classes = serviceClasses
	return &memStorage{
		brokers:        brokerStorage,
		serviceClasses: serviceClassStorage,
		instances:      make(map[string]*memStorageInstance),
		bindings:       make(map[string]*memStorageBinding),
	}
}

func (m *memStorage) Brokers() storage.BrokerStorage {
	return m.brokers
}

func (m *memStorage) ServiceClasses() storage.ServiceClassStorage {
	return m.serviceClasses
}

func (m *memStorage) Instances(ns string) storage.InstanceStorage {
	ret, ok := m.instances[ns]
	if !ok {
		ret = newMemStorageInstance()
		m.instances[ns] = ret
	}
	return ret
}

func (m *memStorage) Bindings(ns string) storage.BindingStorage {
	ret, ok := m.bindings[ns]
	if !ok {
		ret = newMemStorageBinding()
		m.bindings[ns] = ret
	}
	return ret
}
