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
	v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterIDs implements ClusterIDInterface
type FakeClusterIDs struct {
	Fake *FakeServicecatalogV1beta1
}

var clusteridsResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1beta1", Resource: "clusterids"}

var clusteridsKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1beta1", Kind: "ClusterID"}

// Get takes name of the clusterID, and returns the corresponding clusterID object, and an error if there is any.
func (c *FakeClusterIDs) Get(name string, options v1.GetOptions) (result *v1beta1.ClusterID, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(clusteridsResource, name), &v1beta1.ClusterID{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ClusterID), err
}

// List takes label and field selectors, and returns the list of ClusterIDs that match those selectors.
func (c *FakeClusterIDs) List(opts v1.ListOptions) (result *v1beta1.ClusterIDList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(clusteridsResource, clusteridsKind, opts), &v1beta1.ClusterIDList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ClusterIDList{}
	for _, item := range obj.(*v1beta1.ClusterIDList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterIDs.
func (c *FakeClusterIDs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(clusteridsResource, opts))
}

// Create takes the representation of a clusterID and creates it.  Returns the server's representation of the clusterID, and an error, if there is any.
func (c *FakeClusterIDs) Create(clusterID *v1beta1.ClusterID) (result *v1beta1.ClusterID, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(clusteridsResource, clusterID), &v1beta1.ClusterID{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ClusterID), err
}

// Update takes the representation of a clusterID and updates it. Returns the server's representation of the clusterID, and an error, if there is any.
func (c *FakeClusterIDs) Update(clusterID *v1beta1.ClusterID) (result *v1beta1.ClusterID, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clusteridsResource, clusterID), &v1beta1.ClusterID{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ClusterID), err
}

// Delete takes name of the clusterID and deletes it. Returns an error if one occurs.
func (c *FakeClusterIDs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(clusteridsResource, name), &v1beta1.ClusterID{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterIDs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(clusteridsResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ClusterIDList{})
	return err
}

// Patch applies the patch and returns the patched clusterID.
func (c *FakeClusterIDs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.ClusterID, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(clusteridsResource, name, data, subresources...), &v1beta1.ClusterID{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ClusterID), err
}
