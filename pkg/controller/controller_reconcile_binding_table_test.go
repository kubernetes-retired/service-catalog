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

package controller

import (
	scfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	clientgotesting "k8s.io/client-go/testing"
)

// bindingTestCase represents a single row of a table driven test for reconciling bindings
type reconcileBindingTestCase struct {
	// the name of the test
	name string
	// mutator for the service catalog fake client set - called before reconcileBinding is called
	modifyCatalogClient func(*scfake.Clientset) error
	// function to check the error returned by reconcileBinding
	checkReconcileErr func(error) error
	// the number of actions that should have occurred after reconcileBinding
	numActions int
	// function to check each update action. the int is the action number
	checkAction func(int, clientgotesting.Action) error
	// the number of events that should have occurred after reconcileBinding
	numEvents int
	// function to check each event. the int is the event number
	checkEvent func(int, string) error
}
