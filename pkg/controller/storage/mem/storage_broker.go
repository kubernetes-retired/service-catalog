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
	"errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

var (
	errBrokerAlreadyExists = errors.New("broker already exists")
	errNoSuchBroker        = errors.New("no such broker")
)

type memStorageBroker struct {
	brokers map[string]*servicecatalog.Broker
}

func newMemStorageBroker() *memStorageBroker {
	return &memStorageBroker{brokers: make(map[string]*servicecatalog.Broker)}
}

func (m *memStorageBroker) List() ([]*servicecatalog.Broker, error) {
	copy := make([]*servicecatalog.Broker, len(m.brokers))
	i := 0
	for _, br := range m.brokers {
		copy[i] = &servicecatalog.Broker{}
		if err := deepCopy(copy[i], br); err != nil {
			return nil, err
		}
		i++
	}
	return copy, nil
}

func (m *memStorageBroker) Get(name string) (*servicecatalog.Broker, error) {
	br, ok := m.brokers[name]
	if !ok {
		return nil, errNoSuchBroker
	}
	copy := &servicecatalog.Broker{}
	if err := deepCopy(copy, br); err != nil {
		return nil, err
	}
	return copy, nil
}

func (m *memStorageBroker) Create(br *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	if _, err := m.Get(br.Name); err == nil {
		return nil, errBrokerAlreadyExists
	}
	copy1 := &servicecatalog.Broker{}
	if err := deepCopy(copy1, br); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Broker{}
	if err := deepCopy(copy2, br); err != nil {
		return nil, err
	}
	m.brokers[br.Name] = copy1
	return copy2, nil
}

func (m *memStorageBroker) Update(br *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	if _, err := m.Get(br.Name); err != nil {
		return nil, errNoSuchBroker
	}
	copy1 := &servicecatalog.Broker{}
	if err := deepCopy(copy1, br); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Broker{}
	if err := deepCopy(copy2, br); err != nil {
		return nil, err
	}
	m.brokers[br.Name] = copy1
	return copy2, nil
}

func (m *memStorageBroker) Delete(name string) error {
	if _, err := m.Get(name); err != nil {
		return errNoSuchBroker
	}
	delete(m.brokers, name)
	return nil
}
