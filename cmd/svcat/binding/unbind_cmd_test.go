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
	"errors"
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
	testing2 "k8s.io/client-go/testing"

	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
)

func TestUnbindCommand(t *testing.T) {
	const ns = "default"
	testcases := []struct {
		name           string
		fakeInstance   string
		fakeBindings   []string
		wait           bool
		bindingNames   []string
		instanceName   string
		wantOutput     string
		wantError      bool
		allowDiffOrder bool // whether the order of lines in the output can be different from the one in wantOutput
	}{
		{
			name:         "delete binding",
			fakeBindings: []string{"mybinding"},
			bindingNames: []string{"mybinding"},
			wantOutput:   "deleted mybinding",
		},
		{
			name:         "delete binding - fail",
			bindingNames: []string{"badbinding"},
			wantOutput:   "remove binding default/badbinding failed",
			wantError:    true,
		},
		{
			name:         "delete binding and wait",
			fakeBindings: []string{"mybinding"},
			bindingNames: []string{"mybinding"},
			wait:         true,
			wantOutput:   "waiting for the binding(s) to be deleted...\ndeleted mybinding\n",
		},
		{
			name:         "delete binding and wait - fail",
			fakeBindings: []string{"badbinding"},
			bindingNames: []string{"badbinding"},
			wait:         true,
			wantOutput:   "remove binding default/badbinding failed",
			wantError:    true,
		},
		{
			name:           "delete multiple bindings",
			fakeBindings:   []string{"binding1", "binding2"},
			bindingNames:   []string{"binding1", "binding2"},
			wantOutput:     "deleted binding1\ndeleted binding2",
			allowDiffOrder: true,
		},
		{
			name:         "delete multiple bindings - fail",
			fakeBindings: []string{"binding", "badbinding"},
			bindingNames: []string{"binding", "badbinding"},
			wantOutput:   "error:\n  remove binding default/badbinding failed: sabotaged\ndeleted binding\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:           "delete multiple bindings and wait",
			fakeBindings:   []string{"binding1", "binding2"},
			bindingNames:   []string{"binding1", "binding2"},
			wait:           true,
			wantOutput:     "waiting for the binding(s) to be deleted...\ndeleted binding1\ndeleted binding2\n",
			allowDiffOrder: true,
		},
		{
			name:         "delete multiple bindings and wait - fail",
			fakeBindings: []string{"binding", "badbinding"},
			bindingNames: []string{"binding", "badbinding"},
			wait:         true,
			wantOutput:   "error:\n  remove binding default/badbinding failed: sabotaged\nwaiting for the binding(s) to be deleted...\ndeleted binding\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance",
			fakeInstance: "myinstance",
			fakeBindings: []string{"binding"},
			instanceName: "myinstance",
			wantOutput:   "deleted binding\n",
		},
		{
			name:           "unbind instance - multiple bindings",
			fakeInstance:   "myinstance",
			fakeBindings:   []string{"binding1", "binding2"},
			instanceName:   "myinstance",
			wantOutput:     "deleted binding1\ndeleted binding2\n",
			allowDiffOrder: true,
		},
		{
			name:         "unbind instance - partial fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"binding1", "badbinding2"},
			instanceName: "myinstance",
			wantOutput:   "error:\n  remove binding default/badbinding2 failed: sabotaged\ndeleted binding1\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance and wait - partial fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"binding1", "badbinding2"},
			instanceName: "myinstance",
			wait:         true,
			wantOutput:   "error:\n  remove binding default/badbinding2 failed: sabotaged\nwaiting for the binding(s) to be deleted...\ndeleted binding1\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance - fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"badbinding"},
			instanceName: "myinstance",
			wantOutput:   "error:\n  remove binding default/badbinding failed: sabotaged\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance and wait - fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"badbinding"},
			instanceName: "myinstance",
			wait:         true,
			wantOutput:   "error:\n  remove binding default/badbinding failed: sabotaged\ncould not remove all bindings",
			wantError:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			var fakes []runtime.Object
			if tc.fakeInstance != "" {
				fakes = append(fakes, &v1beta1.ServiceInstance{
					ObjectMeta: v1.ObjectMeta{
						Namespace: ns,
						Name:      tc.fakeInstance,
					},
				})
			}
			for _, name := range tc.fakeBindings {
				fakes = append(fakes, &v1beta1.ServiceBinding{
					ObjectMeta: v1.ObjectMeta{
						Namespace: ns,
						Name:      name,
					},
					Spec: v1beta1.ServiceBindingSpec{ServiceInstanceRef: v1beta1.LocalObjectReference{Name: tc.fakeInstance}},
				})
			}
			svcatClient := svcatfake.NewSimpleClientset(fakes...)
			output := &bytes.Buffer{}
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, "default")
			cxt := svcattest.NewContext(output, fakeApp)

			// Sabotage any binding with "bad" in the name
			svcatClient.PrependReactor("delete", "servicebindings",
				func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
					a, _ := action.(testing2.DeleteAction) // cast is guaranteed because we are only handling gets in this reactor
					if strings.Contains(a.GetName(), "bad") {
						return true, nil, errors.New("sabotaged")
					}
					return false, nil, nil
				})

			// Initialize the command arguments
			cmd := &unbindCmd{
				Namespaced: command.NewNamespaced(cxt),
				Waitable:   command.NewWaitable(),
			}
			cmd.Namespace = ns
			cmd.bindingNames = tc.bindingNames
			cmd.instanceName = tc.instanceName
			cmd.Wait = tc.wait

			err := cmd.Run()

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
			if !outputMatches(gotOutput, tc.wantOutput, tc.allowDiffOrder) {
				t.Errorf("unexpected output \n\nWANT:\n%q\n\nGOT:\n%q\n", tc.wantOutput, gotOutput)
			}
		})
	}
}

func outputMatches(gotOutput string, wantOutput string, allowDifferentLineOrder bool) bool {
	if !allowDifferentLineOrder {
		return strings.Contains(gotOutput, wantOutput)
	}

	gotLines := strings.Split(gotOutput, "\n")
	wantLines := strings.Split(wantOutput, "\n")

	for _, wantLine := range wantLines {
		found := false
		for _, gotLine := range gotLines {
			if strings.Contains(gotLine, wantLine) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
