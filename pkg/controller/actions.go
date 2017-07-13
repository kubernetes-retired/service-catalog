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
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

// TestNumberOfActions checks to ensure that actions has the given (in number) number of actions
// in it. Calls f with the failure if it doesn't have that number of actions
func TestNumberOfActions(t *testing.T, name string, f FailfFunc, actions []clientgotesting.Action, number int) bool {
	logContext := ""
	if len(name) > 0 {
		logContext = name + ": "
	}

	if e, a := number, len(actions); e != a {
		t.Logf("%+v\n", actions)
		f(t, "%vUnexpected number of actions: expected %v, got %v;\nactions: %+v", logContext, e, a, actions)
		return false
	}

	return true
}

// AssertNumberOfActions fails if actions was not of length number
func AssertNumberOfActions(t *testing.T, actions []clientgotesting.Action, number int) {
	TestNumberOfActions(t, "" /* name */, Fatalf, actions, number)
}

// AssertUpdateStatus ensures that the status of action is an update on obj
func AssertUpdateStatus(t *testing.T, action clientgotesting.Action, obj interface{}) runtime.Object {
	return AssertActionFor(t, action, "update", "status", obj)
}

// AssertActionFor tests for that verb and subresource are set on action for the given object
func AssertActionFor(t *testing.T, action clientgotesting.Action, verb, subresource string, obj interface{}) runtime.Object {
	r, _ := TestActionFor(t, "" /* name */, Fatalf, action, verb, subresource, obj)
	return r
}

// TestActionFor ensures that the given action, verb and subresource are set on obj
func TestActionFor(
	t *testing.T,
	name string,
	f FailfFunc,
	action clientgotesting.Action,
	verb,
	subresource string,
	obj interface{},
) (runtime.Object, bool) {
	logContext := ""
	if len(name) > 0 {
		logContext = name + ": "
	}

	if e, a := verb, action.GetVerb(); e != a {
		f(t, "%vUnexpected verb: expected %v, got %v", logContext, e, a)
		return nil, false
	}

	var resource string

	switch obj.(type) {
	case *v1alpha1.Broker:
		resource = "brokers"
	case *v1alpha1.ServiceClass:
		resource = "serviceclasses"
	case *v1alpha1.Instance:
		resource = "instances"
	case *v1alpha1.Binding:
		resource = "bindings"
	}

	if e, a := resource, action.GetResource().Resource; e != a {
		f(t, "%vUnexpected resource; expected %v, got %v", logContext, e, a)
		return nil, false
	}

	if e, a := subresource, action.GetSubresource(); e != a {
		f(t, "%vUnexpected subresource; expected %v, got %v", logContext, e, a)
		return nil, false
	}

	rtObject, ok := obj.(runtime.Object)
	if !ok {
		f(t, "%vObject %+v was not a runtime.Object", logContext, obj)
		return nil, false
	}

	paramAccessor, err := metav1.ObjectMetaFor(rtObject)
	if err != nil {
		f(t, "%vError creating ObjectMetaAccessor for param object %+v: %v", logContext, rtObject, err)
		return nil, false
	}

	var (
		objectMeta   metav1.Object
		fakeRtObject runtime.Object
	)

	switch verb {
	case "get":
		getAction, ok := action.(clientgotesting.GetAction)
		if !ok {
			f(t, "%vUnexpected type; failed to convert action %+v to DeleteAction", logContext, action)
			return nil, false
		}

		if e, a := paramAccessor.GetName(), getAction.GetName(); e != a {
			f(t, "%vUnexpected name: expected %v, got %v", logContext, e, a)
			return nil, false
		}

		return nil, true
	case "delete":
		deleteAction, ok := action.(clientgotesting.DeleteAction)
		if !ok {
			f(t, "%vUnexpected type; failed to convert action %+v to DeleteAction", logContext, action)
			return nil, false
		}

		if e, a := paramAccessor.GetName(), deleteAction.GetName(); e != a {
			f(t, "%vUnexpected name: expected %v, got %v", logContext, e, a)
			return nil, false
		}

		return nil, true
	case "create":
		createAction, ok := action.(clientgotesting.CreateAction)
		if !ok {
			f(t, "%vUnexpected type; failed to convert action %+v to CreateAction", logContext, action)
			return nil, false
		}

		fakeRtObject = createAction.GetObject()
		objectMeta, err = metav1.ObjectMetaFor(fakeRtObject)
		if err != nil {
			f(t, "%vError creating ObjectMetaAccessor for %+v", logContext, fakeRtObject)
			return nil, false
		}
	case "update":
		updateAction, ok := action.(clientgotesting.UpdateAction)
		if !ok {
			f(t, "%vUnexpected type; failed to convert action %+v to UpdateAction", logContext, action)
			return nil, false
		}

		fakeRtObject = updateAction.GetObject()
		objectMeta, err = metav1.ObjectMetaFor(fakeRtObject)
		if err != nil {
			f(t, "%vError creating ObjectMetaAccessor for %+v", logContext, fakeRtObject)
			return nil, false
		}
	}

	if e, a := paramAccessor.GetName(), objectMeta.GetName(); e != a {
		f(t, "%vUnexpected name: expected %v, got %v", logContext, e, a)
		return nil, false
	}

	fakeValue := reflect.ValueOf(fakeRtObject)
	paramValue := reflect.ValueOf(obj)

	if e, a := paramValue.Type(), fakeValue.Type(); e != a {
		f(t, "%vUnexpected type of object passed to fake client; expected %v, got %v", logContext, e, a)
		return nil, false
	}

	return fakeRtObject, true
}
