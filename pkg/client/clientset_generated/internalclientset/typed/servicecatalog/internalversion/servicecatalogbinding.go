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
	scheme "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceCatalogBindingsGetter has a method to return a ServiceCatalogBindingInterface.
// A group's client should implement this interface.
type ServiceCatalogBindingsGetter interface {
	ServiceCatalogBindings(namespace string) ServiceCatalogBindingInterface
}

// ServiceCatalogBindingInterface has methods to work with ServiceCatalogBinding resources.
type ServiceCatalogBindingInterface interface {
	Create(*servicecatalog.ServiceCatalogBinding) (*servicecatalog.ServiceCatalogBinding, error)
	Update(*servicecatalog.ServiceCatalogBinding) (*servicecatalog.ServiceCatalogBinding, error)
	UpdateStatus(*servicecatalog.ServiceCatalogBinding) (*servicecatalog.ServiceCatalogBinding, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*servicecatalog.ServiceCatalogBinding, error)
	List(opts v1.ListOptions) (*servicecatalog.ServiceCatalogBindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogBinding, err error)
	ServiceCatalogBindingExpansion
}

// serviceCatalogBindings implements ServiceCatalogBindingInterface
type serviceCatalogBindings struct {
	client rest.Interface
	ns     string
}

// newServiceCatalogBindings returns a ServiceCatalogBindings
func newServiceCatalogBindings(c *ServicecatalogClient, namespace string) *serviceCatalogBindings {
	return &serviceCatalogBindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a serviceCatalogBinding and creates it.  Returns the server's representation of the serviceCatalogBinding, and an error, if there is any.
func (c *serviceCatalogBindings) Create(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (result *servicecatalog.ServiceCatalogBinding, err error) {
	result = &servicecatalog.ServiceCatalogBinding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		Body(serviceCatalogBinding).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceCatalogBinding and updates it. Returns the server's representation of the serviceCatalogBinding, and an error, if there is any.
func (c *serviceCatalogBindings) Update(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (result *servicecatalog.ServiceCatalogBinding, err error) {
	result = &servicecatalog.ServiceCatalogBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		Name(serviceCatalogBinding.Name).
		Body(serviceCatalogBinding).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *serviceCatalogBindings) UpdateStatus(serviceCatalogBinding *servicecatalog.ServiceCatalogBinding) (result *servicecatalog.ServiceCatalogBinding, err error) {
	result = &servicecatalog.ServiceCatalogBinding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		Name(serviceCatalogBinding.Name).
		SubResource("status").
		Body(serviceCatalogBinding).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceCatalogBinding and deletes it. Returns an error if one occurs.
func (c *serviceCatalogBindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceCatalogBindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the serviceCatalogBinding, and returns the corresponding serviceCatalogBinding object, and an error if there is any.
func (c *serviceCatalogBindings) Get(name string, options v1.GetOptions) (result *servicecatalog.ServiceCatalogBinding, err error) {
	result = &servicecatalog.ServiceCatalogBinding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceCatalogBindings that match those selectors.
func (c *serviceCatalogBindings) List(opts v1.ListOptions) (result *servicecatalog.ServiceCatalogBindingList, err error) {
	result = &servicecatalog.ServiceCatalogBindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceCatalogBindings.
func (c *serviceCatalogBindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched serviceCatalogBinding.
func (c *serviceCatalogBindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogBinding, err error) {
	result = &servicecatalog.ServiceCatalogBinding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servicecatalogbindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
