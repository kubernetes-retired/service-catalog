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

package broker_test

import (
	"bytes"
	"fmt"

	. "github.com/kubernetes-sigs/service-catalog/cmd/svcat/broker"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Get Broker Command", func() {
	Describe("NewGetBrokerCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewGetCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("brokers [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("List brokers, optionally filtered by name, scope or namespace"))
			Expect(cmd.Example).To(ContainSubstring("svcat get brokers"))
			Expect(cmd.Example).To(ContainSubstring("svcat get brokers --scope=cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat get brokers --scope=all"))
			Expect(cmd.Example).To(ContainSubstring("svcat get broker minibroker"))
			Expect(len(cmd.Aliases)).To(Equal(2))
		})
	})
	Describe("Validate", func() {
		It("allows broker name arg to be empty", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(BeNil())
		})
		It("optionally parses the broker name argument", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{"minibroker"})
			Expect(err).To(BeNil())
			Expect(cmd.Name).To(Equal("minibroker"))
		})
	})
	Describe("Run", func() {
		It("Calls the pkg/svcat libs RetrieveBrokers with namespace scope and current namespace", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokersReturns(
				[]servicecatalog.Broker{&v1beta1.ServiceBroker{ObjectMeta: v1.ObjectMeta{Name: "minibroker", Namespace: "default"}}},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := GetCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Namespace = "default"
			cmd.Scope = servicecatalog.NamespaceScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveBrokersArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "default",
				Scope:     servicecatalog.NamespaceScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("minibroker"))
		})
		It("Calls the pkg/svcat libs RetrieveBrokers with namespace scope and all namespaces", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokersReturns(
				[]servicecatalog.Broker{
					&v1beta1.ServiceBroker{ObjectMeta: v1.ObjectMeta{Name: "minibroker", Namespace: "default"}},
					&v1beta1.ServiceBroker{ObjectMeta: v1.ObjectMeta{Name: "ups-broker", Namespace: "test-ns"}},
				},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := GetCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Namespace = ""
			cmd.Scope = servicecatalog.NamespaceScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveBrokersArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "",
				Scope:     servicecatalog.NamespaceScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("minibroker"))
			Expect(output).To(ContainSubstring("ups-broker"))
		})
		It("Calls the pkg/svcat libs RetrieveBrokers with all scope and current namespaces", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokersReturns(
				[]servicecatalog.Broker{
					&v1beta1.ClusterServiceBroker{ObjectMeta: v1.ObjectMeta{Name: "global-broker"}},
					&v1beta1.ServiceBroker{ObjectMeta: v1.ObjectMeta{Name: "minibroker", Namespace: "default"}},
				},
				nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := GetCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Formatted:  command.NewFormatted(),
			}
			cmd.Namespace = "default"
			cmd.Scope = servicecatalog.AllScope

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			scopeArg := fakeSDK.RetrieveBrokersArgsForCall(0)
			Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
				Namespace: "default",
				Scope:     servicecatalog.AllScope,
			}))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("global-broker"))
			Expect(output).To(ContainSubstring("minibroker"))
		})
		Context("getting a single broker", func() {
			var (
				brokerName string
				brokerURL  string
				csb        *v1beta1.ClusterServiceBroker
				namespace  string
			)
			BeforeEach(func() {
				brokerName = "global-broker"
				brokerURL = "www.foobar.com"
				namespace = "default"
				csb = &v1beta1.ClusterServiceBroker{
					ObjectMeta: v1.ObjectMeta{
						Name: brokerName,
					},
					Spec: v1beta1.ClusterServiceBrokerSpec{
						CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
							URL:                 brokerURL,
							CatalogRestrictions: &v1beta1.CatalogRestrictions{},
						},
					},
				}
			})
			It("Calls the pkg/svcat libs RetrieveBrokerByID when getting a single broker", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, "default")
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveBrokerByIDReturns(csb, nil)
				fakeApp.SvcatClient = fakeSDK
				cmd := GetCmd{
					Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
					Scoped:     command.NewScoped(),
					Formatted:  command.NewFormatted(),
				}
				cmd.Namespace = "default"
				cmd.Scope = servicecatalog.AllScope
				cmd.Name = brokerName
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveBrokerByIDCallCount()).To(Equal(1))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveBrokerByIDArgsForCall(0)
				Expect(returnedName).To(Equal(brokerName))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(brokerName))
				Expect(output).To(ContainSubstring(brokerURL))
			})
			It("prompts the user for more input when it finds multiple brokers", func() {
				outputBuffer := &bytes.Buffer{}

				fakeApp, _ := svcat.NewApp(nil, nil, "default")
				fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
				fakeSDK.RetrieveBrokerByIDReturns(nil, fmt.Errorf(servicecatalog.MultipleBrokersFoundError+" for broker '"+brokerName+"'"))
				fakeApp.SvcatClient = fakeSDK
				cmd := GetCmd{
					Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
					Scoped:     command.NewScoped(),
					Formatted:  command.NewFormatted(),
				}
				cmd.Namespace = "default"
				cmd.Scope = servicecatalog.AllScope
				cmd.Name = brokerName
				err := cmd.Run()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(servicecatalog.MultipleBrokersFoundError))
				Expect(err.Error()).To(ContainSubstring("specify a scope with --scope"))

				Expect(fakeSDK.RetrieveBrokerByIDCallCount()).To(Equal(1))
				returnedName, returnedScopeOpts := fakeSDK.RetrieveBrokerByIDArgsForCall(0)
				Expect(returnedName).To(Equal(brokerName))
				scopeOpts := servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: namespace,
				}
				Expect(returnedScopeOpts).To(Equal(scopeOpts))

			})
		})
	})
})
