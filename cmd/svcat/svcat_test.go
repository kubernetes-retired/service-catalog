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

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"

	clientgotesting "k8s.io/client-go/testing"

	"encoding/json"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/plugin"
	"github.com/kubernetes-incubator/service-catalog/internal/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var catalogRequestRegex = regexp.MustCompile("/apis/servicecatalog.k8s.io/v1beta1/(.*)")
var coreRequestRegex = regexp.MustCompile("/api/v1/(.*)")

func TestCommandValidation(t *testing.T) {
	testcases := []struct {
		name      string // Test Name
		cmd       string // Command to run
		wantError string // Substring that should be present in the error, empty indicates no error
	}{
		{"viper bug workaround: provision", "provision name --class class --plan plan", ""},
		{"viper bug workaround: bind", "bind name", ""},
		{"describe broker requires name", "describe broker", "name is required"},
		{"describe class requires name", "describe class", "name or uuid is required"},
		{"describe plan requires name", "describe plan", "name or uuid is required"},
		{"describe instance requires name", "describe instance", "name is required"},
		{"describe binding requires name", "describe binding", "name is required"},
		{"unbind requires arg", "unbind", "instance or binding name is required"},
		{"sync requires names", "sync broker", "name is required"},
		{"deprovision requires name", "deprovision", "name is required"},
		{"provision does not accept --param and --params-json",
			`provision name --class class --plan plan --params-json '{}' --param k=v`,
			"--params-json cannot be used with --param"},
		{"bind does not accept --param and --params-json",
			`bind name --params-json '{}' --param k=v`,
			"--params-json cannot be used with --param"},
		{"completion no shell specified", "completion", "Shell not specified"},
		{"completion too many args", "completion arg0 arg1", "Too many arguments. Expected only the shell type"},
		{"completion unsupported shell", "completion unsupportedShell", "Unsupported shell type \"unsupportedShell\""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validateCommand(t, tc.cmd, tc.wantError)
		})
	}
}

