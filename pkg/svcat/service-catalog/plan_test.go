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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan", func() {
	var (
		sdk          *SDK
		svcCatClient *fake.Clientset
		csc          *v1beta1.ClusterServiceClass
		sc           *v1beta1.ServiceClass
		csp          *v1beta1.ClusterServicePlan
		csp2         *v1beta1.ClusterServicePlan
		sp           *v1beta1.ServicePlan
		sp2          *v1beta1.ServicePlan
	)

	BeforeEach(func() {
		csc = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "someclass"}}
		csp = &v1beta1.ClusterServicePlan{ObjectMeta: metav1.ObjectMeta{Name: "foobar"}}
		csp2 = &v1beta1.ClusterServicePlan{
			ObjectMeta: metav1.ObjectMeta{Name: "clusterscopedplan"},
			Spec: v1beta1.ClusterServicePlanSpec{
				ClusterServiceClassRef: v1beta1.ClusterObjectReference{Name: csc.Name},
			},
		}
		sc = &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "somenamespacedclass", Namespace: "default"}}
		sp = &v1beta1.ServicePlan{
			ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: sc.Namespace},
			Spec: v1beta1.ServicePlanSpec{
				ServiceClassRef: v1beta1.LocalObjectReference{Name: sc.Name},
			},
		}
		sp2 = &v1beta1.ServicePlan{ObjectMeta: metav1.ObjectMeta{Name: "namespacescopedplan", Namespace: "ns2"}}
		svcCatClient = fake.NewSimpleClientset(csc, csp, csp2, sc, sp, sp2)
		sdk = &SDK{
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("RetrivePlans", func() {
		It("Calls the generated v1beta1 List method", func() {
			plans, err := sdk.RetrievePlans("", ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plans).Should(ConsistOf(csp, csp2, sp, sp2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(svcCatClient.Actions()[1].Matches("list", "serviceplans")).To(BeTrue())
		})
		It("Filters by namespace scope", func() {
			plans, err := sdk.RetrievePlans("", ScopeOptions{Scope: NamespaceScope, Namespace: "default"})

			Expect(err).NotTo(HaveOccurred())
			Expect(plans).Should(ConsistOf(sp))
			Expect(len(svcCatClient.Actions())).Should(Equal(1))
			Expect(svcCatClient.Actions()[0].Matches("list", "serviceplans")).To(BeTrue())
		})
		It("Filters by cluster scope", func() {
			plans, err := sdk.RetrievePlans("", ScopeOptions{Scope: ClusterScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plans).Should(ConsistOf(csp, csp2))
			Expect(len(svcCatClient.Actions())).Should(Equal(1))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
		It("Filter by class", func() {
			plans, err := sdk.RetrievePlans(csc.Name, ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plans).Should(ConsistOf(csp2))
			Expect(len(svcCatClient.Actions())).Should(Equal(2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(svcCatClient.Actions()[1].Matches("list", "serviceplans")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			errorMessage := "error retrieving list"
			badClient := &fake.Clientset{}
			badClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient
			_, err := sdk.RetrievePlans("", ScopeOptions{Scope: AllScope})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
	})
	Describe("RetrievePlanByName", func() {
		It("Calls the generated v1beta1 List method with the passed in plan name for cluster-scoped plans", func() {
			planName := csp.Name
			singleClient := &fake.Clientset{}
			singleClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{*csp}}, nil
			})
			sdk.ServiceCatalogClient = singleClient

			plan, err := sdk.RetrievePlanByName(planName, ScopeOptions{Scope: ClusterScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.GetName()).To(Equal(planName))
			actions := singleClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			opts := fields.Set{"spec.externalName": planName}
			Expect(actions[0].(testing.ListActionImpl).GetListRestrictions().Fields.Matches(opts)).To(BeTrue())
		})
		It("Calls the generated v1beta1 List method with the passed in plan name for namespace-scoped plans", func() {
			planName := sp.Name
			singleClient := &fake.Clientset{}
			singleClient.AddReactor("list", "serviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServicePlanList{Items: []v1beta1.ServicePlan{*sp}}, nil
			})
			sdk.ServiceCatalogClient = singleClient

			plan, err := sdk.RetrievePlanByName(planName, ScopeOptions{Scope: NamespaceScope, Namespace: "default"})

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.GetName()).To(Equal(planName))
			actions := singleClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "serviceplans")).To(BeTrue())
			opts := fields.Set{"spec.externalName": planName}
			Expect(actions[0].(testing.ListActionImpl).GetListRestrictions().Fields.Matches(opts)).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			planName := "not_real"
			errorMessage := "plan not found"
			badClient := &fake.Clientset{}
			badClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			plan, err := sdk.RetrievePlanByName(planName, ScopeOptions{Scope: ClusterScope})

			Expect(plan).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			opts := fields.Set{"spec.externalName": planName}
			Expect(actions[0].(testing.ListActionImpl).GetListRestrictions().Fields.Matches(opts)).To(BeTrue())
		})
	})
	Describe("RetrievePlanByClassAndName", func() {
		It("Calls the generated v1beta1 List method with the passed in class and plan name for cluster-scoped plans", func() {
			className := csc.Name
			planName := csp2.Name
			singleClient := &fake.Clientset{}
			singleClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			singleClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{*csp2}}, nil
			})
			sdk.ServiceCatalogClient = singleClient

			plan, err := sdk.RetrievePlanByClassAndName(className, planName, ScopeOptions{Scope: ClusterScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.GetName()).To(Equal(planName))
			actions := singleClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			className := csc.Name
			planName := csp2.Name
			errorMessage := "plan not found"
			badClient := &fake.Clientset{}
			badClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			badClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			plan, err := sdk.RetrievePlanByClassAndName(className, planName, ScopeOptions{Scope: ClusterScope})

			Expect(plan).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
	})
	Describe("RetrievePlanByClassIDAndName", func() {
		It("Calls the generated v1beta1 List method with the passed in class kube name and plan external name for cluster-scoped plans", func() {
			classKubeName := csc.Name
			planName := csp2.Name
			singleClient := &fake.Clientset{}
			singleClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			singleClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{*csp2}}, nil
			})
			sdk.ServiceCatalogClient = singleClient

			plan, err := sdk.RetrievePlanByClassIDAndName(classKubeName, planName, ScopeOptions{Namespace: sp.Namespace, Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.GetName()).To(Equal(planName))
			Expect(plan.GetClassID()).To(Equal(classKubeName))
			Expect(plan.GetNamespace()).To(Equal(""))
			actions := singleClient.Actions()
			Expect(len(actions)).To(Equal(2))
			fieldSelector := fields.OneTermEqualSelector(FieldClusterServiceClassRef, classKubeName)
			Expect(actions[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(fieldSelector))
			Expect(actions[1].Matches("list", "serviceplans")).To(BeTrue())
			Expect(actions[1].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(fieldSelector))
		})
		It("Calls the generated v1beta1 List method with the passed in class kube name and plan external name for namespace-scoped plans", func() {
			classKubeName := sc.Name
			planName := sp.Name
			returnPlanCalled := 0
			singleClient := &fake.Clientset{}
			singleClient.AddReactor("list", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceClassList{Items: []v1beta1.ServiceClass{*sc}}, nil
			})
			singleClient.AddReactor("list", "serviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				if returnPlanCalled > 0 {
					return true, &v1beta1.ServicePlanList{Items: []v1beta1.ServicePlan{*sp}}, nil
				}
				returnPlanCalled++
				return true, &v1beta1.ServicePlanList{Items: []v1beta1.ServicePlan{}}, nil
			})
			sdk.ServiceCatalogClient = singleClient

			plan, err := sdk.RetrievePlanByClassIDAndName(classKubeName, planName, ScopeOptions{Namespace: sp.Namespace, Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.GetName()).To(Equal(planName))
			Expect(plan.GetClassID()).To(Equal(classKubeName))
			Expect(plan.GetNamespace()).To(Equal(sp.Namespace))
			actions := singleClient.Actions()
			Expect(len(actions)).To(Equal(4))
			fieldSelector := fields.OneTermEqualSelector(FieldClusterServiceClassRef, classKubeName)
			Expect(actions[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(fieldSelector))
			Expect(actions[1].Matches("list", "serviceplans")).To(BeTrue())
			Expect(actions[1].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(fieldSelector))
			namespacedFieldSelector := fields.OneTermEqualSelector(FieldServiceClassRef, classKubeName)
			Expect(actions[2].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(actions[2].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(namespacedFieldSelector))
			Expect(actions[3].Matches("list", "serviceplans")).To(BeTrue())
			Expect(actions[3].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(namespacedFieldSelector))
		})
		It("Bubbles up errors", func() {
			classKubeName := csc.Name
			planName := csp.Name
			clusterErrorMessage := "clusterplan error"
			namespacedErrorMessage := "namespaceplan error"
			badClient := &fake.Clientset{}
			cspCalled := 0
			spCalled := 0
			badClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				if cspCalled > 0 {
					return true, &v1beta1.ClusterServicePlanList{Items: []v1beta1.ClusterServicePlan{}}, nil
				}
				cspCalled++
				return true, nil, fmt.Errorf(clusterErrorMessage)
			})
			badClient.AddReactor("list", "serviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				if cspCalled > 0 {
					return true, nil, fmt.Errorf(namespacedErrorMessage)
				}
				spCalled++
				return true, &v1beta1.ServicePlanList{Items: []v1beta1.ServicePlan{}}, nil
			})
			sdk.ServiceCatalogClient = badClient

			plan, err := sdk.RetrievePlanByClassIDAndName(classKubeName, planName, ScopeOptions{Scope: AllScope})

			Expect(plan).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(clusterErrorMessage))
			Expect(err.Error()).Should(ContainSubstring(namespacedErrorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(3))
			fieldSelector := fields.OneTermEqualSelector(FieldClusterServiceClassRef, classKubeName)
			Expect(actions[0].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(fieldSelector))
			namespacedFieldSelector := fields.OneTermEqualSelector(FieldServiceClassRef, classKubeName)
			Expect(actions[1].Matches("list", "clusterserviceplans")).To(BeTrue())
			Expect(actions[1].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(namespacedFieldSelector))
			Expect(actions[2].Matches("list", "serviceplans")).To(BeTrue())
			Expect(actions[2].(testing.ListAction).GetListRestrictions().Fields).To(ContainElement(namespacedFieldSelector))
		})
	})
	Describe("RetrievePlanByID", func() {
		It("Calls the generated v1beta1 get method with the passed in Kubernetes name for cluster-scoped plans", func() {
			planID := csp.Name
			_, err := sdk.RetrievePlanByID(planID, ScopeOptions{Scope: ClusterScope})
			Expect(err).NotTo(HaveOccurred())
			actions := svcCatClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(planID))
		})
		It("Calls the generated v1beta1 get method with the passed in Kubernetes name for cluster-scoped plans", func() {
			planID := sp.Name
			_, err := sdk.RetrievePlanByID(planID, ScopeOptions{Scope: NamespaceScope, Namespace: "default"})
			Expect(err).NotTo(HaveOccurred())
			actions := svcCatClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "serviceplans")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(planID))
		})
		It("Bubbles up errors", func() {
			planID := "not_real"
			errorMessage := "plan not found"
			badClient := &fake.Clientset{}
			badClient.AddReactor("get", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			plan, err := sdk.RetrievePlanByID(planID, ScopeOptions{Scope: ClusterScope})

			Expect(plan).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := badClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(planID))
		})
	})
})
