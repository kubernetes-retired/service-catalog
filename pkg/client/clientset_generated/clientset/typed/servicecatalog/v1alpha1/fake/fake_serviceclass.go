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

// FakeServiceClasses implements ServiceClassInterface
type FakeServiceClasses struct {
	Fake *FakeServicecatalogV1alpha1
	ns   string
}

var serviceclassesResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "serviceclasses"}

func (c *FakeServiceClasses) Create(serviceClass *v1alpha1.ServiceClass) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(serviceclassesResource, c.ns, serviceClass), &v1alpha1.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) Update(serviceClass *v1alpha1.ServiceClass) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(serviceclassesResource, c.ns, serviceClass), &v1alpha1.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(serviceclassesResource, c.ns, name), &v1alpha1.ServiceClass{})

	return err
}

func (c *FakeServiceClasses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := core.NewDeleteCollectionAction(serviceclassesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ServiceClassList{})
	return err
}

func (c *FakeServiceClasses) Get(name string) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(serviceclassesResource, c.ns, name), &v1alpha1.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}

func (c *FakeServiceClasses) List(opts v1.ListOptions) (result *v1alpha1.ServiceClassList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(serviceclassesResource, c.ns, opts), &v1alpha1.ServiceClassList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ServiceClassList{}
	for _, item := range obj.(*v1alpha1.ServiceClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceClasses.
func (c *FakeServiceClasses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(serviceclassesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched serviceClass.
func (c *FakeServiceClasses) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(serviceclassesResource, c.ns, name, data, subresources...), &v1alpha1.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ServiceClass), err
}
