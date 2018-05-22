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
		name         string
		fakeInstance string
		fakeBindings []string
		wait         bool
		bindingName  string
		instanceName string
		wantOutput   string
		wantError    bool
	}{
		{
			name:         "delete binding",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			wantOutput:   "deleted mybinding",
		},
		{
			name:        "delete binding - fail",
			bindingName: "badbinding",
			wantOutput:  "remove binding default/badbinding failed",
			wantError:   true,
		},
		{
			name:         "delete binding and wait",
			fakeBindings: []string{"mybinding"},
			bindingName:  "mybinding",
			wait:         true,
			wantOutput:   "waiting for the binding to be deleted...\ndeleted mybinding\n",
		},
		{
			name:         "delete binding and wait - fail",
			fakeBindings: []string{"badbinding"},
			bindingName:  "badbinding",
			wait:         true,
			wantOutput:   "remove binding default/badbinding failed",
			wantError:    true,
		},
		{
			name:         "unbind instance",
			fakeInstance: "myinstance",
			fakeBindings: []string{"binding1", "binding2"},
			instanceName: "myinstance",
			wantOutput:   "deleted binding2\ndeleted binding1\n",
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
			wantOutput:   "error:\n  remove binding default/badbinding2 failed: sabotaged\nwaiting for the bindings to be deleted...\ndeleted binding1\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance - fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"badbinding1", "badbinding2"},
			instanceName: "myinstance",
			wantOutput:   "error:\n  remove binding default/badbinding2 failed: sabotaged\n  remove binding default/badbinding1 failed: sabotaged\ncould not remove all bindings",
			wantError:    true,
		},
		{
			name:         "unbind instance and wait - fail",
			fakeInstance: "myinstance",
			fakeBindings: []string{"badbinding1", "badbinding2"},
			instanceName: "myinstance",
			wait:         true,
			wantOutput:   "error:\n  remove binding default/badbinding2 failed: sabotaged\n  remove binding default/badbinding1 failed: sabotaged\ncould not remove all bindings",
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
				Namespaced:      command.NewNamespacedCommand(cxt),
				WaitableCommand: command.NewWaitableCommand(),
			}
			cmd.Namespace = ns
			cmd.bindingName = tc.bindingName
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
			if !strings.Contains(gotOutput, tc.wantOutput) {
				t.Errorf("unexpected output \n\nWANT:\n%q\n\nGOT:\n%q\n", tc.wantOutput, gotOutput)
			}
		})
	}
}
