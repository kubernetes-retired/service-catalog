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

// ServiceCatalogInstancesGetter has a method to return a ServiceCatalogInstanceInterface.
// A group's client should implement this interface.
type ServiceCatalogInstancesGetter interface {
	ServiceCatalogInstances(namespace string) ServiceCatalogInstanceInterface
}

// ServiceCatalogInstanceInterface has methods to work with ServiceCatalogInstance resources.
type ServiceCatalogInstanceInterface interface {
	Create(*servicecatalog.ServiceCatalogInstance) (*servicecatalog.ServiceCatalogInstance, error)
	Update(*servicecatalog.ServiceCatalogInstance) (*servicecatalog.ServiceCatalogInstance, error)
	UpdateStatus(*servicecatalog.ServiceCatalogInstance) (*servicecatalog.ServiceCatalogInstance, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*servicecatalog.ServiceCatalogInstance, error)
	List(opts v1.ListOptions) (*servicecatalog.ServiceCatalogInstanceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogInstance, err error)
	ServiceCatalogInstanceExpansion
}

// serviceCatalogInstances implements ServiceCatalogInstanceInterface
type serviceCatalogInstances struct {
	client rest.Interface
	ns     string
}

// newServiceCatalogInstances returns a ServiceCatalogInstances
func newServiceCatalogInstances(c *ServicecatalogClient, namespace string) *serviceCatalogInstances {
	return &serviceCatalogInstances{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a serviceCatalogInstance and creates it.  Returns the server's representation of the serviceCatalogInstance, and an error, if there is any.
func (c *serviceCatalogInstances) Create(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (result *servicecatalog.ServiceCatalogInstance, err error) {
	result = &servicecatalog.ServiceCatalogInstance{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		Body(serviceCatalogInstance).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceCatalogInstance and updates it. Returns the server's representation of the serviceCatalogInstance, and an error, if there is any.
func (c *serviceCatalogInstances) Update(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (result *servicecatalog.ServiceCatalogInstance, err error) {
	result = &servicecatalog.ServiceCatalogInstance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		Name(serviceCatalogInstance.Name).
		Body(serviceCatalogInstance).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *serviceCatalogInstances) UpdateStatus(serviceCatalogInstance *servicecatalog.ServiceCatalogInstance) (result *servicecatalog.ServiceCatalogInstance, err error) {
	result = &servicecatalog.ServiceCatalogInstance{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		Name(serviceCatalogInstance.Name).
		SubResource("status").
		Body(serviceCatalogInstance).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceCatalogInstance and deletes it. Returns an error if one occurs.
func (c *serviceCatalogInstances) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceCatalogInstances) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the serviceCatalogInstance, and returns the corresponding serviceCatalogInstance object, and an error if there is any.
func (c *serviceCatalogInstances) Get(name string, options v1.GetOptions) (result *servicecatalog.ServiceCatalogInstance, err error) {
	result = &servicecatalog.ServiceCatalogInstance{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceCatalogInstances that match those selectors.
func (c *serviceCatalogInstances) List(opts v1.ListOptions) (result *servicecatalog.ServiceCatalogInstanceList, err error) {
	result = &servicecatalog.ServiceCatalogInstanceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceCatalogInstances.
func (c *serviceCatalogInstances) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servicecataloginstances").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched serviceCatalogInstance.
func (c *serviceCatalogInstances) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ServiceCatalogInstance, err error) {
	result = &servicecatalog.ServiceCatalogInstance{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servicecataloginstances").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
