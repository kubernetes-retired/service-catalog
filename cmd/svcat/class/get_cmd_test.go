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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Class Command", func() {
	Describe("NewGetClassCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewGetCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("classes [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("List classes, optionally filtered by name, scope or namespace"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat get classes --scope namespace --namespace dev"))
			Expect(cmd.Example).To(ContainSubstring("svcat get class mysql"))
			Expect(cmd.Example).To(ContainSubstring("svcat get class --kube-name 997b8372-8dac-40ac-ae65-758b4a5075a5"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			kubeNameFlag := cmd.Flags().Lookup("kube-name")
			Expect(kubeNameFlag).NotTo(BeNil())
			Expect(kubeNameFlag.Usage).To(ContainSubstring("Whether or not to get the class by its Kubernetes name (the default is by external name)"))

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())
		})
	})
	Describe("Validate", func() {
		It("allows class name arg to be empty", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(BeNil())
		})
		It("optionally parses the class name argument", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{"foobarclass"})
			Expect(err).To(BeNil())
			Expect(cmd.Name).To(Equal("foobarclass"))

			cmd.LookupByKubeName = true
			err = cmd.Validate([]string{"foobarclass"})
			Expect(err).To(BeNil())
			Expect(cmd.KubeName).To(Equal("foobarclass"))
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
		)
		BeforeEach(func() {
			brokerName = "mysql-broker"
			classKubeName = "abc123"
			className = "cluster-mysql"
			namespacedClassKubeName = "xyz-098"
			namespacedClassName = "namespaced-mysql"
			namespace = "potato"

			classToReturn = &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: classKubeName,
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: brokerName,
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
						ExternalID:   "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
						Description:  "A cluster mysql service",
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
		Context("getting all classes", func() {
			It("Calls the pkg/svcat libs RetrieveClasses with all scope and current namespace", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{classToReturn, namespacedClassToReturn}, nil)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.AllScope
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				returnedScopeOpts := fakeSDK.RetrieveClassesArgsForCall(0)
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(className))
				Expect(output).To(ContainSubstring(classToReturn.Spec.Description))
				Expect(output).To(ContainSubstring(namespacedClassName))
				Expect(output).To(ContainSubstring(namespace))
				Expect(output).To(ContainSubstring(namespacedClassToReturn.Spec.Description))
			})
			It("Calls the pkg/svcat libs RetrieveClasses  with namespace scope and current namespace", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{namespacedClassToReturn}, nil)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.NamespaceScope
				cmd.Namespace = namespace
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				returnedScopeOpts := fakeSDK.RetrieveClassesArgsForCall(0)
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.NamespaceScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).NotTo(ContainSubstring(className))
				Expect(output).NotTo(ContainSubstring(classToReturn.Spec.Description))
				Expect(output).To(ContainSubstring(namespacedClassName))
				Expect(output).To(ContainSubstring(namespace))
				Expect(output).To(ContainSubstring(namespacedClassToReturn.Spec.Description))
			})
			It("Calls the pkg/svcat libs RetrieveClasses with cluster scope", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{classToReturn}, nil)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.ClusterScope
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				returnedScopeOpts := fakeSDK.RetrieveClassesArgsForCall(0)
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.ClusterScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(className))
				Expect(output).To(ContainSubstring(classToReturn.Spec.Description))
				Expect(output).NotTo(ContainSubstring(namespacedClassName))
				Expect(output).NotTo(ContainSubstring(namespace))
				Expect(output).NotTo(ContainSubstring(namespacedClassToReturn.Spec.Description))
			})
		})
		Context("getting a single class", func() {
			It("Calls the pkg/svcat libs RetrieveClassByName when getting a single class", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassByNameReturns(classToReturn, nil)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.AllScope
				cmd.Name = className
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(0))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByNameArgsForCall(0)
				Expect(returnedName).To(Equal(className))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(className))
				Expect(output).To(ContainSubstring(classToReturn.Spec.Description))
			})
			It("Calls the pkg/svcat libs RetrieveClassByID when --kube-name is thrown", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassByIDReturns(classToReturn, nil)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.AllScope
				cmd.Namespace = namespace
				cmd.KubeName = classKubeName
				cmd.LookupByKubeName = true
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(0))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByIDArgsForCall(0)
				Expect(returnedName).To(Equal(classKubeName))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(className))
				Expect(output).To(ContainSubstring(classToReturn.Spec.Description))
			})
			It("bubbles up errors", func() {
				errMsg := "kaboom"
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassByIDReturns(nil, fmt.Errorf(errMsg))
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.AllScope
				cmd.Namespace = namespace
				cmd.KubeName = classKubeName
				cmd.LookupByKubeName = true
				err := cmd.Run()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(0))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByIDArgsForCall(0)
				Expect(returnedName).To(Equal(classKubeName))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))
			})
			It("prompts the user for more input when it finds multiple classes", func() {
				errToReturn := fmt.Errorf(servicecatalog.MultipleClassesFoundError + " for '" + className + "'")
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, namespace)
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveClassByIDReturns(nil, errToReturn)
				fakeApp.SvcatClient = fakeSDK
				cxt := svcattest.NewContext(outputBuffer, fakeApp)
				cmd := GetCmd{
					Formatted:  command.NewFormatted(),
					Namespaced: command.NewNamespaced(cxt),
					Scoped:     command.NewScoped(),
				}
				cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
				cmd.Scope = servicecatalog.AllScope
				cmd.Namespace = namespace
				cmd.KubeName = classKubeName
				cmd.LookupByKubeName = true
				err := cmd.Run()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("please specify a scope with --scope or an exact Kubernetes name with --kube-name"))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(0))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveClassByIDArgsForCall(0)
				Expect(returnedName).To(Equal(classKubeName))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))
			})
		})
	})
})
