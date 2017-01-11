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
	errNoSuchServiceClass = errors.New("no such service class")
)

type memStorageServiceClass struct {
	// This gets fetched when a SB is created (or possibly later when refetched).
	// It's static for now to keep compatibility, seems like this could be more dynamic.
	classes map[string]*servicecatalog.ServiceClass
}

func newMemStorageServiceClass() *memStorageServiceClass {
	return &memStorageServiceClass{classes: make(map[string]*servicecatalog.ServiceClass)}
}

func (m *memStorageServiceClass) List() ([]*servicecatalog.ServiceClass, error) {
	ret := make([]*servicecatalog.ServiceClass, len(m.classes))
	i := 0
	for _, val := range m.classes {
		ret[i] = val
		i++
	}
	return ret, nil
}

func (m *memStorageServiceClass) Get(name string) (*servicecatalog.ServiceClass, error) {
	cl, ok := m.classes[name]
	if !ok {
		return nil, errNoSuchServiceClass
	}
	return cl, nil
}

func (m *memStorageServiceClass) Create(sc *servicecatalog.ServiceClass) (*servicecatalog.ServiceClass, error) {
	if _, err := m.Get(sc.Name); err == nil {
		return nil, errInstanceAlreadyExists
	}
	m.classes[sc.Name] = sc
	return sc, nil
}