func TestCommandOutput(t *testing.T) {
	testcases := []struct {
		name            string // Test Name
		cmd             string // Command to run
		golden          string // Relative path to a golden file, compared to the command output
		continueOnError bool   // Should the test stop immediately if the command fails or continue and capture the console output
	}{
		{name: "list all brokers", cmd: "get brokers", golden: "output/get-brokers.txt"},
		{name: "list all brokers (json)", cmd: "get brokers -o json", golden: "output/get-brokers.json"},
		{name: "list all brokers (yaml)", cmd: "get brokers -o yaml", golden: "output/get-brokers.yaml"},
		{name: "get broker", cmd: "get broker ups-broker", golden: "output/get-broker.txt"},
		{name: "get broker (json)", cmd: "get broker ups-broker -o json", golden: "output/get-broker.json"},
		{name: "get broker (yaml)", cmd: "get broker ups-broker -o yaml", golden: "output/get-broker.yaml"},
		{name: "describe broker", cmd: "describe broker ups-broker", golden: "output/describe-broker.txt"},

		{name: "list all classes", cmd: "get classes", golden: "output/get-classes.txt"},
		{name: "list all classes (json)", cmd: "get classes -o json", golden: "output/get-classes.json"},
		{name: "list all classes (yaml)", cmd: "get classes -o yaml", golden: "output/get-classes.yaml"},
		{name: "get class by name", cmd: "get class user-provided-service", golden: "output/get-class.txt"},
		{name: "get class by name (json)", cmd: "get class user-provided-service -o json", golden: "output/get-class.json"},
		{name: "get class by name (yaml)", cmd: "get class user-provided-service -o yaml", golden: "output/get-class.yaml"},
		{name: "get class by uuid", cmd: "get class --uuid 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468", golden: "output/get-class.txt"},
		{name: "describe class by name", cmd: "describe class user-provided-service", golden: "output/describe-class.txt"},
		{name: "describe class uuid", cmd: "describe class --uuid 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468", golden: "output/describe-class.txt"},

		{name: "list all plans", cmd: "get plans", golden: "output/get-plans.txt"},
		{name: "list all plans (json)", cmd: "get plans -o json", golden: "output/get-plans.json"},
		{name: "list all plans (yaml)", cmd: "get plans -o yaml", golden: "output/get-plans.yaml"},
		{name: "get plan by name", cmd: "get plan default", golden: "output/get-plan.txt"},
		{name: "get plan by name (json)", cmd: "get plan default -o json", golden: "output/get-plan.json"},
		{name: "get plan by name (yaml)", cmd: "get plan default -o yaml", golden: "output/get-plan.yaml"},
		{name: "get plan by uuid", cmd: "get plan --uuid 86064792-7ea2-467b-af93-ac9694d96d52", golden: "output/get-plan.txt"},
		{name: "get plan by class/plan name combo", cmd: "get plan user-provided-service/default", golden: "output/get-plan.txt"},
		{name: "get plan by class name", cmd: "get plan --class user-provided-service", golden: "output/get-plans-by-class.txt"},
		{name: "get plan by class/plan name combo", cmd: "get plan --class user-provided-service default", golden: "output/get-plan.txt"},
		{name: "get plan by class/plan uuid combo", cmd: "get plan --uuid --class 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468 86064792-7ea2-467b-af93-ac9694d96d52", golden: "output/get-plan.txt"},
		{name: "get plan by class uuid", cmd: "get plan --uuid --class 4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468", golden: "output/get-plans-by-class.txt"},
		{name: "describe plan by name", cmd: "describe plan default", golden: "output/describe-plan.txt"},
		{name: "describe plan by uuid", cmd: "describe plan --uuid 86064792-7ea2-467b-af93-ac9694d96d52", golden: "output/describe-plan.txt"},
		{name: "describe plan by class/plan name combo", cmd: "describe plan user-provided-service/default", golden: "output/describe-plan.txt"},
		{name: "describe plan with schemas", cmd: "describe plan premium", golden: "output/describe-plan-with-schemas.txt"},
		{name: "describe plan without schemas", cmd: "describe plan premium --show-schemas=false", golden: "output/describe-plan-without-schemas.txt"},

		{name: "list all instances in a namespace", cmd: "get instances -n test-ns", golden: "output/get-instances.txt"},
		{name: "list all instances in a namespace (json)", cmd: "get instances -n test-ns -o json", golden: "output/get-instances.json"},
		{name: "list all instances in a namespace (yaml)", cmd: "get instances -n test-ns -o yaml", golden: "output/get-instances.yaml"},
		{name: "list all instances", cmd: "get instances --all-namespaces", golden: "output/get-instances-all-namespaces.txt"},
		{name: "get instance", cmd: "get instance ups-instance -n test-ns", golden: "output/get-instance.txt"},
		{name: "get instance (json)", cmd: "get instance ups-instance -n test-ns -o json", golden: "output/get-instance.json"},
		{name: "get instance (yaml)", cmd: "get instance ups-instance -n test-ns -o yaml", golden: "output/get-instance.yaml"},
		{name: "describe instance", cmd: "describe instance ups-instance -n test-ns", golden: "output/describe-instance.txt"},

		{name: "list all bindings in a namespace", cmd: "get bindings -n test-ns", golden: "output/get-bindings.txt"},
		{name: "list all bindings in a namespace (json)", cmd: "get bindings -n test-ns -o json", golden: "output/get-bindings.json"},
		{name: "list all bindings in a namespace (yaml)", cmd: "get bindings -n test-ns -o yaml", golden: "output/get-bindings.yaml"},
		{name: "list all bindings", cmd: "get bindings --all-namespaces", golden: "output/get-bindings-all-namespaces.txt"},
		{name: "get binding", cmd: "get binding ups-binding -n test-ns", golden: "output/get-binding.txt"},
		{name: "get binding (json)", cmd: "get binding ups-binding -n test-ns -o json", golden: "output/get-binding.json"},
		{name: "get binding (yaml)", cmd: "get binding ups-binding -n test-ns -o yaml", golden: "output/get-binding.yaml"},
		{name: "describe binding", cmd: "describe binding ups-binding -n test-ns", golden: "output/describe-binding.txt"},
		{name: "describe binding and decode secret", cmd: "describe binding ups-binding -n test-ns --show-secrets", golden: "output/describe-binding-show-secrets.txt"},

		{name: "completion bash", cmd: "completion bash", golden: "output/completion-bash.txt"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output := executeCommand(t, tc.cmd, tc.continueOnError)
			test.AssertEqualsGoldenFile(t, tc.golden, output)
		})
	}
}

// If you add a new command to svcat, this test will fail, because the plugin.yaml
// golden file will be out of date. To fix this, run:
//
//	go test ./cmd/svcat/... -update
//
//
// once. This command updates the golden file according to your new command.
// After you run the update, make sure your tests pass against the new golden
// file:
//
// 	go test ./cmd/svcat/...
//
func TestGenerateManifest(t *testing.T) {
	svcat := buildRootCommand(newContext())

	m := &plugin.Manifest{}
	m.Load(svcat)

	got, err := yaml.Marshal(&m)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	test.AssertEqualsGoldenFile(t, "plugin.yaml", string(got))
}

