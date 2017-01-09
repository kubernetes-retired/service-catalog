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
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/watch"
)

// fakeDynamicResourceClient is a dynamicResourceClient implementation intended for unit tests.
// it contains a watch.Interface and error that it will always return in its Watch func.
//
// The watch.Interface will usually be a (k8s.io/client-go/1.5/pkg/watch).FakeWatcher, which will
// be driven (i.e. 'Add' func called) from unit tests
type fakeDynamicResourceClient struct {
	watchRet    watch.Interface
	watchRetErr error
}

func (f *fakeDynamicResourceClient) Watch(*v1.ListOptions) (watch.Interface, error) {
	return f.watchRet, f.watchRetErr
}
