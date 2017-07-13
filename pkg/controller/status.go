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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AssertInstanceReadyTrue ensures that obj is an instance and has a ready condition of True
// on it
func AssertInstanceReadyTrue(t *testing.T, obj runtime.Object) {
	AssertInstanceReadyCondition(t, obj, v1alpha1.ConditionTrue)
}

// this function is here to maintain compatibility of existing controller test code
func assertInstanceReadyTrue(t *testing.T, obj runtime.Object) {
	AssertInstanceReadyTrue(t, obj)
}

// AssertInstanceReadyFalse ensures that obj is an instance, has a ready condition of False
// on it, and the reason for that ready condition is reason
func AssertInstanceReadyFalse(t *testing.T, obj runtime.Object, reason ...string) {
	AssertInstanceReadyCondition(t, obj, v1alpha1.ConditionFalse, reason...)
}

// this function is here to maintain compatibility of existing controller code
func assertInstanceReadyFalse(t *testing.T, obj runtime.Object, reason ...string) {
	AssertInstanceReadyFalse(t, obj, reason...)
}

// AssertAsyncOpInProgressFalse fails if obj does not have a condition on it that has the type
// of async op in progress and the status of it as false
func AssertAsyncOpInProgressFalse(t *testing.T, obj runtime.Object) {
	instance, ok := obj.(*v1alpha1.Instance)
	if !ok {
		t.Fatalf("Couldn't convert object %+v into a *v1alpha1.Instance", obj)
	}
	if instance.Status.AsyncOpInProgress {
		t.Fatalf("expected AsyncOpInProgress to be false but was %v", instance.Status.AsyncOpInProgress)
	}
}

// AssertAsyncOpInProgressTrue fails if obj does not have a condition on it that has the type
// of async op in progress and the status of it as true
func AssertAsyncOpInProgressTrue(t *testing.T, obj runtime.Object) {
	instance, ok := obj.(*v1alpha1.Instance)
	if !ok {
		t.Fatalf("Couldn't convert object %+v into a *v1alpha1.Instance", obj)
	}
	if !instance.Status.AsyncOpInProgress {
		t.Fatalf("expected AsyncOpInProgress to be true but was %v", instance.Status.AsyncOpInProgress)
	}
}

// AssertInstanceReadyCondition ensures that obj is an instance and that it has a ready condition
// with the given status and reason on it
func AssertInstanceReadyCondition(t *testing.T, obj runtime.Object, status v1alpha1.ConditionStatus, reason ...string) {
	instance, ok := obj.(*v1alpha1.Instance)
	if !ok {
		Fatalf(t, "Couldn't convert object %+v into a *v1alpha1.Instance", obj)
	}

	for _, condition := range instance.Status.Conditions {
		if condition.Type == v1alpha1.InstanceConditionReady && condition.Status != status {
			Fatalf(t, "ready condition had unexpected status; expected %v, got %v", status, condition.Status)
		}
		if len(reason) == 1 && condition.Reason != reason[0] {
			Fatalf(t, "unexpected reason; expected %v, got %v", reason[0], condition.Reason)
		}
	}
}
