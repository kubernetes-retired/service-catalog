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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	clientgofake "k8s.io/client-go/kubernetes/fake"
)

func TestBuildParameters(t *testing.T) {
	secret := &corev1.Secret{
		Data: map[string][]byte{
			"json-key":   []byte("{ \"json\": true }"),
			"string-key": []byte("textFromSecret"),
		},
	}

	cases := []struct {
		name                                  string
		parametersFrom                        []v1beta1.ParametersFromSource
		parameters                            *runtime.RawExtension
		secret                                *corev1.Secret
		expectedParameters                    map[string]interface{}
		expectedParametersWithSecretsRedacted map[string]interface{}
		shouldSucceed                         bool
	}{
		{
			name: "parameters: basic",
			parameters: &runtime.RawExtension{
				Raw: []byte(`{ "p1": "v1", "p2": "v2" }`),
			},
			expectedParameters: map[string]interface{}{
				"p1": "v1",
				"p2": "v2",
			},
			expectedParametersWithSecretsRedacted: map[string]interface{}{
				"p1": "v1",
				"p2": "v2",
			},
			shouldSucceed: true,
		},
		{
			name: "parameters: invalid JSON",
			parameters: &runtime.RawExtension{
				Raw: []byte("not a JSON"),
			},
			shouldSucceed: false,
		},
		{
			name: "parametersFrom: secretKey with blob",
			parametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret",
						Key:  "json-key",
					},
				},
			},
			secret: secret,
			expectedParameters: map[string]interface{}{
				"json": true,
			},
			expectedParametersWithSecretsRedacted: map[string]interface{}{
				"json": "<redacted>",
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: secretKey with invalid blob",
			parametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret",
						Key:  "string-key",
					},
				},
			},
			secret:        secret,
			shouldSucceed: false,
		},
		{
			name: "parametersFrom + parameters: normal",
			parametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret",
						Key:  "json-key",
					},
				},
			},
			parameters: &runtime.RawExtension{
				Raw: []byte(`{ "p1": "v1" }`),
			},
			secret: secret,
			expectedParameters: map[string]interface{}{
				"json": true,
				"p1":   "v1",
			},
			expectedParametersWithSecretsRedacted: map[string]interface{}{
				"json": "<redacted>",
				"p1":   "v1",
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom + parameters: conflict",
			parametersFrom: []v1beta1.ParametersFromSource{
				{
					SecretKeyRef: &v1beta1.SecretKeyReference{
						Name: "secret",
						Key:  "json-key",
					},
				},
			},
			parameters: &runtime.RawExtension{
				Raw: []byte(`{ "json": "v1" }`),
			},
			secret:        secret,
			shouldSucceed: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testBuildParameters(t, tc.parametersFrom, tc.parameters, tc.secret, tc.expectedParameters, tc.expectedParametersWithSecretsRedacted, tc.shouldSucceed)
		})
	}
}

