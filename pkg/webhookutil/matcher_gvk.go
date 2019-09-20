/*
Copyright 2019 The Kubernetes Authors.

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

package webhookutil

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func equalGVK(a metav1.GroupVersionKind, b schema.GroupVersionKind) bool {
	return a.Kind == b.Kind && a.Version == b.Version && a.Group == b.Group
}

// MatchKinds returns error if given obj GVK is not equal to the reqKind GVK
func MatchKinds(obj runtime.Object, reqKind metav1.GroupVersionKind) error {
	gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
	if err != nil {
		return err
	}

	if !equalGVK(reqKind, gvk) {
		return fmt.Errorf("type mismatch: want: %s got: %s", gvk, reqKind)
	}
	return nil
}
