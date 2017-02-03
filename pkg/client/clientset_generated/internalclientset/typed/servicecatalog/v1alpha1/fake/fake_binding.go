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
	api "k8s.io/kubernetes/pkg/api"
	v1 "k8s.io/kubernetes/pkg/api/v1"
	core "k8s.io/kubernetes/pkg/client/testing/core"
	labels "k8s.io/kubernetes/pkg/labels"
	schema "k8s.io/kubernetes/pkg/runtime/schema"
	watch "k8s.io/kubernetes/pkg/watch"
)

// FakeBindings implements BindingInterface
type FakeBindings struct {
	Fake *FakeServicecatalogV1alpha1
	ns   string
}

var bindingsResource = schema.GroupVersionResource{Group: "servicecatalog", Version: "v1alpha1", Resource: "bindings"}

func (c *FakeBindings) Create(binding *v1alpha1.Binding) (result *v1alpha1.Binding, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(bindingsResource, c.ns, binding), &v1alpha1.Binding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Binding), err
}

func (c *FakeBindings) Update(binding *v1alpha1.Binding) (result *v1alpha1.Binding, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(bindingsResource, c.ns, binding), &v1alpha1.Binding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Binding), err
}

func (c *FakeBindings) UpdateStatus(binding *v1alpha1.Binding) (*v1alpha1.Binding, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction(bindingsResource, "status", c.ns, binding), &v1alpha1.Binding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Binding), err
}

func (c *FakeBindings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(bindingsResource, c.ns, name), &v1alpha1.Binding{})

	return err
}

func (c *FakeBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := core.NewDeleteCollectionAction(bindingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BindingList{})
	return err
}

func (c *FakeBindings) Get(name string) (result *v1alpha1.Binding, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(bindingsResource, c.ns, name), &v1alpha1.Binding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Binding), err
}

func (c *FakeBindings) List(opts v1.ListOptions) (result *v1alpha1.BindingList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(bindingsResource, c.ns, opts), &v1alpha1.BindingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BindingList{}
	for _, item := range obj.(*v1alpha1.BindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bindings.
func (c *FakeBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(bindingsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched binding.
func (c *FakeBindings) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1alpha1.Binding, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(bindingsResource, c.ns, name, data, subresources...), &v1alpha1.Binding{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Binding), err
}
