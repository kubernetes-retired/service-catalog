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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Class", func() {
	var (
		sdk          *SDK
		svcCatClient *fake.Clientset
		csc          *v1beta1.ClusterServiceClass
		csc2         *v1beta1.ClusterServiceClass
		sc           *v1beta1.ServiceClass
		sc2          *v1beta1.ServiceClass
	)

	BeforeEach(func() {
		csc = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "foobar", ResourceVersion: "1"}}
		csc2 = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", ResourceVersion: "1"}}
		sc = &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "default", ResourceVersion: "1"}}
		sc2 = &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "ns2", ResourceVersion: "1"}}
		svcCatClient = fake.NewSimpleClientset(csc, csc2, sc, sc2)
		sdk = &SDK{
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("RetrieveClasses", func() {
		It("Calls the generated v1beta1 List methods", func() {
			classes, err := sdk.RetrieveClasses(ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(classes).Should(ConsistOf(csc, csc2, sc, sc2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(svcCatClient.Actions()[1].Matches("list", "serviceclasses")).To(BeTrue())
		})
		It("Filters by namespace scope", func() {
			classes, err := sdk.RetrieveClasses(ScopeOptions{Scope: NamespaceScope, Namespace: "default"})

			Expect(err).NotTo(HaveOccurred())
			Expect(classes).Should(ConsistOf(sc))
			Expect(len(svcCatClient.Actions())).Should(Equal(1))
			Expect(svcCatClient.Actions()[0].Matches("list", "serviceclasses")).To(BeTrue())

		})
		It("Filters by cluster scope", func() {
			classes, err := sdk.RetrieveClasses(ScopeOptions{Scope: ClusterScope, Namespace: "default"})

			Expect(err).NotTo(HaveOccurred())
			Expect(classes).Should(ConsistOf(csc, csc2))
			Expect(len(svcCatClient.Actions())).Should(Equal(1))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceclasses")).To(BeTrue())

		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: badClient,
			}

			_, err := sdk.RetrieveClasses(ScopeOptions{Scope: AllScope})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
		})
	})
	Describe("RetrieveClassByName", func() {
		It("Calls the generated v1beta1 List method with the passed in class name", func() {
			className := csc.Name
			realClient := &fake.Clientset{}
			realClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.RetrieveClassByName(className, ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(class).To(Equal(csc))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("list", "serviceclasses")).To(BeTrue())

			requirements := actions[0].(testing.ListActionImpl).GetListRestrictions().Fields.Requirements()
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].Field).To(Equal("spec.externalName"))
			Expect(requirements[0].Value).To(Equal(className))

			requirements = actions[1].(testing.ListActionImpl).GetListRestrictions().Fields.Requirements()
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].Field).To(Equal("spec.externalName"))
			Expect(requirements[0].Value).To(Equal(className))
		})
		It("Bubbles up errors", func() {
			className := "notreal_class"
			emptyClient := &fake.Clientset{}
			emptyClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{}}, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: emptyClient,
			}
			class, err := sdk.RetrieveClassByName(className, ScopeOptions{Scope: AllScope})

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := emptyClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			requirements := actions[0].(testing.ListActionImpl).GetListRestrictions().Fields.Requirements()
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].Field).To(Equal("spec.externalName"))
			Expect(requirements[0].Value).To(Equal(className))
		})
	})
	Describe("RetrieveClassByID", func() {
		It("Calls the generated v1beta1 get method", func() {
			classID := fmt.Sprintf("%v", csc.UID)
			realClient := &fake.Clientset{}
			realClient.AddReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, csc, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			class, err := sdk.RetrieveClassByID(classID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fmt.Sprintf("%v", class.UID)).To(Equal(classID))
			actions := realClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			errorMessage := "not found"
			emptyClient := &fake.Clientset{}
			emptyClient.AddReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: emptyClient,
			}
			class, err := sdk.RetrieveClassByID("not_real")

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := emptyClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
		})
	})
	Describe("RetrieveClassByPlan", func() {
		It("Calls the generated v1beta1 get method with the plan's parent service class's name", func() {
			classPlan := &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foobar_plan",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: csc.Name,
					},
				},
			}
			realClient := &fake.Clientset{}
			realClient.AddReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, csc, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			class, err := sdk.RetrieveClassByPlan(classPlan)
			Expect(err).NotTo(HaveOccurred())
			Expect(class).To(Equal(csc))
			actions := realClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csc.Name))
		})
		It("Bubbles up errors", func() {
			fakeClassName := "not_real"
			errorMessage := "not found"

			classPlan := &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foobar_plan",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: fakeClassName,
					},
				},
			}
			badClient := &fake.Clientset{}
			badClient.AddReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: badClient,
			}
			class, err := sdk.RetrieveClassByPlan(classPlan)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMessage))
			Expect(class).To(BeNil())
			actions := badClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(fakeClassName))
		})
	})
	Describe("CreateClassFrom", func() {
		It("Calls the generated v1beta1 create method for cluster service class with the passed in class", func() {
			className := "newclass"

			realClient := &fake.Clientset{}
			realClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			realClient.AddReactor("create", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: className}}, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.CreateClassFrom(CreateClassFromOptions{Name: className, From: csc.Name, Scope: ClusterScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(className))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("create", "clusterserviceclasses")).To(BeTrue())
			objectFromRequest := actions[1].(testing.CreateActionImpl).Object.(*v1beta1.ClusterServiceClass)
			Expect(objectFromRequest.Name).To(Equal(className))
			Expect(objectFromRequest.ResourceVersion).To(BeEmpty())
		})
		It("Calls the generated v1beta1 create method for service class with the passed in class", func() {
			className := "newclass"
			classNamespace := "newnamespace"

			realClient := &fake.Clientset{}
			realClient.AddReactor("list", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceClassList{Items: []v1beta1.ServiceClass{*sc}}, nil
			})
			realClient.AddReactor("create", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: className, Namespace: classNamespace}}, nil
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.CreateClassFrom(CreateClassFromOptions{Name: className, Namespace: classNamespace, From: csc.Name, Scope: NamespaceScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(className))
			Expect(class.GetNamespace()).To(Equal(classNamespace))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "serviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("create", "serviceclasses")).To(BeTrue())
			objectFromRequest := actions[1].(testing.CreateActionImpl).Object.(*v1beta1.ServiceClass)
			Expect(objectFromRequest.Name).To(Equal(className))
			Expect(objectFromRequest.Namespace).To(Equal(classNamespace))
			Expect(objectFromRequest.ResourceVersion).To(BeEmpty())
		})
		It("Bubbles up errors for cluster service class", func() {
			className := "newclass"
			errorMessage := "unable to create cluster service class"

			realClient := &fake.Clientset{}
			realClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ClusterServiceClassList{Items: []v1beta1.ClusterServiceClass{*csc}}, nil
			})
			realClient.AddReactor("create", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.CreateClassFrom(CreateClassFromOptions{Name: className, From: csc.Name, Scope: ClusterScope})

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("create", "clusterserviceclasses")).To(BeTrue())
		})
		It("Bubbles up errors for service class", func() {
			className := "newclass"
			classNamespace := "newnamespace"
			errorMessage := "unable to create service class"

			realClient := &fake.Clientset{}
			realClient.AddReactor("list", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceClassList{Items: []v1beta1.ServiceClass{*sc}}, nil
			})
			realClient.AddReactor("create", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.CreateClassFrom(CreateClassFromOptions{Name: className, Namespace: classNamespace, From: csc.Name, Scope: NamespaceScope})

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "serviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("create", "serviceclasses")).To(BeTrue())
		})
	})
})
