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

package v1alpha1

import (
	v1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	scheme "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceCatalogBrokersGetter has a method to return a ServiceCatalogBrokerInterface.
// A group's client should implement this interface.
type ServiceCatalogBrokersGetter interface {
	ServiceCatalogBrokers() ServiceCatalogBrokerInterface
}

// ServiceCatalogBrokerInterface has methods to work with ServiceCatalogBroker resources.
type ServiceCatalogBrokerInterface interface {
	Create(*v1alpha1.ServiceCatalogBroker) (*v1alpha1.ServiceCatalogBroker, error)
	Update(*v1alpha1.ServiceCatalogBroker) (*v1alpha1.ServiceCatalogBroker, error)
	UpdateStatus(*v1alpha1.ServiceCatalogBroker) (*v1alpha1.ServiceCatalogBroker, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ServiceCatalogBroker, error)
	List(opts v1.ListOptions) (*v1alpha1.ServiceCatalogBrokerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceCatalogBroker, err error)
	ServiceCatalogBrokerExpansion
}

// serviceCatalogBrokers implements ServiceCatalogBrokerInterface
type serviceCatalogBrokers struct {
	client rest.Interface
}

// newServiceCatalogBrokers returns a ServiceCatalogBrokers
func newServiceCatalogBrokers(c *ServicecatalogV1alpha1Client) *serviceCatalogBrokers {
	return &serviceCatalogBrokers{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a serviceCatalogBroker and creates it.  Returns the server's representation of the serviceCatalogBroker, and an error, if there is any.
func (c *serviceCatalogBrokers) Create(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (result *v1alpha1.ServiceCatalogBroker, err error) {
	result = &v1alpha1.ServiceCatalogBroker{}
	err = c.client.Post().
		Resource("servicecatalogbrokers").
		Body(serviceCatalogBroker).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceCatalogBroker and updates it. Returns the server's representation of the serviceCatalogBroker, and an error, if there is any.
func (c *serviceCatalogBrokers) Update(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (result *v1alpha1.ServiceCatalogBroker, err error) {
	result = &v1alpha1.ServiceCatalogBroker{}
	err = c.client.Put().
		Resource("servicecatalogbrokers").
		Name(serviceCatalogBroker.Name).
		Body(serviceCatalogBroker).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *serviceCatalogBrokers) UpdateStatus(serviceCatalogBroker *v1alpha1.ServiceCatalogBroker) (result *v1alpha1.ServiceCatalogBroker, err error) {
	result = &v1alpha1.ServiceCatalogBroker{}
	err = c.client.Put().
		Resource("servicecatalogbrokers").
		Name(serviceCatalogBroker.Name).
		SubResource("status").
		Body(serviceCatalogBroker).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceCatalogBroker and deletes it. Returns an error if one occurs.
func (c *serviceCatalogBrokers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("servicecatalogbrokers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceCatalogBrokers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("servicecatalogbrokers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the serviceCatalogBroker, and returns the corresponding serviceCatalogBroker object, and an error if there is any.
func (c *serviceCatalogBrokers) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceCatalogBroker, err error) {
	result = &v1alpha1.ServiceCatalogBroker{}
	err = c.client.Get().
		Resource("servicecatalogbrokers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceCatalogBrokers that match those selectors.
func (c *serviceCatalogBrokers) List(opts v1.ListOptions) (result *v1alpha1.ServiceCatalogBrokerList, err error) {
	result = &v1alpha1.ServiceCatalogBrokerList{}
	err = c.client.Get().
		Resource("servicecatalogbrokers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceCatalogBrokers.
func (c *serviceCatalogBrokers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("servicecatalogbrokers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched serviceCatalogBroker.
func (c *serviceCatalogBrokers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceCatalogBroker, err error) {
	result = &v1alpha1.ServiceCatalogBroker{}
	err = c.client.Patch(pt).
		Resource("servicecatalogbrokers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
