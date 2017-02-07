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

package internalversion

import (
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	api "k8s.io/kubernetes/pkg/api"
	restclient "k8s.io/kubernetes/pkg/client/restclient"
	watch "k8s.io/kubernetes/pkg/watch"
)

// ServiceClassesGetter has a method to return a ServiceClassInterface.
// A group's client should implement this interface.
type ServiceClassesGetter interface {
	ServiceClasses(namespace string) ServiceClassInterface
}

// ServiceClassInterface has methods to work with ServiceClass resources.
type ServiceClassInterface interface {
	Create(*servicecatalog.ServiceClass) (*servicecatalog.ServiceClass, error)
	Update(*servicecatalog.ServiceClass) (*servicecatalog.ServiceClass, error)
	Delete(name string, options *api.DeleteOptions) error
	DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error
	Get(name string) (*servicecatalog.ServiceClass, error)
	List(opts api.ListOptions) (*servicecatalog.ServiceClassList, error)
	Watch(opts api.ListOptions) (watch.Interface, error)
	Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceClass, err error)
	ServiceClassExpansion
}

// serviceClasses implements ServiceClassInterface
type serviceClasses struct {
	client restclient.Interface
	ns     string
}

// newServiceClasses returns a ServiceClasses
func newServiceClasses(c *ServicecatalogClient, namespace string) *serviceClasses {
	return &serviceClasses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a serviceClass and creates it.  Returns the server's representation of the serviceClass, and an error, if there is any.
func (c *serviceClasses) Create(serviceClass *servicecatalog.ServiceClass) (result *servicecatalog.ServiceClass, err error) {
	result = &servicecatalog.ServiceClass{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("serviceclasses").
		Body(serviceClass).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceClass and updates it. Returns the server's representation of the serviceClass, and an error, if there is any.
func (c *serviceClasses) Update(serviceClass *servicecatalog.ServiceClass) (result *servicecatalog.ServiceClass, err error) {
	result = &servicecatalog.ServiceClass{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(serviceClass.Name).
		Body(serviceClass).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceClass and deletes it. Returns an error if one occurs.
func (c *serviceClasses) Delete(name string, options *api.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceClasses) DeleteCollection(options *api.DeleteOptions, listOptions api.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&listOptions, api.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the serviceClass, and returns the corresponding serviceClass object, and an error if there is any.
func (c *serviceClasses) Get(name string) (result *servicecatalog.ServiceClass, err error) {
	result = &servicecatalog.ServiceClass{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceclasses").
		Name(name).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceClasses that match those selectors.
func (c *serviceClasses) List(opts api.ListOptions) (result *servicecatalog.ServiceClassList, err error) {
	result = &servicecatalog.ServiceClassList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&opts, api.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceClasses.
func (c *serviceClasses) Watch(opts api.ListOptions) (watch.Interface, error) {
	return c.client.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("serviceclasses").
		VersionedParams(&opts, api.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched serviceClass.
func (c *serviceClasses) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceClass, err error) {
	result = &servicecatalog.ServiceClass{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("serviceclasses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