func testBuildParameters(t *testing.T, parametersFrom []v1beta1.ParametersFromSource, parameters *runtime.RawExtension, secret *corev1.Secret, expected map[string]interface{}, expectedWithSecretsRdacted map[string]interface{}, shouldSucceed bool) {
	// create a fake kube client
	fakeKubeClient := &clientgofake.Clientset{}
	if secret != nil {
		addGetSecretReaction(fakeKubeClient, secret)
	} else {
		addGetSecretNotFoundReaction(fakeKubeClient)
	}

	actual, actualWithSecretsRedacted, err := buildParameters(fakeKubeClient, "test-ns", parametersFrom, parameters)
	if shouldSucceed {
		if err != nil {
			t.Fatalf("Failed to build parameters: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("incorrect result: diff \n%v", diff.ObjectGoPrintSideBySide(expected, actual))
		}
		if !reflect.DeepEqual(actualWithSecretsRedacted, expectedWithSecretsRdacted) {
			t.Fatalf("incorrect result with redacted secrets: diff \n%v", diff.ObjectGoPrintSideBySide(expectedWithSecretsRdacted, actualWithSecretsRedacted))
		}
	} else {
		if err == nil {
			t.Fatal("Expected error, but got success")
		}
	}
}

func TestGenerateChecksumOfParameters(t *testing.T) {
	cases := []struct {
		name             string
		oldParams        map[string]interface{}
		newParams        map[string]interface{}
		expectedEquality bool
	}{
		{
			name: "same",
			oldParams: map[string]interface{}{
				"a": "first",
				"b": 2,
			},
			newParams: map[string]interface{}{
				"a": "first",
				"b": 2,
			},
			expectedEquality: true,
		},
		{
			name: "different",
			oldParams: map[string]interface{}{
				"a": "first",
				"b": 2,
			},
			newParams: map[string]interface{}{
				"a": "first",
				"b": 3,
			},
			expectedEquality: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			oldChecksum, err := generateChecksumOfParameters(tc.oldParams)
			if err != nil {
				t.Fatalf("failed to generate checksum: %v", err)
			}
			newChecksum, err := generateChecksumOfParameters(tc.newParams)
			if err != nil {
				t.Fatalf("failed to generate checksum: %v", err)
			}
			actualEquality := oldChecksum == newChecksum
			if e, a := tc.expectedEquality, actualEquality; e != a {
				expectedCondition := "be equal"
				if !tc.expectedEquality {
					expectedCondition = "not be equal"
				}
				t.Fatalf("expected checksums to %s: old %q, new %q", expectedCondition, oldChecksum, newChecksum)
			}
		})
	}
}

func TestMergeParameters(t *testing.T) {
	testParams := `{"a":1,"d":{"e":5}}`
	testcases := []struct {
		name     string
		params   *string
		defaults *string
		want     *string
	}{
		{name: "neither params or defaults defined", params: nil, defaults: nil, want: nil},
		{name: "accept provided params when no defaults defined", params: stringPtr(testParams), defaults: nil, want: stringPtr(testParams)},
		{name: "accept provided params when default is empty string", params: stringPtr(testParams), defaults: stringPtr(""), want: stringPtr(testParams)},
		{name: "accept provided params when default is empty object", params: stringPtr(testParams), defaults: stringPtr("{}"), want: stringPtr(testParams)},
		{name: "use default params when no params defined", params: nil, defaults: stringPtr(testParams), want: stringPtr(testParams)},
		{name: "use default params when params is empty string", params: stringPtr(""), defaults: stringPtr(testParams), want: stringPtr(testParams)},
		{name: "use default params when params is empty object", params: stringPtr("{}"), defaults: stringPtr(testParams), want: stringPtr(testParams)},
		{name: "merge params with defaults", params: stringPtr(`{"b":2}`), defaults: stringPtr(testParams), want: stringPtr(`{"a":1,"b":2,"d":{"e":5}}`)},
		{name: "merge params with defaults, override wins", params: stringPtr(`{"a":2}`), defaults: stringPtr(testParams), want: stringPtr(`{"a":2,"d":{"e":5}}`)},
		{name: "merge params with defaults, nested merge", params: stringPtr(`{"d":{"e":2,"f":3}}`), defaults: stringPtr(testParams), want: stringPtr(`{"a":1,"d":{"e":2,"f":3}}`)},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var rawParams, rawDefaults, wantParams *runtime.RawExtension
			if tc.params != nil {
				rawParams = &runtime.RawExtension{Raw: []byte(*tc.params)}
			}
			if tc.defaults != nil {
				rawDefaults = &runtime.RawExtension{Raw: []byte(*tc.defaults)}
			}
			if tc.want != nil {
				wantParams = &runtime.RawExtension{Raw: []byte(*tc.want)}
			}

			gotParams, err := mergeParameters(rawParams, rawDefaults)

			if err != nil {
				t.Fatal(err)
			}

			// shenanigans so that it's easier to compare and print out the real values of the params
			wantPretty := "nil"
			if wantParams != nil {
				wantPretty = string(wantParams.Raw)
			}
			gotPretty := "nil"
			if gotParams != nil {
				gotPretty = string(gotParams.Raw)
			}
			if wantPretty != gotPretty {
				t.Fatalf("WANT:\t%v\nGOT:\t%v", wantPretty, gotPretty)
			}
		})
	}
}

func stringPtr(val string) *string {
	return &val
}
