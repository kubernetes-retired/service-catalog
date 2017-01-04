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
	"errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

var (
	errInstanceAlreadyExists = errors.New("instance already exists")
	errNoSuchInstance        = errors.New("no such instance")
)

type memStorageInstance struct {
	// maps instance ID to instance
	instances map[string]*servicecatalog.Instance
}

func newMemStorageInstance() *memStorageInstance {
	return &memStorageInstance{instances: make(map[string]*servicecatalog.Instance)}
}

func (m *memStorageInstance) List() ([]*servicecatalog.Instance, error) {
	ret := make([]*servicecatalog.Instance, len(m.instances))
	i := 0
	for _, val := range m.instances {
		ret[i] = val
		i++
	}
	return ret, nil
}

func (m *memStorageInstance) Get(name string) (*servicecatalog.Instance, error) {
	ret, ok := m.instances[name]
	if !ok {
		return nil, errNoSuchInstance
	}
	return ret, nil
}

func (m *memStorageInstance) Create(inst *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	if _, err := m.Get(inst.Name); err == nil {
		return nil, errInstanceAlreadyExists
	}
	m.instances[inst.Name] = inst
	return inst, nil
}

func (m *memStorageInstance) Update(inst *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	if _, err := m.Get(inst.Name); err != nil {
		return nil, errNoSuchInstance
	}
	m.instances[inst.Name] = inst
	return inst, nil
}

func (m *memStorageInstance) Delete(name string) error {
	if _, err := m.Get(name); err != nil {
		return errNoSuchInstance
	}
	delete(m.instances, name)
	return nil
}
