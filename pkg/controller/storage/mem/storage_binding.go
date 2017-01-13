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
	errBindingAlreadyExists = errors.New("binding already exists")
	errNoSuchBinding        = errors.New("no such binding")
)

type memStorageBinding struct {
	// maps binding ID to binding
	// TODO: support looking up all bindings for a service instance.
	bindings map[string]*servicecatalog.Binding
}

func newMemStorageBinding() *memStorageBinding {
	return &memStorageBinding{bindings: make(map[string]*servicecatalog.Binding)}
}

func (m *memStorageBinding) List() ([]*servicecatalog.Binding, error) {
	copy := make([]*servicecatalog.Binding, len(m.bindings))
	i := 0
	for _, bin := range m.bindings {
		copy[i] = &servicecatalog.Binding{}
		if err := deepCopy(copy[i], bin); err != nil {
			return nil, err
		}
		i++
	}
	return copy, nil
}

func (m *memStorageBinding) Get(name string) (*servicecatalog.Binding, error) {
	binding, ok := m.bindings[name]
	if !ok {
		return nil, errNoSuchInstance
	}
	copy := &servicecatalog.Binding{}
	if err := deepCopy(copy, binding); err != nil {
		return nil, err
	}
	return copy, nil
}

func (m *memStorageBinding) Create(bin *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	if _, err := m.Get(bin.Name); err == nil {
		return nil, errBindingAlreadyExists
	}
	copy1 := &servicecatalog.Binding{}
	if err := deepCopy(copy1, bin); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Binding{}
	if err := deepCopy(copy2, bin); err != nil {
		return nil, err
	}
	m.bindings[bin.Name] = copy1
	return copy2, nil
}

func (m *memStorageBinding) Update(bin *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	if _, err := m.Get(bin.Name); err != nil {
		return nil, errNoSuchBinding
	}
	copy1 := &servicecatalog.Binding{}
	if err := deepCopy(copy1, bin); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Binding{}
	if err := deepCopy(copy2, bin); err != nil {
		return nil, err
	}
	m.bindings[bin.Name] = copy1
	return copy2, nil
}

func (m *memStorageBinding) Delete(name string) error {
	if _, err := m.Get(name); err != nil {
		return errNoSuchBinding
	}
	delete(m.bindings, name)
	return nil
}
