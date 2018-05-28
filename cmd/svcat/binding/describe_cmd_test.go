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
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
)

func TestDescribeCommand(t *testing.T) {
	const namespace = "default"
	testcases := []struct {
		name         string
		fakeBindings []string
		bindingName  string
		expected     string
		wantError    bool
	}{
		{
			name:         "describe non existing binding",
			fakeBindings: []string{},
			bindingName:  "mybinding",
			expected:     "unable to get binding '" + namespace + ".mybinding': servicebindings.servicecatalog.k8s.io \"mybinding\" not found",
			wantError:    true,
		},
		{
			name:         "describe existing binding",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			expected:     "  Name:        mybinding  \n  Namespace:   " + namespace + "    \n  Status:                 \n  Secret:                 \n  Instance:               \n\nParameters:\n  No parameters defined\n",
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
			cmd := &describeCmd{
				Namespaced: command.NewNamespacedCommand(cxt),
			}
			cmd.Namespace = namespace
			cmd.name = tc.bindingName

			err := cmd.Run()

			if tc.wantError && err == nil {
				t.Errorf("expected a non-zero exit code, but the command succeeded")
			}
			if !tc.wantError && err != nil {
				t.Errorf("expected the command to succeed but it failed with %q", err)
			}

			actual := output.String()
			if err != nil {
				actual += err.Error()
			}
			if actual != tc.expected {
				t.Errorf("Unexpected output:\n\nExpected:\n%q\n\nActual:\n%q\n", tc.expected, actual)
			}
		})
	}
}
