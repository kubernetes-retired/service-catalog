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

// ClusterIDsGetter has a method to return a ClusterIDInterface.
// A group's client should implement this interface.
type ClusterIDsGetter interface {
	ClusterIDs() ClusterIDInterface
}

// ClusterIDInterface has methods to work with ClusterID resources.
type ClusterIDInterface interface {
	Create(*servicecatalog.ClusterID) (*servicecatalog.ClusterID, error)
	Update(*servicecatalog.ClusterID) (*servicecatalog.ClusterID, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*servicecatalog.ClusterID, error)
	List(opts v1.ListOptions) (*servicecatalog.ClusterIDList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ClusterID, err error)
	ClusterIDExpansion
}

// clusterIDs implements ClusterIDInterface
type clusterIDs struct {
	client rest.Interface
}

// newClusterIDs returns a ClusterIDs
func newClusterIDs(c *ServicecatalogClient) *clusterIDs {
	return &clusterIDs{
		client: c.RESTClient(),
	}
}

// Get takes name of the clusterID, and returns the corresponding clusterID object, and an error if there is any.
func (c *clusterIDs) Get(name string, options v1.GetOptions) (result *servicecatalog.ClusterID, err error) {
	result = &servicecatalog.ClusterID{}
	err = c.client.Get().
		Resource("clusterids").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ClusterIDs that match those selectors.
func (c *clusterIDs) List(opts v1.ListOptions) (result *servicecatalog.ClusterIDList, err error) {
	result = &servicecatalog.ClusterIDList{}
	err = c.client.Get().
		Resource("clusterids").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested clusterIDs.
func (c *clusterIDs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("clusterids").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a clusterID and creates it.  Returns the server's representation of the clusterID, and an error, if there is any.
func (c *clusterIDs) Create(clusterID *servicecatalog.ClusterID) (result *servicecatalog.ClusterID, err error) {
	result = &servicecatalog.ClusterID{}
	err = c.client.Post().
		Resource("clusterids").
		Body(clusterID).
		Do().
		Into(result)
	return
}

// Update takes the representation of a clusterID and updates it. Returns the server's representation of the clusterID, and an error, if there is any.
func (c *clusterIDs) Update(clusterID *servicecatalog.ClusterID) (result *servicecatalog.ClusterID, err error) {
	result = &servicecatalog.ClusterID{}
	err = c.client.Put().
		Resource("clusterids").
		Name(clusterID.Name).
		Body(clusterID).
		Do().
		Into(result)
	return
}

// Delete takes name of the clusterID and deletes it. Returns an error if one occurs.
func (c *clusterIDs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("clusterids").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *clusterIDs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("clusterids").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched clusterID.
func (c *clusterIDs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *servicecatalog.ClusterID, err error) {
	result = &servicecatalog.ClusterID{}
	err = c.client.Patch(pt).
		Resource("clusterids").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
