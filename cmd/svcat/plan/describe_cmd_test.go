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
	"strings"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
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
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan --uuid 08e4b43a-36bc-447e-a81f-8202b13e339c"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope namespace --namespace NAMESPACE_NAME"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			uuidFlag := cmd.Flags().Lookup("uuid")
			Expect(uuidFlag).NotTo(BeNil())
			Expect(uuidFlag.Usage).To(ContainSubstring("Whether or not to get the class by UUID (the default is by name)"))

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
			cmd := describeCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Run", func() {
		It("Calls the pkg/svcat libs RetrievePlanByName with cluster scope options", func() {
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
			fakeSDK.RetrievePlanByNameReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.ClusterScope
			cmd.lookupByUUID = false
			cmd.name = planName
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			nameArg, scopeArg := fakeSDK.RetrievePlanByNameArgsForCall(0)
			Expect(nameArg).To(Equal(planName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope: servicecatalog.ClusterScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the pkg/svcat libs RetrievePlanByName with namespace scope options", func() {
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
			fakeSDK.RetrievePlanByNameReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = namespaceName
			cmd.lookupByUUID = false
			cmd.name = planName
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			nameArg, scopeArg := fakeSDK.RetrievePlanByNameArgsForCall(0)
			Expect(nameArg).To(Equal(planName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope:     servicecatalog.NamespaceScope,
				Namespace: namespaceName,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the pkg/svcat libs RetrievePlanByClassAndName with cluster scope options", func() {
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
			fakeSDK.RetrievePlanByClassAndNameReturns(planToReturn, nil)
			fakeSDK.RetrieveClassByPlanReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.ClusterScope
			cmd.lookupByUUID = false
			s := []string{className, planName}
			cmd.name = strings.Join(s, "/")
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			classArg, nameArg, scopeArg := fakeSDK.RetrievePlanByClassAndNameArgsForCall(0)
			Expect(classArg).To(Equal(className))
			Expect(nameArg).To(Equal(planName))
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Scope: servicecatalog.ClusterScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(className))
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
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = namespaceName
			cmd.lookupByUUID = false
			s := []string{className, planName}
			cmd.name = strings.Join(s, "/")
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
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.ClusterScope
			cmd.lookupByUUID = true
			cmd.uuid = planID
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
			cmd := describeCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = namespaceName
			cmd.lookupByUUID = true
			cmd.uuid = planID
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
