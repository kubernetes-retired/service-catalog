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

package meta

import (
	"testing"
	"time"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAccessor(t *testing.T) {
	if GetAccessor() != accessor {
		t.Fatalf("GetAccessor didn't return the pre-initialized accessor")
	}
}

func TestGetNamespace(t *testing.T) {
	const namespace = "testns"
	obj := &sc.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
	}
	ns, err := GetNamespace(obj)
	if err != nil {
		t.Fatalf("error getting namespace (%s)", err)
	}
	if ns != namespace {
		t.Fatalf("actual namespace (%s) wasn't expected (%s)", ns, namespace)
	}
}

func TestDeletionTimestampExists(t *testing.T) {
	obj := &sc.Instance{
		ObjectMeta: metav1.ObjectMeta{},
	}
	exists, err := DeletionTimestampExists(obj)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("deletion timestamp reported as exists when it didn't")
	}
	tme := metav1.NewTime(time.Now())
	obj.DeletionTimestamp = &tme
	exists, err = DeletionTimestampExists(obj)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("deletion timestamp reported as missing when it isn't")
	}
}

func TestGetDeletionTimestamp(t *testing.T) {
	// TODO: implement
	t.Skip("TODO")
}

func TestSetDeletionTimestamp(t *testing.T) {
	// TODO: implement
	t.Skip("TODO")
}
