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
package versions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
)

func TestVersionCommand(t *testing.T) {
	testcases := []struct {
		name       string
		client     bool
		server     bool
		wantOutput string
		wantError  bool
	}{
		{
			name:       "show client version only",
			client:     true,
			server:     false,
			wantOutput: "Client Version: UNKNOWN\n",
			wantError:  false,
		},
		{
			name:       "show server & client version",
			client:     true,
			server:     true,
			wantOutput: "Client Version: UNKNOWN\nServer Version: v0.0.0-master+$Format:%h$\n",
			wantError:  false,
		},
		/*
			{
				name:       "show server & client version - fail",
				client:     true,
				server:     true,
				wantOutput: "Client Version: UNKNOWN\nunable to get version",
				wantError:  false,
			},
		*/
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up fake data
			//TODO: feed fakes array
			var fakes []runtime.Object
			k8sClient := k8sfake.NewSimpleClientset(fakes...)
			svcatClient := svcatfake.NewSimpleClientset()
			output := &bytes.Buffer{}
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, "default")

			cxt := &command.Context{
				Output: output,
				App:    fakeApp,
			}
			versionCommand := &versionCmd{
				cxt,
				tc.client,
				tc.server,
			}

			err := versionCommand.Run()
			if tc.wantError && err == nil {
				t.Errorf("expected a non-zero exit code, but the command succeeded")
			}
			if !tc.wantError && err != nil {
				t.Errorf("expected the command to succeed but it failed with %q", err)
			}

			gotOutput := output.String()
			if err != nil {
				gotOutput += err.Error()
			}
			if !strings.Contains(gotOutput, tc.wantOutput) {
				t.Errorf("unexpected output \n\nWANT:\n%q\n\nGOT:\n%q\n", tc.wantOutput, gotOutput)
			}
		})
	}
}
