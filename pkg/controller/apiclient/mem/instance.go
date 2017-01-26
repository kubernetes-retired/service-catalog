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
	errInstanceAlreadyExists = errors.New("instance already exists")
	errNoSuchInstance        = errors.New("no such instance")
)

type instanceClient struct {
	// maps instance ID to instance
	instances map[string]*servicecatalog.Instance
}

func newInstanceClient() *instanceClient {
	return &instanceClient{instances: make(map[string]*servicecatalog.Instance)}
}

func (c *instanceClient) List() ([]*servicecatalog.Instance, error) {
	copy := make([]*servicecatalog.Instance, len(c.instances))
	i := 0
	for _, inst := range c.instances {
		copy[i] = &servicecatalog.Instance{}
		if err := deepCopy(copy[i], inst); err != nil {
			return nil, err
		}
		i++
	}
	return copy, nil
}

func (c *instanceClient) Get(name string) (*servicecatalog.Instance, error) {
	inst, ok := c.instances[name]
	if !ok {
		return nil, errNoSuchInstance
	}
	copy := &servicecatalog.Instance{}
	if err := deepCopy(copy, inst); err != nil {
		return nil, err
	}
	return copy, nil
}

func (c *instanceClient) Create(inst *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	if _, err := c.Get(inst.Name); err == nil {
		return nil, errInstanceAlreadyExists
	}
	copy1 := &servicecatalog.Instance{}
	if err := deepCopy(copy1, inst); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Instance{}
	if err := deepCopy(copy2, inst); err != nil {
		return nil, err
	}
	c.instances[inst.Name] = copy1
	return copy2, nil
}

func (c *instanceClient) Update(inst *servicecatalog.Instance) (*servicecatalog.Instance, error) {
	if _, err := c.Get(inst.Name); err != nil {
		return nil, errNoSuchInstance
	}
	copy1 := &servicecatalog.Instance{}
	if err := deepCopy(copy1, inst); err != nil {
		return nil, err
	}
	copy2 := &servicecatalog.Instance{}
	if err := deepCopy(copy2, inst); err != nil {
		return nil, err
	}
	c.instances[inst.Name] = copy1
	return copy2, nil
}

func (c *instanceClient) Delete(name string) error {
	if _, err := c.Get(name); err != nil {
		return errNoSuchInstance
	}
	delete(c.instances, name)
	return nil
}
