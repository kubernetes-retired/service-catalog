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

package class

import (
	"bytes"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/internal/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestDescribeCommand(t *testing.T) {
	const namespace = "default"
	testcases := []struct {
		name            string // Test Name
		cmd             string // Command to run
		golden          string // Relative path to a golden file, compared to the command output
		continueOnError bool   // Should the test stop immediately if the command fails or continue and capture the console output
	}{
		{
			name:   "describe class by name",
			cmd:    "describe class user-provided-service",
			golden: "../output/describe-class.txt",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			var fake runtime.Object
			fake = &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: "user provided service",
				},
			}

			svcatClient := svcatfake.NewSimpleClientset(fake)
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, namespace)
			output := &bytes.Buffer{}
			cxt := svcattest.NewContext(output, fakeApp)

			// Initialize the command arguments
			cmd := &describeCmd{
				Context: cxt,
			}
			// Capture all output: stderr and stdout
			cmd.Context.Output = output

			cmd.Run()
			test.AssertEqualsGoldenFile(t, tc.golden, output.String())
		})
	}
}
