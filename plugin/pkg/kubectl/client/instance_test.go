/*
Copyright 2016 The Kubernetes Authors.

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

package client_test

import (
	"errors"

	. "github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instances", func() {
	var (
		client                *PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ServiceInstances      *fakes.FakeServiceInstanceInterface
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ServiceInstances = &fakes.FakeServiceInstanceInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ServiceInstancesReturns(ServiceInstances)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in instance and namespace", func() {
			namespace := "foobar_namespace"
			instanceName := "potato_instance"

			_, err := client.GetInstance(instanceName, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			returnedName, _ := ServiceInstances.GetArgsForCall(0)
			Expect(returnedName).To(Equal(instanceName))
			Expect(ServiceInstances.GetCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			namespace := "foobar_namespace"
			instanceName := "potato_instance"
			errorMessage := "instance not found"

			ServiceInstances.GetReturns(nil, errors.New(errorMessage))
			_, err := client.GetInstance(instanceName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(ServiceInstances.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			namespace := "foobar_namespace"

			_, err := client.ListInstances(namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceInstances.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			ServiceInstances.ListReturns(nil, errors.New("foobar"))
			namespace := "foobar_namespace"

			_, err := client.ListInstances(namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("foobar"))
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceInstances.ListCallCount()).To(Equal(1))
		})
	})
})
