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

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"

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
		csc = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{
			Name: "foobar", ResourceVersion: "1",
			Labels: map[string]string{
				v1beta1.GroupName + "/" + v1beta1.FilterSpecExternalName: "foobar",
			},
		}}
		csc2 = &v1beta1.ClusterServiceClass{
			ObjectMeta: metav1.ObjectMeta{Name: "barbaz", ResourceVersion: "1",
				Labels: map[string]string{
					v1beta1.GroupName + "/" + v1beta1.FilterSpecExternalName: "barbaz",
				},
			}}

		sc = &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{
			Name: "foobar", Namespace: "default", ResourceVersion: "1",
			Labels: map[string]string{
				v1beta1.GroupName + "/" + v1beta1.FilterSpecExternalName: "foobar",
			},
		}}
		sc2 = &v1beta1.ServiceClass{ObjectMeta: metav1.ObjectMeta{
			Name: "barbaz", Namespace: "ns2", ResourceVersion: "1",
			Labels: map[string]string{
				v1beta1.GroupName + "/" + v1beta1.FilterSpecExternalName: "barbaz",
			},
		}}
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
			badClient := fake.NewSimpleClientset()
			errorMessage := "error retrieving list"
			badClient.PrependReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
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
			realClient := fake.NewSimpleClientset(csc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.RetrieveClassByName(className, ScopeOptions{Scope: AllScope})

			Expect(err).NotTo(HaveOccurred())
			Expect(class).To(Equal(csc))
			actions := realClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[1].Matches("list", "serviceclasses")).To(BeTrue())

			requirements, selectable := actions[0].(testing.ListActionImpl).GetListRestrictions().Labels.Requirements()
			Expect(selectable).Should(BeTrue())
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].String()).To(Equal("servicecatalog.k8s.io/spec.externalName=foobar"))

			requirements, selectable = actions[1].(testing.ListActionImpl).GetListRestrictions().Labels.Requirements()
			Expect(selectable).Should(BeTrue())
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].String()).To(Equal("servicecatalog.k8s.io/spec.externalName=foobar"))
		})
		It("Bubbles up errors", func() {
			className := "notreal_class"
			emptyClient := fake.NewSimpleClientset()
			sdk = &SDK{
				ServiceCatalogClient: emptyClient,
			}
			class, err := sdk.RetrieveClassByName(className, ScopeOptions{Scope: AllScope})

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := emptyClient.Actions()
			Expect(actions[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
			requirements, selectable := actions[0].(testing.ListActionImpl).GetListRestrictions().Labels.Requirements()
			Expect(selectable).Should(BeTrue())
			Expect(requirements).ShouldNot(BeEmpty())
			Expect(requirements[0].String()).To(Equal("servicecatalog.k8s.io/spec.externalName=notreal_class"))
		})
	})
	Describe("RetrieveClassByID", func() {
		It("Calls the generated v1beta1 get methods for clusterserviceclass and serviceclass with the passed in name", func() {
			classID := csc.Name
			realClient := fake.NewSimpleClientset(csc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: "",
				Scope:     AllScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(classID))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
			Expect(actions[1].Matches("get", "serviceclasses")).To(BeTrue())
			Expect(actions[1].(testing.GetActionImpl).Name).To(Equal(classID))
		})
		It("Calls only the generated v1beta1 get method for clusterserviceclass when called with cluster scope", func() {
			classID := csc.Name
			realClient := fake.NewSimpleClientset(csc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: "",
				Scope:     ClusterScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(classID))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
		})
		It("Calls only the generated v1beta1 get method for serviceclass when called with namespace scope", func() {
			classID := sc.Name
			realClient := fake.NewSimpleClientset(sc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: sc.Namespace,
				Scope:     NamespaceScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(classID))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "serviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
		})
		It("Bubbles up errors", func() {
			errorMessage := "not found"
			emptyClient := fake.NewSimpleClientset()
			emptyClient.PrependReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, errors.New(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: emptyClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: "",
				Scope:     ClusterScope,
			}
			class, err := sdk.RetrieveClassByID("not_real", scopeOptions)

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := emptyClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
		})
		It("errors when it finds multiple classes", func() {
			classID := csc.Name
			realClient := fake.NewSimpleClientset(csc, sc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: sc.Namespace,
				Scope:     AllScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(MultipleClassesFoundError))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
			Expect(actions[1].Matches("get", "serviceclasses")).To(BeTrue())
			Expect(actions[1].(testing.GetActionImpl).Name).To(Equal(classID))
		})
		It("doesn't short circuit on not-found errors", func() {
			classID := sc.Name
			realClient := fake.NewSimpleClientset(sc)
			realClient.PrependReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, apierrors.NewNotFound(v1beta1.Resource("clusterserviceclass"), classID)
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: sc.Namespace,
				Scope:     AllScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(class.GetName()).To(Equal(classID))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
			Expect(actions[1].Matches("get", "serviceclasses")).To(BeTrue())
			Expect(actions[1].(testing.GetActionImpl).Name).To(Equal(classID))
		})
		It("errors when it receives not-found errors for both types", func() {
			classID := sc.Name
			realClient := fake.NewSimpleClientset()
			realClient.PrependReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, apierrors.NewNotFound(v1beta1.Resource("clusterserviceclass"), classID)
			})
			realClient.PrependReactor("get", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, apierrors.NewNotFound(v1beta1.Resource("clusterserviceclass"), classID)
			})
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			scopeOptions := ScopeOptions{
				Namespace: sc.Namespace,
				Scope:     AllScope,
			}
			class, err := sdk.RetrieveClassByID(classID, scopeOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no matching class found"))
			Expect(class).To(BeNil())
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(2))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classID))
			Expect(actions[1].Matches("get", "serviceclasses")).To(BeTrue())
			Expect(actions[1].(testing.GetActionImpl).Name).To(Equal(classID))
		})
	})
	Describe("RetrieveClassByPlan", func() {
		It("Calls the generated v1beta1 get method for ClusterServiceClasses with the plan's parent service class's name if the plan is a ClusterServicePlan", func() {
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
			realClient := fake.NewSimpleClientset(csc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}
			class, err := sdk.RetrieveClassByPlan(classPlan)
			Expect(err).NotTo(HaveOccurred())
			Expect(class).To(Equal(csc))
			actions := realClient.Actions()
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(csc.Name))
		})
		It("Bubbles up errors from the v1beta1 ClusterServiceClass method", func() {
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
			badClient := fake.NewSimpleClientset()
			badClient.PrependReactor("get", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
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
			Expect(len(actions)).To(Equal(1))
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(fakeClassName))
		})
		Context("ServiceClass", func() {
			var (
				classPlan *v1beta1.ServicePlan
			)
			BeforeEach(func() {
				classPlan = &v1beta1.ServicePlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foobar_plan",
						Namespace: sc.Namespace,
					},
					Spec: v1beta1.ServicePlanSpec{
						ServiceClassRef: v1beta1.LocalObjectReference{
							Name: sc.Name,
						},
					},
				}
			})
			It("Calls the generated v1beta1 get method for ServiceClasses with the plan's parent service class's name if the plan is a ServicePlan", func() {
				realClient := fake.NewSimpleClientset(sc)
				sdk = &SDK{
					ServiceCatalogClient: realClient,
				}
				class, err := sdk.RetrieveClassByPlan(classPlan)
				Expect(err).NotTo(HaveOccurred())
				Expect(class).To(Equal(sc))
				actions := realClient.Actions()
				Expect(len(actions)).To(Equal(1))
				Expect(actions[0].Matches("get", "serviceclasses")).To(BeTrue())
				Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classPlan.Spec.ServiceClassRef.Name))
				Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(classPlan.Namespace))
			})
			It("Bubbles up errors from the v1beta1 ServiceClass method", func() {
				errorMessage := "not found"

				badClient := fake.NewSimpleClientset()
				badClient.PrependReactor("get", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
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
				Expect(len(actions)).To(Equal(1))
				Expect(actions[0].Matches("get", "serviceclasses")).To(BeTrue())
				Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(classPlan.Spec.ServiceClassRef.Name))
				Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(classPlan.Namespace))
			})
		})
	})
	Describe("CreateClassFrom", func() {
		It("Calls the generated v1beta1 create method for cluster service class with the passed in class", func() {
			className := "newclass"

			realClient := fake.NewSimpleClientset(csc)
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
			classNamespace := sc.Namespace
			realClient := fake.NewSimpleClientset(sc)
			sdk = &SDK{
				ServiceCatalogClient: realClient,
			}

			class, err := sdk.CreateClassFrom(CreateClassFromOptions{Name: className, Namespace: classNamespace, From: sc.Name, Scope: NamespaceScope})

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

			realClient := fake.NewSimpleClientset(csc)
			realClient.PrependReactor("create", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
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
			classNamespace := sc.Namespace
			errorMessage := "unable to create service class"

			realClient := fake.NewSimpleClientset(sc)
			realClient.PrependReactor("create", "serviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
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
