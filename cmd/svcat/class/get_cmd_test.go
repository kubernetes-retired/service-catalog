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
	"errors"
	"strings"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestListClasses(t *testing.T) {
	const ns = "default"
	testcases := []struct {
		name                  string
		scope                 servicecatalog.Scope
		fakeClusterClasses    []string
		fakeNamespacedClasses []string
		wantResults           int
		wantOutput            string
		wantError             bool
	}{
		{
			name:                  "get classes from cluster and current namespace",
			scope:                 servicecatalog.AllScope,
			fakeClusterClasses:    []string{"my-cluster-class"},
			fakeNamespacedClasses: []string{"my-ns-class"},
			wantResults:           2,
			wantOutput:            "my-ns-class\nmy-cluster-class",
			wantError:             false,
		},
		{
			name:                  "get classes from cluster only",
			scope:                 servicecatalog.ClusterScope,
			fakeClusterClasses:    []string{"my-cluster-class"},
			fakeNamespacedClasses: []string{"my-ns-class"},
			wantResults:           1,
			wantOutput:            "my-cluster-class",
			wantError:             false,
		},
		{
			name:                  "get classes current namespace only",
			scope:                 servicecatalog.NamespaceScope,
			fakeClusterClasses:    []string{"my-cluster-class"},
			fakeNamespacedClasses: []string{"my-ns-class"},
			wantResults:           1,
			wantOutput:            "my-ns-class",
			wantError:             false,
		},
		{
			name:                  "get classes - bubbles cluster errors",
			scope:                 servicecatalog.AllScope,
			fakeClusterClasses:    []string{"badclass"},
			fakeNamespacedClasses: []string{"my-ns-class"},
			wantOutput:            "unable to list cluster-scoped classes (sabotaged)",
			wantError:             true,
		},
		{
			name:                  "get classes - bubbles namespace errors",
			scope:                 servicecatalog.AllScope,
			fakeClusterClasses:    []string{"my-cluster-class"},
			fakeNamespacedClasses: []string{"badclass"},
			wantOutput:            "unable to list classes in \"default\" (sabotaged)",
			wantError:             true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			var fakes []runtime.Object
			for _, name := range tc.fakeClusterClasses {
				fakes = append(fakes, &v1beta1.ClusterServiceClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1beta1.ClusterServiceClassSpec{
						CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
							ExternalName: name,
						},
					},
				})
			}
			for _, name := range tc.fakeNamespacedClasses {
				fakes = append(fakes, &v1beta1.ServiceClass{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
					},
					Spec: v1beta1.ServiceClassSpec{
						CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
							ExternalName: name,
						},
					},
				})
			}
			svcatClient := svcatfake.NewSimpleClientset(fakes...)
			output := &bytes.Buffer{}
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, ns)
			cxt := svcattest.NewContext(output, fakeApp)

			// Sabotage the get calls, if necessary
			for _, name := range tc.fakeClusterClasses {
				if strings.Contains(name, "bad") {
					svcatClient.PrependReactor("list", "clusterserviceclasses",
						func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, errors.New("sabotaged")
						})
					break
				}
			}
			for _, name := range tc.fakeNamespacedClasses {
				if strings.Contains(name, "bad") {
					svcatClient.PrependReactor("list", "serviceclasses",
						func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, errors.New("sabotaged")
						})
					break
				}
			}

			// Initialize the command arguments
			cmd := &getCmd{
				Namespaced:   command.NewNamespaced(cxt),
				Scoped:       command.NewScoped(),
				outputFormat: "table",
			}
			cmd.Namespace = ns
			cmd.Scope = tc.scope

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
			if !svcattest.OutputMatches(gotOutput, tc.wantOutput, true) {
				t.Errorf("unexpected output \n\nWANT:\n%q\n\nGOT:\n%q\n", tc.wantOutput, gotOutput)
			}
		})
	}
}

var _ = Describe("Get Classes Command", func() {
	Describe("NewGetClassesCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewGetCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("classes [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("List classes, optionally filtered by name, scope or namespace"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes --scope namespace --namespace dev"))
			Expect(len(cmd.Aliases)).To(Equal(2))
		})
	})
	Describe("Validate", func() {
		It("allows class name arg to be empty", func() {
			cmd := &getCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(BeNil())
		})
		It("optionally parses the class name argument", func() {
			cmd := &getCmd{}
			err := cmd.Validate([]string{"mysqldb"})
			Expect(err).To(BeNil())
			Expect(cmd.name).To(Equal("mysqldb"))
		})
	})
	Describe("Run", func() {
		It("Calls the pkg/svcat libs RetrieveClasses with namespace scope and current namespace", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassesReturns(
				[]servicecatalog.Class{&v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "mysqldb", Namespace: "default"}}},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Namespace = "default"
			cmd.Scope = servicecatalog.NamespaceScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "default",
				Scope:     servicecatalog.NamespaceScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("mysqldb"))
		})
		It("Calls the pkg/svcat libs RetrieveClasses with namespace scope and all namespaces", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassesReturns(
				[]servicecatalog.Class{
					&v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "mysqldb", Namespace: "default"}},
					&v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "postgresdb", Namespace: "test-ns"}},
				},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Namespace = ""
			cmd.Scope = servicecatalog.NamespaceScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "",
				Scope:     servicecatalog.NamespaceScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("mysqldb"))
			Expect(output).To(ContainSubstring("postgresdb"))
		})
		It("Calls the pkg/svcat libs RetrieveClasses with all scope and current namespaces", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassesReturns(
				[]servicecatalog.Class{
					&v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "mysqldb"}},
					&v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "postgresdb", Namespace: "default"}},
				},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Namespace = "default"
			cmd.Scope = servicecatalog.AllScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "default",
				Scope:     servicecatalog.AllScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("mysqldb"))
			Expect(output).To(ContainSubstring("postgresdb"))
		})
	})
})
