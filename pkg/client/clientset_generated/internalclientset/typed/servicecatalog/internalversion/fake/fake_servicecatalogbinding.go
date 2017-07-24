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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceCatalogBindings implements ServiceCatalogBindingInterface
type FakeServiceCatalogBindings struct {
	Fake *FakeServicecatalog
	ns   string
}

var servicecatalogbindingsResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "", Resource: "servicecatalogbindings"}

var servicecatalogbindingsKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "", Kind: "ServiceCatalogBinding"}

func (c *FakeServiceCatalogBindings) Create(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (result *servicecatalog.ServiceCatalogBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicecatalogbindingsResource, c.ns, serviceCatalogBinding), &servicecatalog.ServiceCatalogBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogBinding), err
}

func (c *FakeServiceCatalogBindings) Update(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (result *servicecatalog.ServiceCatalogBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicecatalogbindingsResource, c.ns, serviceCatalogBinding), &servicecatalog.ServiceCatalogBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogBinding), err
}

func (c *FakeServiceCatalogBindings) UpdateStatus(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (*servicecatalog.ServiceCatalogBinding, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(servicecatalogbindingsResource, "status", c.ns, serviceCatalogBinding), &servicecatalog.ServiceCatalogBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogBinding), err
}

func (c *FakeServiceCatalogBindings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicecatalogbindingsResource, c.ns, name), &servicecatalog.ServiceCatalogBinding{})

	return err
}

func (c *FakeServiceCatalogBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicecatalogbindingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &servicecatalog.ServiceCatalogBindingList{})
	return err
}

func (c *FakeServiceCatalogBindings) Get(name string, options v1.GetOptions) (result *servicecatalog.ServiceCatalogBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicecatalogbindingsResource, c.ns, name), &servicecatalog.ServiceCatalogBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogBinding), err
}

func (c *FakeServiceCatalogBindings) List(opts v1.ListOptions) (result *servicecatalog.ServiceCatalogBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicecatalogbindingsResource, servicecatalogbindingsKind, c.ns, opts), &servicecatalog.ServiceCatalogBindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &servicecatalog.ServiceCatalogBindingList{}
	for _, item := range obj.(*servicecatalog.ServiceCatalogBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceCatalogBindings.
func (c *FakeServiceCatalogBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicecatalogbindingsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched serviceCatalogBinding.
func (c *FakeServiceCatalogBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicecatalogbindingsResource, c.ns, name, data, subresources...), &servicecatalog.ServiceCatalogBinding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogBinding), err
}
