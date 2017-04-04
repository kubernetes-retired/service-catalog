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
	"testing"

	"k8s.io/kubernetes/pkg/api"
)

const (
	defaultCtxNS = "defaultTestNS"
	ctxNS        = "testNS"
)

func TestKeyRoot(t *testing.T) {
	ctx := api.NewContext()
	ctx = api.WithNamespace(ctx, ctxNS)
	keyer := Keyer{DefaultNamespace: defaultCtxNS}
	root := keyer.KeyRoot(ctx)
	if root != ctxNS {
		t.Fatalf("key root '%s' wasn't expected '%s'", root, ctxNS)
	}
	ctx = api.NewContext()
	root = keyer.NewRoot(ctx)
	if root != keyer.DefaultNamespace {
		t.Fatalf("key root '%s' wasn't expected '%s'", root, keyer.DefaultNamespace)
	}
}

func TestKey(t *testing.T) {
	t.Skip("TODO")
}

func TestNamespaceAndNameFromKey(t *testing.T) {
	t.Skip("TODO")
}
