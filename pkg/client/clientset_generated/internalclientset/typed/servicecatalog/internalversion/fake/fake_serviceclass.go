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

// FakeServiceClasses implements ServiceClassInterface
type FakeServiceClasses struct {
	Fake *FakeServicecatalog
	ns   string
}

var serviceclassesResource = schema.GroupVersionResource{Group: "servicecatalog", Version: "", Resource: "serviceclasses"}

func (c *FakeServiceClasses) Create(serviceClass *servicecatalog.ServiceClass) (result *servicecatalog.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(serviceclassesResource, c.ns, serviceClass), &servicecatalog.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceClass), err
}

func (c *FakeServiceClasses) Update(serviceClass *servicecatalog.ServiceClass) (result *servicecatalog.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(serviceclassesResource, c.ns, serviceClass), &servicecatalog.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceClass), err
}

func (c *FakeServiceClasses) Delete(name string, options *api.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(serviceclassesResource, c.ns, name), &servicecatalog.ServiceClass{})

	return err
}

func (c *FakeServiceClasses) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	action := core.NewDeleteCollectionAction(serviceclassesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &servicecatalog.ServiceClassList{})
	return err
}

func (c *FakeServiceClasses) Get(name string) (result *servicecatalog.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(serviceclassesResource, c.ns, name), &servicecatalog.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceClass), err
}

func (c *FakeServiceClasses) List(opts api.ListOptions) (result *servicecatalog.ServiceClassList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(serviceclassesResource, c.ns, opts), &servicecatalog.ServiceClassList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &servicecatalog.ServiceClassList{}
	for _, item := range obj.(*servicecatalog.ServiceClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceClasses.
func (c *FakeServiceClasses) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(serviceclassesResource, c.ns, opts))

}

// Patch applies the patch and returns the patched serviceClass.
func (c *FakeServiceClasses) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceClass, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(serviceclassesResource, c.ns, name, data, subresources...), &servicecatalog.ServiceClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*servicecatalog.ServiceClass), err
}
