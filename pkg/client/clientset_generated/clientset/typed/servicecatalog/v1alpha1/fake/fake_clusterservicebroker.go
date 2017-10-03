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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterServiceBrokers implements ClusterServiceBrokerInterface
type FakeClusterServiceBrokers struct {
	Fake *FakeServicecatalogV1alpha1
}

var clusterservicebrokersResource = schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "clusterservicebrokers"}

var clusterservicebrokersKind = schema.GroupVersionKind{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Kind: "ClusterServiceBroker"}

func (c *FakeClusterServiceBrokers) Create(clusterServiceBroker *v1alpha1.ClusterServiceBroker) (result *v1alpha1.ClusterServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(clusterservicebrokersResource, clusterServiceBroker), &v1alpha1.ClusterServiceBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterServiceBroker), err
}

func (c *FakeClusterServiceBrokers) Update(clusterServiceBroker *v1alpha1.ClusterServiceBroker) (result *v1alpha1.ClusterServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(clusterservicebrokersResource, clusterServiceBroker), &v1alpha1.ClusterServiceBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterServiceBroker), err
}

func (c *FakeClusterServiceBrokers) UpdateStatus(clusterServiceBroker *v1alpha1.ClusterServiceBroker) (*v1alpha1.ClusterServiceBroker, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(clusterservicebrokersResource, "status", clusterServiceBroker), &v1alpha1.ClusterServiceBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterServiceBroker), err
}

func (c *FakeClusterServiceBrokers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(clusterservicebrokersResource, name), &v1alpha1.ClusterServiceBroker{})
	return err
}

func (c *FakeClusterServiceBrokers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(clusterservicebrokersResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ClusterServiceBrokerList{})
	return err
}

func (c *FakeClusterServiceBrokers) Get(name string, options v1.GetOptions) (result *v1alpha1.ClusterServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(clusterservicebrokersResource, name), &v1alpha1.ClusterServiceBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterServiceBroker), err
}

func (c *FakeClusterServiceBrokers) List(opts v1.ListOptions) (result *v1alpha1.ClusterServiceBrokerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(clusterservicebrokersResource, clusterservicebrokersKind, opts), &v1alpha1.ClusterServiceBrokerList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterServiceBrokerList{}
	for _, item := range obj.(*v1alpha1.ClusterServiceBrokerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterServiceBrokers.
func (c *FakeClusterServiceBrokers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(clusterservicebrokersResource, opts))
}

// Patch applies the patch and returns the patched clusterServiceBroker.
func (c *FakeClusterServiceBrokers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ClusterServiceBroker, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(clusterservicebrokersResource, name, data, subresources...), &v1alpha1.ClusterServiceBroker{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterServiceBroker), err
}
