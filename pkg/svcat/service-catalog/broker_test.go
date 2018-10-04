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
	"errors"
	"fmt"
	"time"

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
		csb.Status.Conditions = append(csb.Status.Conditions,
			v1beta1.ServiceBrokerCondition{
				Type:   v1beta1.ServiceBrokerConditionReady,
				Status: v1beta1.ConditionTrue,
			})
		csb2 = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "barbaz"}}
		csb2.Status.Conditions = append(csb2.Status.Conditions,
			v1beta1.ServiceBrokerCondition{
				Type:   v1beta1.ServiceBrokerConditionFailed,
				Status: v1beta1.ConditionTrue,
			})
		sb = &v1beta1.ServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "default"}}
		sb2 = &v1beta1.ServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "ns2"}}
		svcCatClient = fake.NewSimpleClientset(csb, csb2, sb, sb2)
		sdk = &SDK{
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("BrokerHasStatus", func() {
		It("returns true if the provided broker has the provided status", func() {
			hasStatus := sdk.BrokerHasStatus(csb, v1beta1.ServiceBrokerConditionReady)
			Expect(hasStatus).To(BeTrue())
		})
		It("returns false if the provided broker does not have the provided status marked true", func() {
			hasStatus := sdk.BrokerHasStatus(csb2, v1beta1.ServiceBrokerConditionReady)
			Expect(hasStatus).To(BeFalse())
		})
	})
	Describe("Deregister", func() {
		It("deletes a ClusterServiceBroker by calling the v1beta1 Delete method with the passed in arguement", func() {
			brokerName := "foobar"
			scopeOptions := ScopeOptions{
				Namespace: "",
				Scope:     ClusterScope,
			}
			err := sdk.Deregister(brokerName, &scopeOptions)

			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("delete", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.DeleteActionImpl).Name).To(Equal(brokerName))
		})
		It("deletes a namespaced ServiceBroker by calling the v1beta1 Delete method with the passed in arguement", func() {
			scopeOptions := ScopeOptions{
				Namespace: sb.Namespace,
				Scope:     NamespaceScope,
			}
			err := sdk.Deregister(sb.Name, &scopeOptions)

			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("delete", "servicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.DeleteActionImpl).Name).To(Equal(sb.Name))
		})
		It("Bubbles up errors", func() {
			errorMessage := "error deregistering broker"
			brokerName := "potato_broker"
			scopeOptions := ScopeOptions{
				Namespace: "",
				Scope:     ClusterScope,
			}
			badClient := &fake.Clientset{}
			badClient.AddReactor("delete", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			err := sdk.Deregister(brokerName, &scopeOptions)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
		It("Bubbles up namespaced broker errors", func() {
			errorMessage := "error deregistering broker"
			brokerName := "potato_broker"
			scopeOptions := ScopeOptions{
				Namespace: sb.Namespace,
				Scope:     NamespaceScope,
			}
			badClient := &fake.Clientset{}
			badClient.AddReactor("delete", "servicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			err := sdk.Deregister(brokerName, &scopeOptions)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
	Describe("IsBrokerFailed", func() {
		It("returns true if the broker is in the failed status", func() {
			status := sdk.IsBrokerFailed(csb2)
			Expect(status).To(BeTrue())
		})
		It("returns false if the broker is not in the failed status", func() {
			status := sdk.IsBrokerFailed(csb)
			Expect(status).To(BeFalse())
		})
	})
	Describe("IsBrokerReady", func() {
		It("returns true if the broker is in the ready status", func() {
			status := sdk.IsBrokerReady(csb)
			Expect(status).To(BeTrue())
		})
		It("returns false if the broker is not in the ready status", func() {
			status := sdk.IsBrokerReady(csb2)
			Expect(status).To(BeFalse())
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
		It("creates a namespaced broker by calling the v1beta1 Create method with the passed in arguments", func() {
			brokerName := "potato_broker"
			url := "http://potato.com"
			basicSecret := "potatobasicsecret"
			caFile := "assets/ca"
			namespace := "potatonamespace"
			planRestrictions := []string{"potatoplana", "potatoplanb"}
			relistBehavior := v1beta1.ServiceBrokerRelistBehaviorDuration
			relistDuration := &metav1.Duration{Duration: 10 * time.Minute}
			skipTLS := true

			opts := &RegisterOptions{
				BasicSecret:      basicSecret,
				CAFile:           caFile,
				Namespace:        namespace,
				PlanRestrictions: planRestrictions,
				RelistBehavior:   relistBehavior,
				RelistDuration:   relistDuration,
				SkipTLS:          skipTLS,
			}
			scopeOpts := &ScopeOptions{
				Namespace: namespace,
				Scope:     NamespaceScope,
			}
			broker, err := sdk.Register(brokerName, url, opts, scopeOpts)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).NotTo(BeNil())
			Expect(broker.GetName()).To(Equal(brokerName))
			Expect(broker.GetURL()).To(Equal(url))
			Expect(broker.GetSpec().CABundle).To(Equal([]byte("foo\n")))
			Expect(broker.GetSpec().InsecureSkipTLSVerify).To(BeTrue())
			Expect(broker.GetSpec().RelistBehavior).To(Equal(relistBehavior))
			Expect(broker.GetSpec().RelistDuration).To(Equal(relistDuration))
			Expect(broker.GetSpec().CatalogRestrictions.ServicePlan).To(Equal(planRestrictions))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("create", "servicebrokers")).To(BeTrue())
			objectFromRequest := actions[0].(testing.CreateActionImpl).Object.(*v1beta1.ServiceBroker)
			Expect(objectFromRequest.ObjectMeta.Name).To(Equal(brokerName))
			Expect(objectFromRequest.Spec.URL).To(Equal(url))
			Expect(objectFromRequest.Spec.AuthInfo.Basic.SecretRef.Name).To(Equal(basicSecret))
			Expect(objectFromRequest.Spec.CABundle).To(Equal([]byte("foo\n")))
			Expect(objectFromRequest.Spec.InsecureSkipTLSVerify).To(BeTrue())
			Expect(objectFromRequest.Spec.RelistBehavior).To(Equal(relistBehavior))
			Expect(objectFromRequest.Spec.RelistDuration).To(Equal(relistDuration))
			Expect(objectFromRequest.Spec.CatalogRestrictions.ServicePlan).To(Equal(planRestrictions))
		})
		It("creates a namespace service broker with a bearer secret", func() {
			brokerName := "potato_broker"
			url := "http://potato.com"
			namespace := "potatonamespace"
			bearerSecret := "potatobearersecret"
			opts := &RegisterOptions{
				Namespace:    namespace,
				BearerSecret: bearerSecret,
			}
			scopeOpts := &ScopeOptions{
				Namespace: namespace,
				Scope:     NamespaceScope,
			}

			broker, err := sdk.Register(brokerName, url, opts, scopeOpts)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).NotTo(BeNil())
			Expect(broker.GetName()).To(Equal(brokerName))
			Expect(broker.GetURL()).To(Equal(url))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("create", "servicebrokers")).To(BeTrue())
			objectFromRequest := actions[0].(testing.CreateActionImpl).Object.(*v1beta1.ServiceBroker)
			Expect(objectFromRequest.ObjectMeta.Name).To(Equal(brokerName))
			Expect(objectFromRequest.Spec.URL).To(Equal(url))
			Expect(objectFromRequest.Spec.AuthInfo.Bearer.SecretRef.Name).To(Equal(bearerSecret))
		})
		It("Bubbles up namespace service broker errors", func() {
			errorMessage := "error provisioning broker"
			brokerName := "potato_broker"
			url := "http://potato.com"
			badClient := &fake.Clientset{}
			badClient.AddReactor("create", "servicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient
			scopeOpts := &ScopeOptions{
				Namespace: "default",
				Scope:     NamespaceScope,
			}

			broker, err := sdk.Register(brokerName, url, &RegisterOptions{}, scopeOpts)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
		It("creates a cluster service broker by calling the v1beta1 Create method with the passed in arguments", func() {
			brokerName := "potato_broker"
			url := "http://potato.com"
			basicSecret := "potatobasicsecret"
			caFile := "assets/ca"
			namespace := "potatonamespace"
			planRestrictions := []string{"potatoplana", "potatoplanb"}
			relistBehavior := v1beta1.ServiceBrokerRelistBehaviorDuration
			relistDuration := &metav1.Duration{Duration: 10 * time.Minute}
			skipTLS := true

			opts := &RegisterOptions{
				BasicSecret:      basicSecret,
				CAFile:           caFile,
				Namespace:        namespace,
				PlanRestrictions: planRestrictions,
				RelistBehavior:   relistBehavior,
				RelistDuration:   relistDuration,
				SkipTLS:          skipTLS,
			}
			scopeOpts := &ScopeOptions{
				Namespace: namespace,
				Scope:     ClusterScope,
			}
			broker, err := sdk.Register(brokerName, url, opts, scopeOpts)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).NotTo(BeNil())
			Expect(broker.GetName()).To(Equal(brokerName))
			Expect(broker.GetURL()).To(Equal(url))
			Expect(broker.GetSpec().CABundle).To(Equal([]byte("foo\n")))
			Expect(broker.GetSpec().InsecureSkipTLSVerify).To(BeTrue())
			Expect(broker.GetSpec().RelistBehavior).To(Equal(relistBehavior))
			Expect(broker.GetSpec().RelistDuration).To(Equal(relistDuration))
			Expect(broker.GetSpec().CatalogRestrictions.ServicePlan).To(Equal(planRestrictions))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("create", "clusterservicebrokers")).To(BeTrue())
			objectFromRequest := actions[0].(testing.CreateActionImpl).Object.(*v1beta1.ClusterServiceBroker)
			Expect(objectFromRequest.ObjectMeta.Name).To(Equal(brokerName))
			Expect(objectFromRequest.Spec.URL).To(Equal(url))
			Expect(objectFromRequest.Spec.AuthInfo.Basic.SecretRef.Name).To(Equal(basicSecret))
			Expect(objectFromRequest.Spec.CABundle).To(Equal([]byte("foo\n")))
			Expect(objectFromRequest.Spec.InsecureSkipTLSVerify).To(BeTrue())
			Expect(objectFromRequest.Spec.RelistBehavior).To(Equal(relistBehavior))
			Expect(objectFromRequest.Spec.RelistDuration).To(Equal(relistDuration))
			Expect(objectFromRequest.Spec.CatalogRestrictions.ServicePlan).To(Equal(planRestrictions))
		})
		It("creates a cluster service broker with a bearer secret", func() {
			brokerName := "potato_broker"
			url := "http://potato.com"
			namespace := "potatonamespace"
			bearerSecret := "potatobearersecret"
			opts := &RegisterOptions{
				Namespace:    namespace,
				BearerSecret: bearerSecret,
			}
			scopeOpts := &ScopeOptions{
				Namespace: namespace,
				Scope:     ClusterScope,
			}

			broker, err := sdk.Register(brokerName, url, opts, scopeOpts)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).NotTo(BeNil())
			Expect(broker.GetName()).To(Equal(brokerName))
			Expect(broker.GetURL()).To(Equal(url))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("create", "clusterservicebrokers")).To(BeTrue())
			objectFromRequest := actions[0].(testing.CreateActionImpl).Object.(*v1beta1.ClusterServiceBroker)
			Expect(objectFromRequest.ObjectMeta.Name).To(Equal(brokerName))
			Expect(objectFromRequest.Spec.URL).To(Equal(url))
			Expect(objectFromRequest.Spec.AuthInfo.Bearer.SecretRef.Name).To(Equal(bearerSecret))
		})
		It("Bubbles up cluster service broker errors", func() {
			errorMessage := "error provisioning broker"
			brokerName := "potato_broker"
			url := "http://potato.com"
			badClient := &fake.Clientset{}
			badClient.AddReactor("create", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient
			scopeOpts := &ScopeOptions{
				Namespace: "",
				Scope:     ClusterScope,
			}

			broker, err := sdk.Register(brokerName, url, &RegisterOptions{}, scopeOpts)

			Expect(broker).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
		})
	})
	Describe("Sync", func() {
		It("Uses the generated v1beta1 Retrieve method to get the broker, and then updates it with a new RelistRequests", func() {
			err := sdk.Sync(csb.Name, ScopeOptions{Scope: ClusterScope}, 3)
			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			Expect(len(actions) >= 2).To(BeTrue())
			Expect(actions[0].Matches("get", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csb.Name))

			Expect(actions[1].Matches("update", "clusterservicebrokers")).To(BeTrue())
			Expect(actions[1].(testing.UpdateActionImpl).Object.(*v1beta1.ClusterServiceBroker).Spec.RelistRequests).Should(BeNumerically(">", 0))
		})
		It("Uses the generated v1beta1 Retrieve method to get the broker with namespace", func() {
			sdk.Sync(csb.Name, ScopeOptions{Scope: NamespaceScope, Namespace: "namespace"}, 3)

			actions := svcCatClient.Actions()
			Expect(len(actions) >= 1).To(BeTrue())
			Expect(actions[0].Matches("get", "servicebrokers")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csb.Name))
		})
	})
	Describe("WaitForBroker", func() {
		var (
			counter        int
			interval       time.Duration
			notReady       v1beta1.ServiceBrokerCondition
			notReadyBroker *v1beta1.ClusterServiceBroker
			timeout        time.Duration
			waitClient     *fake.Clientset
		)
		BeforeEach(func() {
			counter = 0
			interval = 100 * time.Millisecond
			notReady = v1beta1.ServiceBrokerCondition{Type: v1beta1.ServiceBrokerConditionReady, Status: v1beta1.ConditionFalse}
			notReadyBroker = &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: csb.Name}}
			notReadyBroker.Status.Conditions = []v1beta1.ServiceBrokerCondition{notReady}
			timeout = 1 * time.Second
			waitClient = &fake.Clientset{}

			waitClient.AddReactor("get", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				counter++
				return true, notReadyBroker, nil
			})
			sdk.ServiceCatalogClient = waitClient
		})

		It("waits until the broker is ready to return", func() {
			waitClient.PrependReactor("get", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				if counter > 5 {
					return true, csb, nil
				}
				return false, nil, nil
			})

			broker, err := sdk.WaitForBroker(csb.Name, interval, &timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(broker).To(Equal(csb))
			actions := waitClient.Actions()
			Expect(len(actions)).Should(BeNumerically(">", 1))
			for _, v := range actions {
				Expect(v.Matches("get", "clusterservicebrokers")).To(BeTrue())
				Expect(v.(testing.GetActionImpl).Name).To(Equal(csb.Name))
			}
		})
		It("waits until the broker is failed to return", func() {
			failedBroker := &v1beta1.ClusterServiceBroker{ObjectMeta: metav1.ObjectMeta{Name: csb.Name}}
			failed := v1beta1.ServiceBrokerCondition{Type: v1beta1.ServiceBrokerConditionFailed, Status: v1beta1.ConditionTrue}
			failedBroker.Status.Conditions = []v1beta1.ServiceBrokerCondition{failed}
			waitClient.PrependReactor("get", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				if counter > 5 {
					return true, failedBroker, nil
				}
				return false, nil, nil
			})

			broker, err := sdk.WaitForBroker(csb.Name, interval, &timeout)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker).To(Equal(failedBroker))
			actions := waitClient.Actions()
			Expect(len(actions)).Should(BeNumerically(">", 1))
			for _, v := range actions {
				Expect(v.Matches("get", "clusterservicebrokers")).To(BeTrue())
				Expect(v.(testing.GetActionImpl).Name).To(Equal(csb.Name))
			}
		})
		It("times out if the broker never becomes ready or failed", func() {
			broker, err := sdk.WaitForBroker(csb.Name, interval, &timeout)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timed out"))
			Expect(broker).To(Equal(notReadyBroker))
			actions := waitClient.Actions()
			Expect(len(actions)).Should(BeNumerically(">", 1))
			for _, v := range actions {
				Expect(v.Matches("get", "clusterservicebrokers")).To(BeTrue())
				Expect(v.(testing.GetActionImpl).Name).To(Equal(csb.Name))
			}
		})
		It("bubbles up errors", func() {
			errorMessage := "backend exploded"
			waitClient.PrependReactor("get", "clusterservicebrokers", func(action testing.Action) (bool, runtime.Object, error) {
				if counter > 5 {
					return true, nil, errors.New(errorMessage)
				}
				return false, nil, nil
			})

			broker, err := sdk.WaitForBroker(csb.Name, interval, &timeout)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
			Expect(broker).To(BeNil())
			actions := waitClient.Actions()
			Expect(len(actions)).Should(BeNumerically(">", 1))
			for _, v := range actions {
				Expect(v.Matches("get", "clusterservicebrokers")).To(BeTrue())
				Expect(v.(testing.GetActionImpl).Name).To(Equal(csb.Name))
			}

		})
	})
})
