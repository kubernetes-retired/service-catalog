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

func TestPrettyNames(t *testing.T) {
	e := `ServiceInstance (K8S: "k8s" ExternalName: "extern")`
	g := Name(ServiceInstance, "k8s", "extern")
	if g != e {
		t.Fatalf("Unexpected value of PrettyName String; expected %v, got %v", e, g)
	}
}
