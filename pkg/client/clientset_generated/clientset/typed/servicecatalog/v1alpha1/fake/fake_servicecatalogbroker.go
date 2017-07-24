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
	v1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeServiceCatalogBrokers implements ServiceCatalogBrokerInterface
type FakeServiceCatalogBrokers struct {
	Fake *FakeServicecatalogV1alpha1
}

var servicecatalogbrokersResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "servicecatalogbrokers"}

var servicecatalogbrokersKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Kind: "ServiceCatalogBroker"}

func (c *FakeServiceCatalogBrokers) Create(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (result *v1alpha1.ServiceCatalogBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(servicecatalogbrokersResource, serviceCatalogBroker), &v1alpha1.ServiceCatalogBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceCatalogBroker), err
}

func (c *FakeServiceCatalogBrokers) Update(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (result *v1alpha1.ServiceCatalogBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(servicecatalogbrokersResource, serviceCatalogBroker), &v1alpha1.ServiceCatalogBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceCatalogBroker), err
}

func (c *FakeServiceCatalogBrokers) UpdateStatus(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (*v1alpha1.ServiceCatalogBroker, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(servicecatalogbrokersResource, "status", serviceCatalogBroker), &v1alpha1.ServiceCatalogBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceCatalogBroker), err
}

func (c *FakeServiceCatalogBrokers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(servicecatalogbrokersResource, name), &v1alpha1.ServiceCatalogBroker{})
	return err
}

func (c *FakeServiceCatalogBrokers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(servicecatalogbrokersResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ServiceCatalogBrokerList{})
	return err
}

func (c *FakeServiceCatalogBrokers) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceCatalogBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(servicecatalogbrokersResource, name), &v1alpha1.ServiceCatalogBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceCatalogBroker), err
}

func (c *FakeServiceCatalogBrokers) List(opts v1.ListOptions) (result *v1alpha1.ServiceCatalogBrokerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(servicecatalogbrokersResource, servicecatalogbrokersKind, opts), &v1alpha1.ServiceCatalogBrokerList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ServiceCatalogBrokerList{}
	for _, item := range obj.(*v1alpha1.ServiceCatalogBrokerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceCatalogBrokers.
func (c *FakeServiceCatalogBrokers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(servicecatalogbrokersResource, opts))
}

// Patch applies the patch and returns the patched serviceCatalogBroker.
func (c *FakeServiceCatalogBrokers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceCatalogBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(servicecatalogbrokersResource, name, data, subresources...), &v1alpha1.ServiceCatalogBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceCatalogBroker), err
}
