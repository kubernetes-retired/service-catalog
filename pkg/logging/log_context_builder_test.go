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

package logging

import (
	"testing"
)

func TestLogContextBuilderKind(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetKind(ServiceInstance)

	e := "ServiceInstance"
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderNamespace(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetNamespace("Namespace")

	e := `"Namespace"`
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderName(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetName("Name")

	e := `"Name"`
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderKindAndNamespace(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetKind(ServiceInstance).SetNamespace("Namespace")

	e := `ServiceInstance "Namespace"`
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderKindAndName(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetKind(ServiceInstance).SetName("Name")

	e := `ServiceInstance "Name"`
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderKindNamespaceName(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetKind(ServiceInstance).SetNamespace("Namespace").SetName("Name")

	e := `ServiceInstance "Namespace/Name"`
	g := lbc.String()
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderMsg(t *testing.T) {
	lbc := LogContextBuilder{}

	e := `Msg`
	g := lbc.Message("Msg")
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestLogContextBuilderContextAndMsg(t *testing.T) {
	lbc := LogContextBuilder{}

	lbc.SetKind(ServiceInstance).SetNamespace("Namespace").SetName("Name")

	e := `ServiceInstance "Namespace/Name": Msg`
	g := lbc.Message("Msg")
	if g != e {
		t.Fatalf("Unexpected value of LogContextBuilder String; expected %v, got %v", e, g)
	}
}
