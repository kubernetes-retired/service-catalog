/*
Copyright 2016 The Kubernetes Authors.

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

package watch

import (
	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

// dynamicResourceClient is an interface that allows callers to watch arbitrary resource types
// in Kubernetes. It's passed to functions that expect to call 'Watch()' on the returned
// watch.Interface and convert the returned event objects to such an arbitrary resource type.
// Generally such functions will take this interface so they can be more easily unit tested.
//
// this interface has two important implementations:
//
// - realDynamicResourceClient - a Kubernetes implementation that's based on
//	 a *(k8s.io/client-go/1.5/dynamic).ResourceClient
// - fakeDynamicResourceClient - a manually-driven implementation that should be used in unit
//   tests
type dynamicResourceClient interface {
	Watch(*v1.ListOptions) (watch.Interface, error)
}

// realDynamicResourceClient is a dynamicResourceClient implementation that uses a
// *(k8s.io/client-go/1.5/dynamic).ResourceClient to implement its Watch func
type realDynamicResourceClient struct {
	rc *dynamic.ResourceClient
}

func newRealDynamicResourceClient(rc *dynamic.ResourceClient) *realDynamicResourceClient {
	return &realDynamicResourceClient{rc: rc}
}

func (r *realDynamicResourceClient) Watch(opts *v1.ListOptions) (watch.Interface, error) {
	return r.rc.Watch(opts)
}
