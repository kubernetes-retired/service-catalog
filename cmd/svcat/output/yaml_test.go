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

package output

import (
	"bytes"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestWriteParameters(t *testing.T) {
	testcases := []struct {
		name       string                // Test name
		parameters *runtime.RawExtension // Parameters tested
		output     string                // Expected output
	}{
		{"Nil parameter", nil, ""},
		{"JSON w/data parameter", &runtime.RawExtension{Raw: []byte(`{"foo":"bar"}`)}, "\nParameters:\n  foo: bar\n"},
		{"JSON empty parameter", &runtime.RawExtension{Raw: []byte(`{}`)}, "\nParameters:\n  {}\n"},
		{"String parameter", &runtime.RawExtension{Raw: []byte("param")}, "\nParameters:\nparam\n"},
		{"Empty string parameter", &runtime.RawExtension{Raw: []byte("")}, "\nParameters:\n  No parameters defined\n"},
	}

	for _, tc := range testcases {
		output := &bytes.Buffer{}
		writeParameters(output, tc.parameters)
		if tc.output != output.String() {
			t.Errorf("%v: Output mismatch: expected \"%v\", actual \"%v\"", tc.name, tc.output, output.String())
		}
	}
}
