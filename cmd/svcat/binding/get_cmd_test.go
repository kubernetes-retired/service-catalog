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

package binding

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
)

func TestGetCommand(t *testing.T) {
	const namespace = "default"
	testcases := []struct {
		name          string
		fakeBindings  []string
		bindingName   string
		outputFormat  string
		expectedError string
		wantError     bool
	}{
		{
			name:          "get non existing binding",
			fakeBindings:  []string{},
			bindingName:   "mybinding",
			expectedError: "unable to get binding '" + namespace + ".mybinding'",
			wantError:     true,
		},
		{
			name:         "get all existing bindings with unknown output format",
			fakeBindings: []string{"myfirstbinding", "mysecondbinding"},
			bindingName:  "",
			outputFormat: "unknown",
			wantError:    false,
		},
		{
			name:         "get all existing bindings with json output format",
			fakeBindings: []string{"myfirstbinding", "mysecondbinding"},
			bindingName:  "",
			outputFormat: output.FormatJSON,
			wantError:    false,
		},
		{
			name:         "get all existing bindings with yaml output format",
			fakeBindings: []string{"myfirstbinding", "mysecondbinding"},
			bindingName:  "",
			outputFormat: output.FormatYAML,
			wantError:    false,
		},
		{
			name:         "get all existing bindings with table output format",
			fakeBindings: []string{"myfirstbinding", "mysecondbinding"},
			bindingName:  "",
			outputFormat: output.FormatTable,
			wantError:    false,
		},
		{
			name:         "get existing binding with unknown output format",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			outputFormat: "unknown",
			wantError:    false,
		},
		{
			name:         "get existing binding with json output format",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			outputFormat: output.FormatJSON,
			wantError:    false,
		},
		{
			name:         "get existing binding with yaml output format",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			outputFormat: output.FormatYAML,
			wantError:    false,
		},
		{
			name:         "get existing binding with table output format",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			outputFormat: output.FormatTable,
			wantError:    false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			var fakes []runtime.Object
			for _, name := range tc.fakeBindings {
				fakes = append(fakes, &v1beta1.ServiceBinding{
					ObjectMeta: v1.ObjectMeta{
						Namespace: namespace,
						Name:      name,
					},
					Spec: v1beta1.ServiceBindingSpec{},
				})
			}

			svcatClient := svcatfake.NewSimpleClientset(fakes...)
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, namespace)
			output := &bytes.Buffer{}
			cxt := svcattest.NewContext(output, fakeApp)

			// Initialize the command arguments
			cmd := &getCmd{
				Namespaced: command.NewNamespaced(cxt),
				Formatted:  command.NewFormatted(),
			}
			cmd.Namespace = namespace
			cmd.name = tc.bindingName
			cmd.OutputFormat = tc.outputFormat

			err := cmd.Run()

			if tc.wantError {
				if err == nil {
					t.Errorf("expected a non-zero exit code, but the command succeeded")
				}

				errorOutput := err.Error()
				if !strings.Contains(errorOutput, tc.expectedError) {
					t.Errorf("Unexpected output:\n\nExpected:\n%q\n\nActual:\n%q\n", tc.expectedError, errorOutput)
				}
			}
			if !tc.wantError && err != nil {
				t.Errorf("expected the command to succeed but it failed with %q", err)
			}
		})
	}
}
