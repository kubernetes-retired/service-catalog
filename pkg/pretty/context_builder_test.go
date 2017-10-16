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

package pretty

import (
	"testing"
)

func TestPrettyContextBuilderKind(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetKind(ServiceInstance)

	e := "ServiceInstance"
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderNamespace(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetNamespace("Namespace")

	e := `"Namespace"`
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderName(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetName("Name")

	e := `"Name"`
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderKindAndNamespace(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetKind(ServiceInstance).SetNamespace("Namespace")

	e := `ServiceInstance "Namespace"`
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderKindAndName(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetKind(ServiceInstance).SetName("Name")

	e := `ServiceInstance "Name"`
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderKindNamespaceName(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetKind(ServiceInstance).SetNamespace("Namespace").SetName("Name")

	e := `ServiceInstance "Namespace/Name"`
	g := pcb.String()
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderMsg(t *testing.T) {
	pcb := ContextBuilder{}

	e := `Msg`
	g := pcb.Message("Msg")
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}

func TestPrettyContextBuilderContextAndMsg(t *testing.T) {
	pcb := ContextBuilder{}

	pcb.SetKind(ServiceInstance).SetNamespace("Namespace").SetName("Name")

	e := `ServiceInstance "Namespace/Name": Msg`
	g := pcb.Message("Msg")
	if g != e {
		t.Fatalf("Unexpected value of ContextBuilder String; expected %v, got %v", e, g)
	}
}
