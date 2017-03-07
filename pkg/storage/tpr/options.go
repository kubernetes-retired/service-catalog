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

package tpr

import (
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
)

// Options is the set of options to create a new TPR storage interface
type Options struct {
	HasNamespace     bool
	RESTOptions      generic.RESTOptions
	DefaultNamespace string
	Client           clientset.Interface
	SingularKind     Kind
	NewSingularFunc  func(string, string) runtime.Object
	ListKind         Kind
	NewListFunc      func() runtime.Object
	CheckObjectFunc  func(runtime.Object) error
	DestroyFunc      func()
	Keyer            Keyer
}
