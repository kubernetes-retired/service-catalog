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

// FakeInstances implements InstanceInterface
type FakeInstances struct {
	Fake *FakeServicecatalogV1alpha1
	ns   string
}

var instancesResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "instances"}

func (c *FakeInstances) Create(instance *v1alpha1.Instance) (result *v1alpha1.Instance, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(instancesResource, c.ns, instance), &v1alpha1.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Instance), err
}

func (c *FakeInstances) Update(instance *v1alpha1.Instance) (result *v1alpha1.Instance, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(instancesResource, c.ns, instance), &v1alpha1.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Instance), err
}

func (c *FakeInstances) UpdateStatus(instance *v1alpha1.Instance) (*v1alpha1.Instance, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction(instancesResource, "status", c.ns, instance), &v1alpha1.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Instance), err
}

func (c *FakeInstances) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(instancesResource, c.ns, name), &v1alpha1.Instance{})

	return err
}

func (c *FakeInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := core.NewDeleteCollectionAction(instancesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.InstanceList{})
	return err
}

func (c *FakeInstances) Get(name string) (result *v1alpha1.Instance, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(instancesResource, c.ns, name), &v1alpha1.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Instance), err
}

func (c *FakeInstances) List(opts v1.ListOptions) (result *v1alpha1.InstanceList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(instancesResource, c.ns, opts), &v1alpha1.InstanceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.InstanceList{}
	for _, item := range obj.(*v1alpha1.InstanceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested instances.
func (c *FakeInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(instancesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched instance.
func (c *FakeInstances) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1alpha1.Instance, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(instancesResource, c.ns, name, data, subresources...), &v1alpha1.Instance{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Instance), err
}
