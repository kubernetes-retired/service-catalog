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

// ServiceCatalogServiceClassesGetter has a method to return a ServiceCatalogServiceClassInterface.
// A group's client should implement this interface.
type ServiceCatalogServiceClassesGetter interface {
	ServiceCatalogServiceClasses() ServiceCatalogServiceClassInterface
}

// ServiceCatalogServiceClassInterface has methods to work with ServiceCatalogServiceClass resources.
type ServiceCatalogServiceClassInterface interface {
	Create(*v1alpha1.ServiceCatalogServiceClass) (*v1alpha1.ServiceCatalogServiceClass, error)
	Update(*v1alpha1.ServiceCatalogServiceClass) (*v1alpha1.ServiceCatalogServiceClass, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ServiceCatalogServiceClass, error)
	List(opts v1.ListOptions) (*v1alpha1.ServiceCatalogServiceClassList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceCatalogServiceClass, err error)
	ServiceCatalogServiceClassExpansion
}

// serviceCatalogServiceClasses implements ServiceCatalogServiceClassInterface
type serviceCatalogServiceClasses struct {
	client rest.Interface
}

// newServiceCatalogServiceClasses returns a ServiceCatalogServiceClasses
func newServiceCatalogServiceClasses(c *ServicecatalogV1alpha1Client) *serviceCatalogServiceClasses {
	return &serviceCatalogServiceClasses{
		client: c.RESTClient(),
	}
}

// Create takes the representation of a serviceCatalogServiceClass and creates it.  Returns the server's representation of the serviceCatalogServiceClass, and an error, if there is any.
func (c *serviceCatalogServiceClasses) Create(serviceCatalogServiceClass *v1alpha1.ServiceCatalogServiceClass) (result *v1alpha1.ServiceCatalogServiceClass, err error) {
	result = &v1alpha1.ServiceCatalogServiceClass{}
	err = c.client.Post().
		Resource("servicecatalogserviceclasses").
		Body(serviceCatalogServiceClass).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceCatalogServiceClass and updates it. Returns the server's representation of the serviceCatalogServiceClass, and an error, if there is any.
func (c *serviceCatalogServiceClasses) Update(serviceCatalogServiceClass *v1alpha1.ServiceCatalogServiceClass) (result *v1alpha1.ServiceCatalogServiceClass, err error) {
	result = &v1alpha1.ServiceCatalogServiceClass{}
	err = c.client.Put().
		Resource("servicecatalogserviceclasses").
		Name(serviceCatalogServiceClass.Name).
		Body(serviceCatalogServiceClass).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceCatalogServiceClass and deletes it. Returns an error if one occurs.
func (c *serviceCatalogServiceClasses) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("servicecatalogserviceclasses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceCatalogServiceClasses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("servicecatalogserviceclasses").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the serviceCatalogServiceClass, and returns the corresponding serviceCatalogServiceClass object, and an error if there is any.
func (c *serviceCatalogServiceClasses) Get(name string, options v1.GetOptions) (result *v1alpha1.ServiceCatalogServiceClass, err error) {
	result = &v1alpha1.ServiceCatalogServiceClass{}
	err = c.client.Get().
		Resource("servicecatalogserviceclasses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceCatalogServiceClasses that match those selectors.
func (c *serviceCatalogServiceClasses) List(opts v1.ListOptions) (result *v1alpha1.ServiceCatalogServiceClassList, err error) {
	result = &v1alpha1.ServiceCatalogServiceClassList{}
	err = c.client.Get().
		Resource("servicecatalogserviceclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceCatalogServiceClasses.
func (c *serviceCatalogServiceClasses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("servicecatalogserviceclasses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched serviceCatalogServiceClass.
func (c *serviceCatalogServiceClasses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ServiceCatalogServiceClass, err error) {
	result = &v1alpha1.ServiceCatalogServiceClass{}
	err = c.client.Patch(pt).
		Resource("servicecatalogserviceclasses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
