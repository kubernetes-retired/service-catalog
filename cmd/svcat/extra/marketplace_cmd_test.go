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

package extra_test

import (
	"bytes"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	. "github.com/kubernetes-incubator/service-catalog/cmd/svcat/extra"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	_ "github.com/kubernetes-incubator/service-catalog/internal/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Register Command", func() {
	Describe("NewMarketplaceCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewMarketplaceCmd(cxt)
			Expect(*cmd).NotTo(BeNil())

			Expect(cmd.Use).To(Equal("marketplace"))
			Expect(cmd.Short).To(ContainSubstring("List available service offerings"))
			Expect(cmd.Example).To(ContainSubstring("svcat marketplace --namespace dev"))
			Expect(cmd.Aliases).To(ConsistOf("marketplace", "mp"))

			urlFlag := cmd.Flags().Lookup("namespace")
			Expect(urlFlag).NotTo(BeNil())
			Expect(urlFlag.Usage).To(ContainSubstring("If present, the namespace scope for this request"))
		})
	})
	Describe("Validate", func() {
	})
	Describe("Run", func() {
		It("Calls the pkg/svcat libs methods to retrieve all classes and plans and prints output to the user", func() {
			namespace := "banana"

			className := "foobarclass"
			classDescription := "This class foobars"
			className2 := "barbazclass"
			classDescription2 := "This class barbazs"
			planName := "foobarplan1"
			planName2 := "foobarplan2"
			planName3 := "barbazplan"
			classToReturn := &v1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      className,
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						Description:  classDescription,
						ExternalName: className,
					},
				},
			}
			classToReturn2 := &v1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      className2,
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						Description:  classDescription2,
						ExternalName: className2,
					},
				},
			}
			planToReturn := &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planName,
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: className,
					},
				},
			}
			planToReturn2 := &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planName2,
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName2,
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: className,
					},
				},
			}
			planToReturn3 := &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      planName3,
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName3,
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: className2,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}
			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{classToReturn, classToReturn2}, nil)
			fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{planToReturn, planToReturn2, planToReturn3}, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := MarketplaceCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Formatted:  command.NewFormatted(),
			}
			cmd.Namespace = namespace

			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
			scopeOpts := fakeSDK.RetrieveClassesArgsForCall(0)
			Expect(scopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}))

			Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(1))
			class, scopeOpts := fakeSDK.RetrievePlansArgsForCall(0)
			Expect(class).To(Equal(""))
			Expect(scopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(planName2))
			Expect(output).To(ContainSubstring(classDescription))
			Expect(output).To(ContainSubstring(className2))
			Expect(output).To(ContainSubstring(planName3))
			Expect(output).To(ContainSubstring(classDescription2))
		})
	})
})
