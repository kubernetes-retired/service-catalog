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

package plan

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
	servicecatalogfakes "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
)

func TestListPlans(t *testing.T) {
	const ns = "default"
	testcases := []struct {
		name                string
		scope               servicecatalog.Scope
		fakeClusterPlans    []string
		fakeNamespacedPlans []string
		wantResults         int
		wantOutput          string
		wantError           bool
	}{
		{
			name:                "get plans from cluster and current namespace",
			scope:               servicecatalog.AllScope,
			fakeClusterPlans:    []string{"my-cluster-plan"},
			fakeNamespacedPlans: []string{"my-ns-plan"},
			wantResults:         2,
			wantOutput:          "my-ns-plan\nmy-cluster-plan",
			wantError:           false,
		},
		{
			name:                "get plans from cluster only",
			scope:               servicecatalog.ClusterScope,
			fakeClusterPlans:    []string{"my-cluster-plan"},
			fakeNamespacedPlans: []string{"my-ns-plan"},
			wantResults:         1,
			wantOutput:          "my-cluster-plan",
			wantError:           false,
		},
		{
			name:                "get plans current namespace only",
			scope:               servicecatalog.NamespaceScope,
			fakeClusterPlans:    []string{"my-cluster-plan"},
			fakeNamespacedPlans: []string{"my-ns-plan"},
			wantResults:         1,
			wantOutput:          "my-ns-plan",
			wantError:           false,
		},
		{
			name:                "get plans - bubbles cluster errors",
			scope:               servicecatalog.AllScope,
			fakeClusterPlans:    []string{"badplan"},
			fakeNamespacedPlans: []string{"my-ns-plan"},
			wantOutput:          "unable to list cluster-scoped plans (sabotaged)",
			wantError:           true,
		},
		{
			name:                "get plans - bubbles namespace errors",
			scope:               servicecatalog.AllScope,
			fakeClusterPlans:    []string{"my-cluster-plan"},
			fakeNamespacedPlans: []string{"badplan"},
			wantOutput:          "unable to list plans in \"default\" (sabotaged)",
			wantError:           true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			var fakes []runtime.Object
			for _, name := range tc.fakeClusterPlans {
				fakes = append(fakes, &v1beta1.ClusterServicePlan{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1beta1.ClusterServicePlanSpec{
						CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
							ExternalName: name,
						},
					},
				})
			}
			for _, name := range tc.fakeNamespacedPlans {
				fakes = append(fakes, &v1beta1.ServicePlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns,
					},
					Spec: v1beta1.ServicePlanSpec{
						CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
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
			for _, name := range tc.fakeClusterPlans {
				if strings.Contains(name, "bad") {
					svcatClient.PrependReactor("list", "clusterserviceplans",
						func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, errors.New("sabotaged")
						})
					break
				}
			}
			for _, name := range tc.fakeNamespacedPlans {
				if strings.Contains(name, "bad") {
					svcatClient.PrependReactor("list", "serviceplans",
						func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, errors.New("sabotaged")
						})
					break
				}
			}

			// Initialize the command arguments
			cmd := &getCmd{
				Namespaced: command.NewNamespaced(cxt),
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
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

var _ = Describe("Get Plans Command", func() {
	Describe("NewGetPlansCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewGetCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("plans [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("List plans, optionally filtered by name, class, scope or namespace"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans --scope namespace --namespace dev"))
			Expect(len(cmd.Aliases)).To(Equal(2))
		})
	})
	Describe("Validate", func() {
		It("allows plan name arg to be empty", func() {
			cmd := &getCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(BeNil())
		})
		It("optionally parses the plan name argument", func() {
			cmd := &getCmd{}
			err := cmd.Validate([]string{"myplan"})
			Expect(err).To(BeNil())
			Expect(cmd.name).To(Equal("myplan"))
		})
	})
	Describe("Run", func() {
		It("Calls the pkg/svcat libs RetrievePlans with namespace scope and current namespace", func() {
			planName := "myplan"
			planNamespace := "default"

			planToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: planNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{planToReturn}, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = planNamespace
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			_, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: planNamespace,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
		})
		It("Calls the pkg/svcat libs RetrievePlans with namespace scope and all namespaces", func() {
			planOneName := "myplan"
			planOneNamespace := "default"

			planTwoName := "anotherplan"
			planTwoNamespace := "test-ns"

			planOneToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: planOneNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planOneName,
					},
				},
			}

			planTwoToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: planTwoNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planTwoName,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{planOneToReturn, planTwoToReturn}, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = ""
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			_, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: "",
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planOneName))
			Expect(output).To(ContainSubstring(planTwoName))
		})
		It("Calls the pkg/svcat libs RetrievePlans with all scope and current namespaces", func() {
			planOneName := "myplan"

			planTwoName := "anotherplan"
			planTwoNamespace := "default"

			planOneToReturn := &v1beta1.ClusterServicePlan{
				Spec: v1beta1.ClusterServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planOneName,
					},
				},
			}

			planTwoToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: planTwoNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planTwoName,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{planOneToReturn, planTwoToReturn}, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := getCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Scope = servicecatalog.AllScope
			cmd.Namespace = planTwoNamespace
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			_, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: planTwoNamespace,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planOneName))
			Expect(output).To(ContainSubstring(planTwoName))
		})
	})
})
