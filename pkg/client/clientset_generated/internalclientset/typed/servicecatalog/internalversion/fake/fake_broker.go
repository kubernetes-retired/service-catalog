/*
Copyright 2017 The Kubernetes Authors.

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

package fake

import (
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	api "k8s.io/kubernetes/pkg/api"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	schema "k8s.io/kubernetes/pkg/runtime/schema"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakeBrokers implements BrokerInterface
type FakeBrokers struct {
	Fake *FakeServicecatalog
}

var brokersResource = schema.GroupVersionResource{Group: "", Version: "", Resource: "brokers"}

func (c *FakeBrokers) Create(broker *servicecatalog.Broker) (result *servicecatalog.Broker, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootCreateAction(brokersResource, broker), &servicecatalog.Broker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.Broker), err
}

func (c *FakeBrokers) Update(broker *servicecatalog.Broker) (result *servicecatalog.Broker, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootUpdateAction(brokersResource, broker), &servicecatalog.Broker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.Broker), err
}

func (c *FakeBrokers) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewRootDeleteAction(brokersResource, name), &servicecatalog.Broker{})
	return err
}

func (c *FakeBrokers) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewRootDeleteCollectionAction(brokersResource, listOptions)

	_, err := c.Fake.Invokes(action, &servicecatalog.BrokerList{})
	return err
}

func (c *FakeBrokers) Get(name string) (result *servicecatalog.Broker, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootGetAction(brokersResource, name), &servicecatalog.Broker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.Broker), err
}

func (c *FakeBrokers) List(opts api.ListOptions) (result *servicecatalog.BrokerList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootListAction(brokersResource, opts), &servicecatalog.BrokerList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &servicecatalog.BrokerList{}
	for _, item := range obj.(*servicecatalog.BrokerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested brokers.
func (c *FakeBrokers) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewRootWatchAction(brokersResource, opts))
}

// Patch applies the patch and returns the patched broker.
func (c *FakeBrokers) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *servicecatalog.Broker, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootPatchSubresourceAction(brokersResource, name, data, subresources...), &servicecatalog.Broker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.Broker), err
}
