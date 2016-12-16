package storage

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
	ret := make([]*servicecatalog.Binding, len(m.bindings))
	i := 0
	for _, val := range m.bindings {
		ret[i] = val
		i++
	}
	return ret, nil
}
func (m *memStorageBinding) Get(name string) (*servicecatalog.Binding, error) {
	ret, ok := m.bindings[name]
	if !ok {
		return nil, errNoSuchBinding
	}
	return ret, nil
}
func (m *memStorageBinding) Create(bin *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	if _, err := m.Get(bin.Name); err == nil {
		return nil, errBindingAlreadyExists
	}
	m.bindings[bin.Name] = bin
	return bin, nil
}
func (m *memStorageBinding) Update(bin *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	if _, err := m.Get(bin.Name); err != nil {
		return nil, errNoSuchBinding
	}
	m.bindings[bin.Name] = bin
	return bin, nil
}
func (m *memStorageBinding) Delete(name string) error {
	if _, err := m.Get(name); err != nil {
		return errNoSuchBinding
	}
	delete(m.bindings, name)
	return nil
}
