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
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		client       *PluginClient
		err          error
		svcCatClient *fake.Clientset
		sb           *v1beta1.ClusterServiceBroker
		sb2          *v1beta1.ClusterServiceBroker
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())

		sb = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "foobar"}}
		sb2 = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "barbaz"}}
		svcCatClient = fake.NewSimpleClientset(sb, sb2)
		client.ScClient = svcCatClient
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in broker", func() {
			brokerName := "foobar"

			broker, err := client.GetBroker(brokerName)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker.Name).To(Equal(brokerName))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(brokerName))
		})
		It("Bubbles up errors", func() {
			brokerName := "banana"

			broker, err := client.GetBroker(brokerName)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(brokerName))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			brokers, err := client.ListBrokers()

			Expect(err).NotTo(HaveOccurred())
			Expect(brokers.Items).Should(ConsistOf(*sb, *sb2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			client.ScClient = badClient
			_, err := client.ListBrokers()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
		})
	})
})
