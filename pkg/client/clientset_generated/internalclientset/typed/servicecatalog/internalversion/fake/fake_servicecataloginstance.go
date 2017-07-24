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

// FakeServiceCatalogInstances implements ServiceCatalogInstanceInterface
type FakeServiceCatalogInstances struct {
	Fake *FakeServicecatalog
	ns   string
}

var servicecataloginstancesResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "", Resource: "servicecataloginstances"}

var servicecataloginstancesKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "", Kind: "ServiceCatalogInstance"}

func (c *FakeServiceCatalogInstances) Create(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (result *servicecatalog.ServiceCatalogInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicecataloginstancesResource, c.ns, serviceCatalogInstance), &servicecatalog.ServiceCatalogInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogInstance), err
}

func (c *FakeServiceCatalogInstances) Update(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (result *servicecatalog.ServiceCatalogInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicecataloginstancesResource, c.ns, serviceCatalogInstance), &servicecatalog.ServiceCatalogInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogInstance), err
}

func (c *FakeServiceCatalogInstances) UpdateStatus(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (*servicecatalog.ServiceCatalogInstance, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(servicecataloginstancesResource, "status", c.ns, serviceCatalogInstance), &servicecatalog.ServiceCatalogInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogInstance), err
}

func (c *FakeServiceCatalogInstances) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicecataloginstancesResource, c.ns, name), &servicecatalog.ServiceCatalogInstance{})

	return err
}

func (c *FakeServiceCatalogInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicecataloginstancesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &servicecatalog.ServiceCatalogInstanceList{})
	return err
}

func (c *FakeServiceCatalogInstances) Get(name string, options v1.GetOptions) (result *servicecatalog.ServiceCatalogInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicecataloginstancesResource, c.ns, name), &servicecatalog.ServiceCatalogInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogInstance), err
}

func (c *FakeServiceCatalogInstances) List(opts v1.ListOptions) (result *servicecatalog.ServiceCatalogInstanceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicecataloginstancesResource, servicecataloginstancesKind, c.ns, opts), &servicecatalog.ServiceCatalogInstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &servicecatalog.ServiceCatalogInstanceList{}
	for _, item := range obj.(*servicecatalog.ServiceCatalogInstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceCatalogInstances.
func (c *FakeServiceCatalogInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicecataloginstancesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched serviceCatalogInstance.
func (c *FakeServiceCatalogInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogInstance, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicecataloginstancesResource, c.ns, name, data, subresources...), &servicecatalog.ServiceCatalogInstance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceCatalogInstance), err
}
