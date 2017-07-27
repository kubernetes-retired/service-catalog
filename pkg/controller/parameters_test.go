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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
)

func TestBuildParameters(t *testing.T) {
	stringSecret := &v1.Secret{
		Data: map[string][]byte{
			"secret-key": []byte("textFromSecret"),
		},
	}
	jsonSecret := &v1.Secret{
		Data: map[string][]byte{
			"secret-key": []byte("{ \"json\": true }"),
		},
	}

	cases := []struct {
		name           string
		parametersFrom []v1alpha1.ParametersFromSource
		parameters     []v1alpha1.Parameter
		secret         *v1.Secret
		expected       map[string]interface{}
		shouldSucceed  bool
	}{
		{
			name: "parameters: basic",
			parameters: []v1alpha1.Parameter{
				{
					Name:  "p1",
					Type:  v1alpha1.ValueTypeString,
					Value: "v1",
				}, {
					Name:  "p2",
					Type:  v1alpha1.ValueTypeString,
					Value: "v2",
				},
			},
			expected: map[string]interface{}{
				"p1": "v1",
				"p2": "v2",
			},
			shouldSucceed: true,
		},
		{
			name: "parameters: json",
			parameters: []v1alpha1.Parameter{
				{
					Name:  "json",
					Type:  v1alpha1.ValueTypeJSON,
					Value: "{ \"bool\": true, \"string\": \"str\", \"obj\": { \"child\": \"s\"} }",
				},
			},
			expected: map[string]interface{}{
				"json": map[string]interface{}{
					"bool":   true,
					"string": "str",
					"obj": map[string]interface{}{
						"child": "s",
					},
				},
			},
			shouldSucceed: true,
		},
		{
			name: "parameters: value from secret key",
			parameters: []v1alpha1.Parameter{
				{
					Name: "p1",
					Type: v1alpha1.ValueTypeString,
					ValueFrom: &v1alpha1.ParameterSource{
						SecretKeyRef: &v1alpha1.SecretKeyReference{
							Name: "secret",
							Key:  "secret-key",
						},
					},
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"p1": "textFromSecret",
			},
			shouldSucceed: true,
		},
		{
			name: "parameters: secret not found",
			parameters: []v1alpha1.Parameter{
				{
					Name: "p1",
					Type: v1alpha1.ValueTypeString,
					ValueFrom: &v1alpha1.ParameterSource{
						SecretKeyRef: &v1alpha1.SecretKeyReference{
							Name: "secret",
							Key:  "key",
						},
					},
				},
			},
			secret:        nil, // Not found
			shouldSucceed: false,
		},
		{
			name: "parameters: conflict",
			parameters: []v1alpha1.Parameter{
				{
					Name:  "p1",
					Type:  v1alpha1.ValueTypeString,
					Value: "v1",
				}, {
					Name:  "p1",
					Type:  v1alpha1.ValueTypeString,
					Value: "v2",
				},
			},
			expected: map[string]interface{}{
				"p1": "v2",
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: inline",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					Value: &runtime.RawExtension{
						Raw: []byte("{ \"bool\": true, \"string\": \"str\", \"obj\": { \"child\": \"s\"} }"),
					},
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"bool":   true,
				"string": "str",
				"obj": map[string]interface{}{
					"child": "s",
				},
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: secretKey with blob",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					SecretKeyRef: &v1alpha1.SecretKeyReference{
						Name: "secret",
						Key:  "secret-key",
					},
				},
			},
			secret: jsonSecret,
			expected: map[string]interface{}{
				"json": true,
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: secret",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					SecretRef: &v1alpha1.SecretReference{
						Type: v1alpha1.ValueTypeString,
						Name: "secret",
					},
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"secret-key": "textFromSecret",
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: secret with json",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					SecretRef: &v1alpha1.SecretReference{
						Type: v1alpha1.ValueTypeJSON,
						Name: "secret",
					},
				},
			},
			secret: jsonSecret,
			expected: map[string]interface{}{
				"secret-key": map[string]interface{}{
					"json": true,
				},
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom: name + secret",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					Name: "nesting",
					SecretRef: &v1alpha1.SecretReference{
						Type: v1alpha1.ValueTypeString,
						Name: "secret",
					},
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"nesting": map[string]interface{}{
					"secret-key": "textFromSecret",
				},
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom + parameters: normal",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					SecretRef: &v1alpha1.SecretReference{
						Type: v1alpha1.ValueTypeString,
						Name: "secret",
					},
				},
			},
			parameters: []v1alpha1.Parameter{
				{
					Name:  "p1",
					Type:  v1alpha1.ValueTypeString,
					Value: "v1",
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"secret-key": "textFromSecret",
				"p1":         "v1",
			},
			shouldSucceed: true,
		},
		{
			name: "parametersFrom + parameters: conflict",
			parametersFrom: []v1alpha1.ParametersFromSource{
				{
					Name: "p1",
					SecretRef: &v1alpha1.SecretReference{
						Type: v1alpha1.ValueTypeString,
						Name: "secret",
					},
				},
			},
			parameters: []v1alpha1.Parameter{
				{
					Name:  "p1",
					Type:  v1alpha1.ValueTypeString,
					Value: "v1",
				},
			},
			secret: stringSecret,
			expected: map[string]interface{}{
				"p1": "v1",
			},
			shouldSucceed: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testBuildParameters(t, tc.parametersFrom, tc.parameters, tc.secret, tc.expected, tc.shouldSucceed)
		})
	}
}

func testBuildParameters(t *testing.T, parametersFrom []v1alpha1.ParametersFromSource, parameters []v1alpha1.Parameter, secret *v1.Secret, expected map[string]interface{}, shouldSucceed bool) {
	// create a fake kube client
	fakeKubeClient := &clientgofake.Clientset{}
	if secret != nil {
		addGetSecretReaction(fakeKubeClient, secret)
	} else {
		addGetSecretNotFoundReaction(fakeKubeClient)
	}

	actual, err := buildParameters(fakeKubeClient, "test-ns", parametersFrom, parameters)
	if shouldSucceed {
		if err != nil {
			t.Fatalf("Failed to build parameters: %v", err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("incorrect result: diff \n%v", diff.ObjectGoPrintSideBySide(expected, actual))
		}
	} else {
		if err == nil {
			t.Fatal("Expected error, but got success")
		}
	}
}
