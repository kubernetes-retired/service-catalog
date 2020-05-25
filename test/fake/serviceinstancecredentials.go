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
	v1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1beta1typed "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBindings is a wrapper around the generated fake
// ServiceBindings that clones the ServiceBinding objects
// being passed to UpdateStatus. This is a workaround until the generated fake
// clientset does its own copying.
type ServiceBindings struct {
	v1beta1typed.ServiceBindingInterface
}

func (c *ServiceBindings) UpdateStatus(ctx context.Context, serviceBinding *v1beta1.ServiceBinding, opts v1.UpdateOptions) (*v1beta1.ServiceBinding, error) {
	instanceCopy := serviceBinding.DeepCopy()
	_, err := c.ServiceBindingInterface.UpdateStatus(ctx, instanceCopy, opts)
	return serviceBinding, err
}
