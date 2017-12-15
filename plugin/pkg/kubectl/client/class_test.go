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

var _ = Describe("Class", func() {
	var (
		client                *PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ClusterServiceClasses *fakes.FakeClusterServiceClassInterface
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ClusterServiceClasses = &fakes.FakeClusterServiceClassInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ClusterServiceClassesReturns(ClusterServiceClasses)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in class", func() {
			className := "foobar"
			_, err = client.GetClass(className)

			Expect(ClusterServiceClasses.GetCallCount()).To(Equal(1))
			name, _ := ClusterServiceClasses.GetArgsForCall(0)
			Expect(name).To(Equal(className))
		})
		It("Bubbles up errors", func() {
			errorMessage := "class not found"
			ClusterServiceClasses.GetReturns(nil, errors.New(errorMessage))

			_, err := client.GetClass("banana")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceClasses.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			_, err := client.ListClasses()

			Expect(err).NotTo(HaveOccurred())
			Expect(ClusterServiceClasses.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			errorMessage := "foobar"
			ClusterServiceClasses.ListReturns(nil, errors.New(errorMessage))

			_, err := client.ListClasses()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceClasses.ListCallCount()).To(Equal(1))
		})
	})
})
