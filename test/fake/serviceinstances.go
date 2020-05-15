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
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	watch "k8s.io/apimachinery/pkg/watch"

	v1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1beta1typed "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
)

// ServiceInstances is a wrapper around the generated fake ServiceInstances
// that clones the ServiceInstance objects being passed to UpdateStatus. This is a
// workaround until the generated fake clientset does its own copying.
type ServiceInstances struct {
	v1beta1typed.ServiceInstanceInterface
}

func (c *ServiceInstances) Create(ctx context.Context, serviceInstance *v1beta1.ServiceInstance, opts v1.CreateOptions) (*v1beta1.ServiceInstance, error) {
	return c.ServiceInstanceInterface.Create(ctx, serviceInstance, opts)
}

func (c *ServiceInstances) Update(ctx context.Context, serviceInstance *v1beta1.ServiceInstance, opts v1.UpdateOptions) (*v1beta1.ServiceInstance, error) {
	instanceCopy := serviceInstance.DeepCopy()
	updatedInstance, err := c.ServiceInstanceInterface.Update(ctx, instanceCopy, opts)
	if updatedInstance != nil {
		updatedInstance.ResourceVersion = rand.String(10)
	}
	return updatedInstance, err
}

func (c *ServiceInstances) UpdateStatus(ctx context.Context, serviceInstance *v1beta1.ServiceInstance, opts v1.UpdateOptions) (*v1beta1.ServiceInstance, error) {
	instanceCopy := serviceInstance.DeepCopy()
	updatedInstance, err := c.ServiceInstanceInterface.UpdateStatus(ctx, instanceCopy, opts)
	if updatedInstance != nil {
		updatedInstance.ResourceVersion = rand.String(10)
	}
	return updatedInstance, err
}

func (c *ServiceInstances) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.ServiceInstanceInterface.Delete(ctx, name, opts)
}

func (c *ServiceInstances) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return c.ServiceInstanceInterface.DeleteCollection(ctx, opts, listOpts)
}

func (c *ServiceInstances) Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.ServiceInstance, error) {
	return c.ServiceInstanceInterface.Get(ctx, name, opts)
}

func (c *ServiceInstances) List(ctx context.Context, opts v1.ListOptions) (*v1beta1.ServiceInstanceList, error) {
	return c.ServiceInstanceInterface.List(ctx, opts)
}

// Watch returns a watch.Interface that watches the requested serviceInstances.
func (c *ServiceInstances) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.ServiceInstanceInterface.Watch(ctx, opts)
}

// Patch applies the patch and returns the patched serviceInstance.
func (c *ServiceInstances) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ServiceInstance, err error) {
	return c.ServiceInstanceInterface.Patch(ctx, name, pt, data, opts, subresources...)
}
