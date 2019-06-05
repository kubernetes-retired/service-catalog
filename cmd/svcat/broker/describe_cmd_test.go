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
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Describe Command", func() {
	Describe("NewDescribeCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewDescribeCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("broker NAME"))
			Expect(cmd.Short).To(ContainSubstring("Show details of a specific broker"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe broker asb"))
			Expect(len(cmd.Aliases)).To(Equal(2))
		})
	})

	Describe("Validate", func() {
		It("succeeds if a broker name is provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{"bananabroker"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("errors if a broker name is not provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("a broker name is required"))
		})
	})
	Describe("Run", func() {
		var (
			brokerName               string
			brokerToReturn           *v1beta1.ClusterServiceBroker
			brokerURL                string
			namespace                string
			namespacedBrokerToReturn *v1beta1.ServiceBroker
		)
		BeforeEach(func() {
			brokerName = "foobarbroker"
			brokerURL = "www.foobar.com"
			namespace = "banana-namespace"

			brokerToReturn = &v1beta1.ClusterServiceBroker{
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
			namespacedBrokerToReturn = &v1beta1.ServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name:      brokerName,
					Namespace: namespace,
				},
				Spec: v1beta1.ServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL:                 brokerURL,
						CatalogRestrictions: &v1beta1.CatalogRestrictions{},
					},
				},
			}
		})
		It("Calls the pkg/svcat libs RetrieveBrokerByID method with the passed in variables and prints output to the user", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokerByIDReturns(brokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       brokerName,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
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
			Expect(output).To(ContainSubstring("Scope:    cluster"))
		})
		It("prints out a namespaced broker when it only finds a namespace broker", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokerByIDReturns(namespacedBrokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       brokerName,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
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
			Expect(output).To(ContainSubstring("Scope:       namespace "))
		})
		It("bubbles up errors", func() {
			outputBuffer := &bytes.Buffer{}
			errMsg := "incompatible potato"
			errToReturn := fmt.Errorf(errMsg)

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokerByIDReturns(nil, errToReturn)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       brokerName,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
			err := cmd.Run()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsg))

			Expect(fakeSDK.RetrieveBrokerByIDCallCount()).To(Equal(1))
			returnedName, returnedScopeOpts := fakeSDK.RetrieveBrokerByIDArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))
			scopeOpts := servicecatalog.ScopeOptions{
				Scope:     servicecatalog.AllScope,
				Namespace: namespace,
			}
			Expect(returnedScopeOpts).To(Equal(scopeOpts))
		})
		It("prompts the user for more input when it gets a MultipleBrokersFound error", func() {
			outputBuffer := &bytes.Buffer{}
			errToReturn := fmt.Errorf(servicecatalog.MultipleBrokersFoundError + " for '" + brokerName + "'")

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RetrieveBrokerByIDReturns(nil, errToReturn)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := DescribeCmd{
				Context:    cxt,
				Namespaced: command.NewNamespaced(cxt),
				Name:       brokerName,
				Scoped:     command.NewScoped(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.AllScope
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