// TestNamespacedCommands verifies that all commands that are namespace scoped
// handle setting the namespace using the current context, --namespace and --all-namespaces flags.
func TestNamespacedCommands(t *testing.T) {
	const contextNS = "from-context"
	const flagNS = "from-flag"
	const allNS = ""

	testcases := []struct {
		name   string
		cmd    string
		wantNS string
	}{
		{name: "get instances with flag namespace", cmd: "get instances --namespace " + flagNS, wantNS: flagNS},
		{name: "get instances with context namespace", cmd: "get instances", wantNS: contextNS},
		{name: "get all instances", cmd: "get instances --all-namespaces", wantNS: allNS},

		{name: "describe instance with flag namespace", cmd: "describe instance NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "describe instance with context namespace", cmd: "describe instances NAME", wantNS: contextNS},

		{name: "provision with flag namespace", cmd: "provision --class CLASS --plan PLAN NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "provision with context namespace", cmd: "provision --class CLASS --plan PLAN NAME", wantNS: contextNS},

		{name: "deprovision with flag namespace", cmd: "deprovision NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "deprovision with context namespace", cmd: "deprovision NAME", wantNS: contextNS},

		{name: "bind with flag namespace", cmd: "bind NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "bind with context namespace", cmd: "bind NAME", wantNS: contextNS},

		{name: "unbind with flag namespace", cmd: "unbind NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "unbind with context namespace", cmd: "unbind NAME", wantNS: contextNS},

		{name: "get bindings with flag namespace", cmd: "get bindings --namespace " + flagNS, wantNS: flagNS},
		{name: "get bindings with context namespace", cmd: "get bindings", wantNS: contextNS},
		{name: "get all bindings", cmd: "get bindings --all-namespaces", wantNS: allNS},

		{name: "describe binding with flag namespace", cmd: "describe binding NAME --namespace " + flagNS, wantNS: flagNS},
		{name: "describe binding with context namespace", cmd: "describe binding NAME", wantNS: contextNS},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()

			cxt := newContext()
			cxt.App = &svcat.App{
				CurrentNamespace: contextNS,
				SDK:              &servicecatalog.SDK{ServiceCatalogClient: fakeClient},
			}
			cxt.Output = ioutil.Discard

			executeFakeCommand(t, tc.cmd, cxt, true)

			gotNamespace := fakeClient.Actions()[0].GetNamespace()
			if tc.wantNS != gotNamespace {
				t.Fatalf("the wrong namespace was used. WANT: %q, GOT: %q", tc.wantNS, gotNamespace)
			}
		})
	}
}

// TestParametersForBinding confirms that parameters given as --param or --param-json work the same way
func TestParametersForBinding(t *testing.T) {
	testcases := []struct {
		name   string
		cmd    string
		params map[string]interface{}
	}{
		{
			name: "bind with --param",
			cmd:  "bind NAME --param foo=bar --param baz=boo",
			params: map[string]interface{}{
				"foo": "bar",
				"baz": "boo",
			},
		},
		{
			name: "bind with --params-json",
			cmd:  "bind NAME --params-json {\"foo\":\"bar\",\"baz\":\"boo\"}",
			params: map[string]interface{}{
				"foo": "bar",
				"baz": "boo",
			},
		},
		{
			name: "bind with --params-json with a sub object",
			cmd:  "bind NAME --params-json {\"foo\":{\"faa\":\"bar\",\"baz\":\"boo\"}}",
			params: map[string]interface{}{
				"foo": map[string]interface{}{
					"faa": "bar",
					"baz": "boo",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset()

			cxt := newContext()
			cxt.App = &svcat.App{
				SDK: &servicecatalog.SDK{ServiceCatalogClient: fakeClient},
			}
			cxt.Output = ioutil.Discard

			executeFakeCommand(t, tc.cmd, cxt, true)

			if c := fakeClient.Actions(); len(c) != 1 {
				t.Fatal("Expected only 1 action, got ", c)
			}
			action := fakeClient.Actions()[0]

			if action.GetVerb() != "create" {
				t.Fatal("Expected a create action, but got ", action.GetVerb())
			}
			createAction, ok := action.(clientgotesting.CreateAction)
			if !ok {
				t.Fatal(t, "Unexpected type; failed to convert action %+v to CreateAction", action)

			}

			fakeObject := createAction.GetObject()

			binding, ok := fakeObject.(*v1beta1.ServiceBinding)
			if !ok {
				t.Fatal(t, "Failed to cast object to binding: ", fakeObject)
			}

			var params map[string]interface{}
			if err := json.Unmarshal(binding.Spec.Parameters.Raw, &params); err != nil {
				t.Error("failed to unmarshal binding.Spec.Parameters")
			}

			if eq := reflect.DeepEqual(params, tc.params); !eq {
				t.Errorf("parameters mismatch, \nwant: %+v, \ngot: %+v", tc.params, params)
			}
		})
	}
}

// executeCommand runs a svcat command against a fake k8s api,
// returning the cli output.
func executeCommand(t *testing.T, cmd string, continueOnErr bool) string {
	// Fake the k8s api server
	apisvr := newAPIServer()
	defer apisvr.Close()

	// Generate a test kubeconfig pointing at the server
	kubeconfig, err := writeTestKubeconfig(apisvr.URL)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer os.Remove(kubeconfig)

	// Setup the svcat command
	svcat, _, err := buildCommand(cmd, newContext(), kubeconfig)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Capture all output: stderr and stdout
	output := &bytes.Buffer{}
	svcat.SetOutput(output)

	err = svcat.Execute()
	if err != nil && !continueOnErr {
		t.Fatalf("%+v", err)
	}

	return output.String()
}

// executeCommand runs a svcat command against a fake k8s api,
// returning the cli output.
func executeFakeCommand(t *testing.T, cmd string, fakeContext *command.Context, continueOnErr bool) string {
	// Setup the svcat command
	svcat, _, err := buildCommand(cmd, fakeContext, "")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Capture all output: stderr and stdout
	output := &bytes.Buffer{}
	svcat.SetOutput(output)

	err = svcat.Execute()
	if err != nil && !continueOnErr {
		t.Fatalf("%+v", err)
	}

	return output.String()
}

// validateCommand validates a svcat command arguments
func validateCommand(t *testing.T, cmd string, wantError string) {
	// Fake the k8s api server
	apisvr := newAPIServer()
	defer apisvr.Close()

	// Generate a test kubeconfig pointing at the server
	kubeconfig, err := writeTestKubeconfig(apisvr.URL)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer os.Remove(kubeconfig)

	// Setup the svcat command
	svcat, targetCmd, err := buildCommand(cmd, newContext(), kubeconfig)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Skip running the actual command because we are only validating
	targetCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// Capture all output: stderr and stdout
	output := &bytes.Buffer{}
	svcat.SetOutput(output)

	err = svcat.Execute()
	if wantError == "" {
		if err != nil {
			t.Fatalf("%+v", err)
		}
	} else {
		gotError := ""
		if err != nil {
			gotError = err.Error()
		}
		if !strings.Contains(gotError, wantError) {
			t.Fatalf("unexpected error \n\nWANT:\n%q\n\nGOT:\n%q\n", wantError, gotError)
		}
	}
}

// buildCommand parses a command string.
func buildCommand(cmd string, cxt *command.Context, kubeconfig string) (rootCmd *cobra.Command, targetCmd *cobra.Command, err error) {
	rootCmd = buildRootCommand(cxt)
	args := strings.Split(cmd, " ")
	args = append(args, "--kubeconfig", kubeconfig)
	rootCmd.SetArgs(args)

	targetCmd, _, err = rootCmd.Find(args)

	return rootCmd, targetCmd, err
}

func newContext() *command.Context {
	return &command.Context{
		Viper: viper.New(),
	}
}

func newAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(apihandler))
}

// apihandler handles requests to the service catalog endpoint.
// When a request is received, it looks up the response from the testdata directory.
// Example:
// GET /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers responds with testdata/clusterservicebrokers.json
func apihandler(w http.ResponseWriter, r *http.Request) {
	catalogMatch := catalogRequestRegex.FindStringSubmatch(r.RequestURI)
	coreMatch := coreRequestRegex.FindStringSubmatch(r.RequestURI)

	if len(catalogMatch) == 0 && len(coreMatch) == 0 {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("unexpected request %s %s doesn't match %q or %q", r.Method, r.RequestURI, catalogRequestRegex, coreRequestRegex)))
		return
	}

	if r.Method != http.MethodGet {
		// Anything more interesting than a GET, i.e. it relies upon server behavior
		// probably should be an integration test instead
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("unallowed method for request %s %s", r.Method, r.RequestURI)))
		return
	}

	var match string
	if len(catalogMatch) > 0 {
		match = filepath.Join("catalog", catalogMatch[1])
	} else {
		match = filepath.Join("core", coreMatch[1])
	}

	relpath, err := url.PathUnescape(match)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("could not unescape path %s (%s)", match, err)))
		return
	}
	responseFile := filepath.Join("responses", relpath+".json")
	_, response, err := test.GetTestdata(responseFile)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("request %s has no matching testdata at %s (%s)", r.RequestURI, responseFile, err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func writeTestKubeconfig(fakeURL string) (string, error) {
	_, configT, err := test.GetTestdata("kubeconfig.tmpl.yaml")
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"Server": fakeURL,
	}
	t := template.Must(template.New("kubeconfig").Parse(string(configT)))

	f, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return "", errors.Wrap(err, "unable to create a temporary kubeconfig file")
	}
	defer f.Close()

	err = t.Execute(f, data)
	return f.Name(), errors.Wrap(err, "error executing the kubeconfig template")
}
