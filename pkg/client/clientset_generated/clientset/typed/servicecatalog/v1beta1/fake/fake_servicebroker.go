/*
Copyright 2018 The Kubernetes Authors.

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
	v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceBrokers implements ServiceBrokerInterface
type FakeServiceBrokers struct {
	Fake *FakeServicecatalogV1beta1
	ns   string
}

var servicebrokersResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1beta1", Resource: "servicebrokers"}

var servicebrokersKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1beta1", Kind: "ServiceBroker"}

// Get takes name of the serviceBroker, and returns the corresponding serviceBroker object, and an error if there is any.
func (c *FakeServiceBrokers) Get(name string, options v1.GetOptions) (result *v1beta1.ServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicebrokersResource, c.ns, name), &v1beta1.ServiceBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ServiceBroker), err
}

// List takes label and field selectors, and returns the list of ServiceBrokers that match those selectors.
func (c *FakeServiceBrokers) List(opts v1.ListOptions) (result *v1beta1.ServiceBrokerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicebrokersResource, servicebrokersKind, c.ns, opts), &v1beta1.ServiceBrokerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ServiceBrokerList{}
	for _, item := range obj.(*v1beta1.ServiceBrokerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceBrokers.
func (c *FakeServiceBrokers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicebrokersResource, c.ns, opts))

}

// Create takes the representation of a serviceBroker and creates it.  Returns the server's representation of the serviceBroker, and an error, if there is any.
func (c *FakeServiceBrokers) Create(serviceBroker *v1beta1.ServiceBroker) (result *v1beta1.ServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicebrokersResource, c.ns, serviceBroker), &v1beta1.ServiceBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ServiceBroker), err
}

// Update takes the representation of a serviceBroker and updates it. Returns the server's representation of the serviceBroker, and an error, if there is any.
func (c *FakeServiceBrokers) Update(serviceBroker *v1beta1.ServiceBroker) (result *v1beta1.ServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicebrokersResource, c.ns, serviceBroker), &v1beta1.ServiceBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ServiceBroker), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeServiceBrokers) UpdateStatus(serviceBroker *v1beta1.ServiceBroker) (*v1beta1.ServiceBroker, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(servicebrokersResource, "status", c.ns, serviceBroker), &v1beta1.ServiceBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ServiceBroker), err
}

// Delete takes name of the serviceBroker and deletes it. Returns an error if one occurs.
func (c *FakeServiceBrokers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicebrokersResource, c.ns, name), &v1beta1.ServiceBroker{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceBrokers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicebrokersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ServiceBrokerList{})
	return err
}

// Patch applies the patch and returns the patched serviceBroker.
func (c *FakeServiceBrokers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicebrokersResource, c.ns, name, data, subresources...), &v1beta1.ServiceBroker{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ServiceBroker), err
}
