/*
Copyright 2018 The Kubernetes Authors.

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
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgotesting "k8s.io/client-go/testing"
)

// TestGetClusterIDDefaulting ensures the ID defaulting works to
// create an ID on access if one is not set.
func TestGetClusterIDDefaulting(t *testing.T) {
	_, _, _, tc, _ := newTestController(t, noFakeActions())
	tc.setClusterID("")
	if tc.getClusterID() == "" {
		t.Fatalf("cluster id should have been generated and filled in upon request")
	}
	t.Log(tc.getClusterID())
}

// TestGetClusterIDConcurrently make sure that there is a consistent
// state if two calls are made concurrently. This test should be run
// many times on a processor capable of running multiple goroutines.
func TestGetClusterIDConcurrently(t *testing.T) {
	_, _, _, tc, _ := newTestController(t, noFakeActions())
	tc.setClusterID("")

	var a, b string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { a = tc.getClusterID(); wg.Done() }()
	go func() { b = tc.getClusterID(); wg.Done() }()
	wg.Wait()
	if a != b {
		t.Fatal("a and b should match", a, b)
	}
	if tc.getClusterID() == "" {
		t.Fatalf("cluster id should have been generated and filled in upon request")
	}
	t.Log(tc.getClusterID())

}

// TestGetClusterIDRoundTrip soley tests the controllers ID accessor
// functions to ensure we get back out what we put in.
func TestGetClusterIDRoundTrip(t *testing.T) {
	_, _, _, tc, _ := newTestController(t, noFakeActions())
	tc.setClusterID("")
	tc.setClusterID(testClusterID)
	if tc.getClusterID() != testClusterID {
		t.Fatalf("should have got the same string out that we put in")
	}
}

// TestMonitorConfigMapNoConfigmap ensures that if we don't have a
// configmap ID, but the ID in the controller is set, that the ID in
// the controller does not change. The configmap is also created with
// the existing value filled in.
func TestMonitorConfigMapNoConfigmap(t *testing.T) {
	kc, _, _, tc, _ := newTestController(t, noFakeActions())
	kc.AddReactor("get", "configmaps", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		m := make(map[string]string)
		m["id"] = testClusterID
		return true, nil, errors.NewNotFound(schema.GroupResource{"core", "configmap"}, DefaultClusterIDConfigMapName)
	})
	tc.setClusterID(testClusterID)
	tc.monitorConfigMap()
	if tc.getClusterID() != testClusterID {
		t.Fatalf("should have got the same string out that we put in")
	}
	if kc.Actions()[1].Matches("create", "configmaps") {
		createdCM := kc.Actions()[1].(clientgotesting.CreateAction).GetObject().(*corev1.ConfigMap)
		if id, ok := createdCM.Data["id"]; !(ok && id == testClusterID) {
			t.Fatalf("new configmap should have id as existing testClusterID. Had id %q", id)
		}
	} else {
		t.Fatalf("should have created a new configmap")
	}
}

// TestMonitorConfigMapNoConfigmapNoExistingClusterID checks that if a
// configmap does not exist, it is created and filled in with a
// generated ID
func TestMonitorConfigMapNoConfigmapNoExistingClusterID(t *testing.T) {
	kc, _, _, tc, _ := newTestController(t, noFakeActions())
	kc.AddReactor("get", "configmaps", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		m := make(map[string]string)
		m["id"] = testClusterID
		return true, nil, errors.NewNotFound(schema.GroupResource{"core", "configmap"}, DefaultClusterIDConfigMapName)
	})
	tc.setClusterID("")
	tc.monitorConfigMap()
	if tc.getClusterID() == "" {
		t.Fatalf("cluster id should have been generated and filled in upon request")
	}
	if kc.Actions()[1].Matches("create", "configmaps") {
		createdCM := kc.Actions()[1].(clientgotesting.CreateAction).GetObject().(*corev1.ConfigMap)
		if id, ok := createdCM.Data["id"]; !(ok && id != "") {
			t.Fatalf("new configmap should have a non-blank id")
		}
	} else {
		t.Fatalf("should have created a new configmap")
	}

}

// TestMonitorConfigMapConfigmapOverride checks that the ID is set to
// the value contained in the ConfigMap, even if an existing ID is
// present in the controller.
func TestMonitorConfigMapConfigmapOverride(t *testing.T) {
	kc, _, _, tc, _ := newTestController(t, noFakeActions())
	kc.AddReactor("get", "configmaps", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		m := make(map[string]string)
		m["id"] = testClusterID // override existing ID with standard test ID
		return true, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: DefaultClusterIDConfigMapName,
			},
			Data: m,
		}, nil
	})
	tc.setClusterID("non-cluster-id") // existing id to be overridden
	tc.monitorConfigMap()
	if tc.getClusterID() != testClusterID {
		t.Fatalf("should have got the override id from the configmap")
	}
}

// TestMonitorConfigMapConfigmapWithNoData checks that if a configmap
// with no data to have an ID is returned, that we update the
// configmap to have an ID field with the currently existing ID from
// the controller
func TestMonitorConfigMapConfigmapWithNoData(t *testing.T) {
	kc, _, _, tc, _ := newTestController(t, noFakeActions())
	blankcm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultClusterIDConfigMapName,
		},
	}
	kc.AddReactor("get", "configmaps", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, blankcm, nil
	})
	tc.setClusterID(testClusterID)
	tc.monitorConfigMap()
	if tc.getClusterID() != testClusterID {
		t.Fatalf("should have got the set cluster id")
	}
	if expectedCMget := kc.Actions()[0]; expectedCMget.GetVerb() != "get" {
		t.Fatalf("get configmap is first")
	}
	if expectedCMupdate := kc.Actions()[1]; expectedCMupdate.GetVerb() == "update" {
		updatedCM := expectedCMupdate.(clientgotesting.UpdateAction).GetObject().(*corev1.ConfigMap)
		if id := updatedCM.Data["id"]; id != testClusterID {
			t.Fatalf("configmap should have been updated with the existing clusterid, was %q, expected %q", id, testClusterID)
		}
	} else {
		t.Fatalf("configmap should have been updated with the existing clusterid")
	}
}

// TestMonitorConfigMapConfigmapWithNoData checks that if a configmap
// with no data to have an ID is returned, that we update the
// configmap to have an ID field with the currently existing ID from
// the controller
func TestMonitorConfigMapConfigmapWithOtherData(t *testing.T) {
	kc, _, _, tc, _ := newTestController(t, noFakeActions())
	kc.AddReactor("get", "configmaps", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		m := make(map[string]string)
		m["notid"] = "other-non-id-stuff-that-needs-to-be-perserved"
		return true, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: DefaultClusterIDConfigMapName,
			},
			Data: m,
		}, nil
	})
	tc.setClusterID(testClusterID)
	tc.monitorConfigMap()
	if tc.getClusterID() != testClusterID {
		t.Fatalf("should have got the set cluster id")
	}
	if expectedCMget := kc.Actions()[0]; expectedCMget.GetVerb() != "get" {
		t.Fatalf("get configmap is first")
	}
	if expectedCMupdate := kc.Actions()[1]; expectedCMupdate.GetVerb() == "update" {
		updatedCM := expectedCMupdate.(clientgotesting.UpdateAction).GetObject().(*corev1.ConfigMap)
		if notid := updatedCM.Data["notid"]; notid == "" {
			t.Fatalf("configmap should have another key")
		}
		if id := updatedCM.Data["id"]; id != testClusterID {
			t.Fatalf("configmap should have been updated with the existing clusterid")
		}
	} else {
		t.Fatalf("configmap should have been updated with the existing clusterid")
	}
}
