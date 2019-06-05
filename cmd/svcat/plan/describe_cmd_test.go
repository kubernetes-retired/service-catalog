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

package plan_test

import (
	"bytes"
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	. "github.com/kubernetes-sigs/service-catalog/cmd/svcat/plan"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Describe Command", func() {
	Describe("NewDescribeCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewDescribeCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("plan NAME"))
			Expect(cmd.Short).To(ContainSubstring("Show details of a specific plan"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan standard800"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan --kube-name 08e4b43a-36bc-447e-a81f-8202b13e339c"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope namespace --namespace NAMESPACE_NAME"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			kubeNameFlag := cmd.Flags().Lookup("kube-name")
			Expect(kubeNameFlag).NotTo(BeNil())
			Expect(kubeNameFlag.Usage).To(ContainSubstring("Whether or not to get the class by its Kubernetes name (the default is by external name)"))

			showSchemaFlag := cmd.Flags().Lookup("show-schemas")
			Expect(showSchemaFlag).NotTo(BeNil())
			Expect(showSchemaFlag.Usage).To(ContainSubstring("Whether or not to show instance and binding parameter schemas"))

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())
			Expect(scopeFlag.Usage).To(ContainSubstring("Limit the command to a particular scope: cluster or namespace"))

			namespaceFlag := cmd.Flags().Lookup("namespace")
			Expect(namespaceFlag).NotTo(BeNil())
			Expect(namespaceFlag.Usage).To(ContainSubstring("If present, the namespace scope for this request"))
		})
	})
	Describe("Validate", func() {
		It("errors if no argument is provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Run", func() {
		var (
			clusterServiceClass *v1beta1.ClusterServiceClass
			clusterServicePlan  *v1beta1.ClusterServicePlan
			cmd                 *DescribeCmd
			defaultNamespace    string
			defaultServiceClass *v1beta1.ServiceClass
			defaultServicePlan  *v1beta1.ServicePlan
			fakeApp             *svcat.App
			fakeSDK             *servicecatalogfakes.FakeSvcatClient
			outputBuffer        *bytes.Buffer
		)
		BeforeEach(func() {
			defaultNamespace = "default"
			fakeApp, _ = svcat.NewApp(nil, nil, "default")
			fakeSDK = new(servicecatalogfakes.FakeSvcatClient)
			fakeApp.SvcatClient = fakeSDK
			outputBuffer = &bytes.Buffer{}

			cmd = &DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}

			clusterServiceClass = &v1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "csc-123",
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: "myclusterserviclass",
					},
				},
			}
			clusterServicePlan = &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "csp-123",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: clusterServiceClass.Name,
					},
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "myclusterserviceplan",
					},
				},
			}
			defaultServiceClass = &v1beta1.ServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dsc-456",
					Namespace: defaultNamespace,
				},
				Spec: v1beta1.ServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: "mydefaultserviceclass",
					},
				},
			}
			defaultServicePlan = &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dsp-456",
					Namespace: defaultNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					ServiceClassRef: v1beta1.LocalObjectReference{
						Name: defaultServiceClass.Name,
					},
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "mydefaultserviceplan",
					},
				},
			}
		})
		It("Calls the pkg/svcat libs RetrievePlanByName with cluster scope options", func() {
			fakeSDK.RetrievePlanByNameReturns(clusterServicePlan, nil)
			fakeSDK.RetrieveClassByPlanReturns(clusterServiceClass, nil)

			cmd.Scope = servicecatalog.ClusterScope
			cmd.LookupByKubeName = false
			cmd.Name = clusterServicePlan.Spec.ExternalName
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			nameArg, scopeArg := fakeSDK.RetrievePlanByNameArgsForCall(0)
			Expect(nameArg).To(Equal(clusterServicePlan.Spec.ExternalName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope: servicecatalog.ClusterScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(clusterServicePlan.Spec.ExternalName))
			Expect(output).To(ContainSubstring(clusterServiceClass.Spec.ExternalName))
		})
		It("Calls the pkg/svcat libs RetrievePlanByName with namespace scope options", func() {
			fakeSDK.RetrievePlanByNameReturns(defaultServicePlan, nil)
			fakeSDK.RetrieveClassByPlanReturns(defaultServiceClass, nil)
			cmd := DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = defaultNamespace
			cmd.LookupByKubeName = false
			cmd.Name = defaultServicePlan.Spec.ExternalName

			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			nameArg, scopeArg := fakeSDK.RetrievePlanByNameArgsForCall(0)
			Expect(nameArg).To(Equal(defaultServicePlan.Spec.ExternalName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: defaultNamespace,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(defaultServicePlan.Spec.ExternalName))
			Expect(output).To(ContainSubstring(defaultServiceClass.Spec.ExternalName))
		})
		It("Calls the pkg/svcat libs RetrievePlanByClassAndName with cluster scope options", func() {
			fakeSDK.RetrievePlanByClassAndNameReturns(clusterServicePlan, nil)
			fakeSDK.RetrieveClassByPlanReturns(clusterServiceClass, nil)
			cmd := DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.ClusterScope
			cmd.LookupByKubeName = false
			s := []string{clusterServiceClass.Spec.ExternalName, clusterServicePlan.Spec.ExternalName}
			cmd.Name = strings.Join(s, "/")
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			classArg, nameArg, scopeArg := fakeSDK.RetrievePlanByClassAndNameArgsForCall(0)
			Expect(classArg).To(Equal(clusterServiceClass.Spec.ExternalName))
			Expect(nameArg).To(Equal(clusterServicePlan.Spec.ExternalName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope: servicecatalog.ClusterScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(clusterServicePlan.Spec.ExternalName))
			Expect(output).To(ContainSubstring(clusterServiceClass.Spec.ExternalName))
		})
		It("Calls the pkg/svcat libs RetrievePlanByClassAndName with namespace scope options", func() {
			namespaceName := "default"
			planName := "namespaceplan"
			className := "clusterclass"

			planToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespaceName,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
					},
				},
			}

			classToReturn := &v1beta1.ClusterServiceClass{
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlanByClassAndNameReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = namespaceName
			cmd.LookupByKubeName = false
			s := []string{className, planName}
			cmd.Name = strings.Join(s, "/")
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			classArg, nameArg, scopeArg := fakeSDK.RetrievePlanByClassAndNameArgsForCall(0)
			Expect(classArg).To(Equal(className))
			Expect(nameArg).To(Equal(planName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: namespaceName,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the pkg/svcat libs RetrievePlanByID with cluster scope options", func() {
			planID := "clusterplanid"
			planName := "clusterplan"
			className := "clusterclass"

			planToReturn := &v1beta1.ClusterServicePlan{
				Spec: v1beta1.ClusterServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
					},
				},
			}

			classToReturn := &v1beta1.ClusterServiceClass{
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlanByIDReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.ClusterScope
			cmd.LookupByKubeName = true
			cmd.KubeName = planID
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			idArg, scopeArg := fakeSDK.RetrievePlanByIDArgsForCall(0)
			Expect(idArg).To(Equal(planID))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope: servicecatalog.ClusterScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the pkg/svcat libs RetrievePlanByID with namespace scope options", func() {
			namespaceName := "default"
			planID := "namespaceplanid"
			planName := "namespaceplan"
			className := "clusterclass"

			planToReturn := &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespaceName,
				},
				Spec: v1beta1.ServicePlanSpec{
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
					},
				},
			}

			classToReturn := &v1beta1.ClusterServiceClass{
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrievePlanByIDReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := DescribeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = namespaceName
			cmd.LookupByKubeName = true
			cmd.KubeName = planID
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			idArg, scopeArg := fakeSDK.RetrievePlanByIDArgsForCall(0)
			Expect(idArg).To(Equal(planID))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: namespaceName,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
		})
	})
})
