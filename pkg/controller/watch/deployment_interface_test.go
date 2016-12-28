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
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	extv1beta1 "k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/watch"
)

// v1beta1.DeploymentInterface compatible type for use in unit tests
type fakeDeploymentInterface struct {
	listRet     *v1beta1.DeploymentList
	listRetErr  error
	watchRet    watch.Interface
	watchRetErr error
}

func (f *fakeDeploymentInterface) Create(*v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentInterface) Update(*v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentInterface) UpdateStatus(*v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentInterface) Delete(string, *api.DeleteOptions) error {
	return nil
}
func (f *fakeDeploymentInterface) DeleteCollection(*api.DeleteOptions, api.ListOptions) error {
	return nil
}
func (f *fakeDeploymentInterface) Get(string) (*v1beta1.Deployment, error) {
	return nil, nil
}
func (f *fakeDeploymentInterface) List(api.ListOptions) (*extv1beta1.DeploymentList, error) {
	return f.listRet, f.listRetErr
}
func (f *fakeDeploymentInterface) Watch(api.ListOptions) (watch.Interface, error) {
	return f.watchRet, f.watchRetErr
}
func (f *fakeDeploymentInterface) Patch(string, api.PatchType, []byte, ...string) (*v1beta1.Deployment, error) {
	return nil, nil
}

func (f *fakeDeploymentInterface) Rollback(*v1beta1.DeploymentRollback) error {
	return nil
}
