/*
Copyright 2017 The Kubernetes Authors.

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

var _ = Describe("Broker", func() {
	var (
		client                *PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ClusterServiceBrokers *fakes.FakeClusterServiceBrokerInterface
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ClusterServiceBrokers = &fakes.FakeClusterServiceBrokerInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ClusterServiceBrokersReturns(ClusterServiceBrokers)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in broker", func() {
			brokerName := "foobar"
			_, err = client.GetBroker(brokerName)

			Expect(ClusterServiceBrokers.GetCallCount()).To(Equal(1))
			name, _ := ClusterServiceBrokers.GetArgsForCall(0)
			Expect(name).To(Equal(brokerName))
		})
		It("Bubbles up errors", func() {
			errorMessage := "broker not found"
			ClusterServiceBrokers.GetReturns(nil, errors.New(errorMessage))

			_, err := client.GetBroker("banana")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceBrokers.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			_, err := client.ListBrokers()

			Expect(err).NotTo(HaveOccurred())
			Expect(ClusterServiceBrokers.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			errorMessage := "foobar"
			ClusterServiceBrokers.ListReturns(nil, errors.New(errorMessage))

			_, err := client.ListBrokers()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceBrokers.ListCallCount()).To(Equal(1))
		})
	})
})
