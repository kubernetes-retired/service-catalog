/*
Copyright 2019 The Kubernetes Authors.

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

package class_test

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kubernetes-sigs/service-catalog/cmd/svcat/class"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = Describe("Describe Command", func() {
	Describe("NewDescribeCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewDescribeCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("class NAME"))
			Expect(cmd.Short).To(ContainSubstring("Show details of a specific class"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe class mysqldb"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe class --kube-name 997b8372-8dac-40ac-ae65-758b4a5075a5"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			kubeNameFlag := cmd.Flags().Lookup("kube-name")
			Expect(kubeNameFlag).NotTo(BeNil())
			Expect(kubeNameFlag.Usage).To(ContainSubstring("Whether or not to get the class by its Kubernetes name (the default is by external name)"))

			namespaceFlag := cmd.Flags().Lookup("namespace")
			Expect(namespaceFlag).NotTo(BeNil())

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())
		})
	})

	Describe("Validate", func() {
		It("succeeds if a class name is provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{"bananaclass"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("errors if a class name is not provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("a class external name or Kubernetes name is required"))
		})
	})
	Describe("Run", func() {
		var (
			brokerName              string
			classKubeName           string
			className               string
			classToReturn           *v1beta1.ClusterServiceClass
			namespace               string
			namespacedClassKubeName string
			namespacedClassName     string
			namespacedClassToReturn *v1beta1.ServiceClass
			planKubeName            string
			planName                string
			planToReturn            *v1beta1.ClusterServicePlan
		)
		BeforeEach(func() {
			brokerName = "mysql-broker"
			classKubeName = "abc123"
			className = "mysql"
			namespacedClassKubeName = "xyz-098"
			namespacedClassName = "namespaced-mysql"
			namespace = "potato"
			planKubeName = "acfhbvc-12345"
			planName = "10mb-mysql"

			classToReturn = &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: classKubeName,
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: brokerName,
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
						ExternalID:   "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
						Description:  "A mysql service",
					},
				},
			}
			planToReturn = &v1beta1.ClusterServicePlan{
				ObjectMeta: v1.ObjectMeta{
					Name: planKubeName,
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceBrokerName: brokerName,
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: planName,
						ExternalID:   "khabskdjasbdja",
						Description:  "A 10 MB mysql database",
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: classKubeName,
					},
				},
			}
			namespacedClassToReturn = &v1beta1.ServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name:      namespacedClassKubeName,
					Namespace: namespace,
				},
				Spec: v1beta1.ServiceClassSpec{
					ServiceBrokerName: brokerName,
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: namespacedClassName,
						ExternalID:   "qwerty-12345",
						Description:  "A namespaced mysql service",
					},
				},
			}
		})
		It("Calls the pkg/svcat libs RetrieveClassByName method with the passed in variables, and then calls the generated RetrievePlans with the returned class's kube name and prints output to the user", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByNameReturns(classToReturn, nil)
			fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{planToReturn}, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       className,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(1))
			returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByNameArgsForCall(0)
			Expect(returnedName).To(Equal(className))
			scopeOpts := servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}
			Expect(returnedScopeOpts).To(Equal(scopeOpts))

			Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(1))
			returnedName, returnedScopeOpts = fakeSDK.RetrievePlansArgsForCall(0)
			Expect(returnedName).To(Equal(classKubeName))
			scopeOpts = servicecatalog.ScopeOptions{
				Scope: servicecatalog.AllScope,
			}
			Expect(returnedScopeOpts).To(Equal(scopeOpts))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring("Scope:             cluster"))
			Expect(output).To(ContainSubstring(planName))
			Expect(output).To(ContainSubstring(planToReturn.Spec.Description))
		})
		It("prints out a namespaced class when it only finds a namespace class ", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByNameReturns(namespacedClassToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       namespacedClassName,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(1))
			returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByNameArgsForCall(0)
			Expect(returnedName).To(Equal(namespacedClassName))
			scopeOpts := servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}
			Expect(returnedScopeOpts).To(Equal(scopeOpts))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(namespacedClassName))
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring("Scope:             namespace"))
		})
		It("Calls the pkg/svcat libs RetrieveClassByID method when --kube-name is thrown", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByIDReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:          cxt,
				Namespaced:       command.NewNamespaced(cxt),
				KubeName:         classKubeName,
				LookupByKubeName: true,
				Scoped:           command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
			returnedID, returnedScopeOpts := fakeSDK.RetrieveClassByIDArgsForCall(0)
			Expect(returnedID).To(Equal(classKubeName))
			scopeOpts := servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}
			Expect(returnedScopeOpts).To(Equal(scopeOpts))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring("Scope:             cluster"))
		})
		It("bubbles up errors", func() {
			errMsg := "banana error"
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByNameReturns(nil, fmt.Errorf(errMsg))
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       className,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})
		It("prompts the user for more input when it gets a MultipleClassesFound error", func() {
			errToReturn := fmt.Errorf(servicecatalog.MultipleClassesFoundError + " for '" + className + "'")
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByNameReturns(nil, errToReturn)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       className,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("specify a scope with --scope or an exact Kubernetes name with --kube-name"))
		})
		It("bubbles up errors from RetrievePlans", func() {
			errMsg := "plan error"
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveClassByNameReturns(classToReturn, nil)
			fakeSDK.RetrievePlansReturns(nil, fmt.Errorf(errMsg))
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       className,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))
		})
	})
})
