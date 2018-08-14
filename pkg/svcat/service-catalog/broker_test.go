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

package servicecatalog_test

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		sdk          *SDK
		svcCatClient *fake.Clientset
		csb          *v1beta1.ClusterServiceBroker
		csb2         *v1beta1.ClusterServiceBroker
		sb           *v1beta1.ServiceBroker
		sb2          *v1beta1.ServiceBroker
	)

	BeforeEach(func() {
		csb = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "foobar"}}
		csb2 = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "barbaz"}}
		sb = &v1beta1.ServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "default"}}
		sb2 = &v1beta1.ServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "ns2"}}
		svcCatClient = fake.NewSimpleClientset(csb, csb2, sb, sb2)
		sdk = &SDK{
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("Deregister", func() {
		It("deletes a broker by calling the v1beta1 Delete method with the passed in arguement", func() {
			brokerName := "foobar"

			err := sdk.Deregister(brokerName)

			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("delete", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.DeleteActionImpl).Name).To(Equal(brokerName))
		})
		It("Bubbles up errors", func() {
			errorMessage := "error deregistering broker"
			brokerName := "potato_broker"
			badClient := &fake.Clientset{}
			badClient.AddReactor("delete", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			err := sdk.Deregister(brokerName)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
	Describe("RetrieveBrokers", func() {
		It("Calls the generated v1beta1 List methods", func() {
			brokers, err := sdk.RetrieveBrokers(ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(brokers).Should(ConsistOf(csb, csb2, sb, sb2))
			actions := svcCatClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
			Expect(svcCatClient.Actions()[1].Matches("list", "servicebrokers")).To(BeTrue())
		})
		It("Filters by namespace scope", func() {
			brokers, err := sdk.RetrieveBrokers(ScopeOptions{Scope: NamespaceScope, Namespace: "default"})

			Expect(err).NotTo(HaveOccurred())
			Expect(brokers).Should(ConsistOf(sb))
			actions := svcCatClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "servicebrokers")).To(BeTrue())
		})
		It("Filters by cluster scope", func() {
			brokers, err := sdk.RetrieveBrokers(ScopeOptions{Scope: ClusterScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(brokers).Should(ConsistOf(csb, csb2))
			actions := svcCatClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
		})
		It("Bubbles up cluster-scoped errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient
			_, err := sdk.RetrieveBrokers(ScopeOptions{Scope: AllScope})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
		})
		It("Bubbles up namespace-scoped errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "servicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient
			_, err := sdk.RetrieveBrokers(ScopeOptions{Scope: AllScope})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("list", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[1].Matches("list", "servicebrokers")).To(BeTrue())
		})
	})
	Describe("RetrieveBroker", func() {
		It("Calls the generated v1beta1 List method with the passed in broker", func() {
			broker, err := sdk.RetrieveBroker(csb.Name)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).To(Equal(csb))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csb.Name))
		})
		It("Bubbles up errors", func() {
			brokerName := "banana"

			broker, err := sdk.RetrieveBroker(brokerName)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(brokerName))
		})
	})
	Describe("RetrieveBrokerByClass", func() {
		It("Calls the generated v1beta1 List method with the passed in class's parent broker", func() {
			sc := &v1beta1.ClusterServiceClass{Spec: v1beta1.ClusterServiceClassSpec{ClusterServiceBrokerName: csb.Name}}
			broker, err := sdk.RetrieveBrokerByClass(sc)

			Expect(broker).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csb.Name))
		})

		It("Bubbles up errors", func() {
			brokerName := "banana"
			sc := &v1beta1.ClusterServiceClass{Spec: v1beta1.ClusterServiceClassSpec{ClusterServiceBrokerName: brokerName}}
			broker, err := sdk.RetrieveBrokerByClass(sc)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(brokerName))
		})
	})
	Describe("Register", func() {
		It("creates a broker by calling the v1beta1 Create method with the passed in arguements", func() {
			brokerName := "potato_broker"
			url := "http://potato.com"

			broker, err := sdk.Register(brokerName, url)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).NotTo(BeNil())
			Expect(broker.Name).To(Equal(brokerName))
			Expect(broker.Spec.URL).To(Equal(url))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("create", "clusterservicebrokers")).To(BeTrue())
			objectFromRequest := actions[0].(testing.CreateActionImpl).Object.(*v1beta1.ClusterServiceBroker)
			Expect(objectFromRequest.ObjectMeta.Name).To(Equal(brokerName))
			Expect(objectFromRequest.Spec.URL).To(Equal(url))
		})
		It("Bubbles up errors", func() {
			errorMessage := "error provisioning broker"
			brokerName := "potato_broker"
			url := "http://potato.com"
			badClient := &fake.Clientset{}
			badClient.AddReactor("create", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			broker, err := sdk.Register(brokerName, url)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
	Describe("Sync", func() {
		It("Useds the generated v1beta1 Retrieve method to get the broker, and then updates it with a new RelistRequests", func() {
			err := sdk.Sync(csb.Name, 3)
			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			Expect(len(actions) >= 2).To(BeTrue())
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csb.Name))

			Expect(actions[1].Matches("update", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[1].(testing.UpdateActionImpl).Object.(*v1beta1.ClusterServiceBroker).Spec.RelistRequests).Should(BeNumerically(">", 0))
		})
	})
})
