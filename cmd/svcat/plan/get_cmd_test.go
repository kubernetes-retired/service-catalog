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

package plan_test

import (
	"bytes"
	"errors"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	. "github.com/kubernetes-sigs/service-catalog/cmd/svcat/plan"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	servicecatalogfakes "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "github.com/kubernetes-sigs/service-catalog/internal/test"
)

var _ = Describe("Get Plans Command", func() {
	Describe("NewGetPlansCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewGetCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("plans [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("List plans, optionally filtered by name, class, scope or namespace"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat get plans --scope namespace --namespace dev"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			classFlag := cmd.Flags().Lookup("class")
			Expect(classFlag).NotTo(BeNil())
			Expect(classFlag.Usage).To(ContainSubstring("Filter plans based on class. When --kube-name is specified, the class name is interpreted as a kubernetes name."))

			kubeNameFlag := cmd.Flags().Lookup("kube-name")
			Expect(kubeNameFlag).NotTo(BeNil())
			Expect(kubeNameFlag.Usage).To(ContainSubstring("Whether or not to get the plan by its Kubernetes name (the default is by external name)"))

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())

			namespaceFlag := cmd.Flags().Lookup("namespace")
			Expect(namespaceFlag).NotTo(BeNil())

			allNamespacesFlag := cmd.Flags().Lookup("all-namespaces")
			Expect(allNamespacesFlag).NotTo(BeNil())
		})
	})
	Describe("Validate", func() {
		It("allows plan name arg to be empty", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(BeNil())
		})
		It("optionally parses the plan name argument", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{"myplan"})
			Expect(err).To(BeNil())
			Expect(cmd.Name).To(Equal("myplan"))
		})
		It("populates kubeName and classKubeName when lookupByKubeName is set", func() {
			cmd := &GetCmd{
				LookupByKubeName: true,
				ClassFilter:      "myclass",
			}
			err := cmd.Validate([]string{"myplan", "--class", "foo", "--kube-name"})
			Expect(err).To(BeNil())
			Expect(cmd.KubeName).To(Equal("myplan"))
			Expect(cmd.ClassKubeName).To(Equal("myclass"))
		})
		It("parses a combined class/plan k8s name argument when --kube-name is set", func() {
			cmd := &GetCmd{
				LookupByKubeName: true,
			}
			err := cmd.Validate([]string{"myclass/myplan", "--kube-name"})
			Expect(err).To(BeNil())
			Expect(cmd.KubeName).To(Equal("myplan"))
			Expect(cmd.ClassKubeName).To(Equal("myclass"))
		})
		It("errors when passed an unparseable combined class/plan k8s name argument when --kube-name is set", func() {
			cmd := &GetCmd{
				LookupByKubeName: true,
			}
			combinationArg := "myclass/myplan/myotherthing"
			err := cmd.Validate([]string{combinationArg, "--kube-name"})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse class/plan k8s name combination"))
			Expect(err.Error()).To(ContainSubstring(combinationArg))
		})
		It("populates className when provided a class filter and --kube-name is not set", func() {
			cmd := &GetCmd{
				ClassFilter: "myclass",
			}
			err := cmd.Validate([]string{"myplan", "--class", "foo"})
			Expect(err).To(BeNil())
			Expect(cmd.Name).To(Equal("myplan"))
			Expect(cmd.ClassName).To(Equal("myclass"))
		})
		It("parses a combined class/plan name argument", func() {
			cmd := &GetCmd{}
			err := cmd.Validate([]string{"myclass/myplan"})
			Expect(err).To(BeNil())
			Expect(cmd.Name).To(Equal("myplan"))
			Expect(cmd.ClassName).To(Equal("myclass"))
		})
		It("errors when passed an unparseable combination arg", func() {
			cmd := &GetCmd{}
			combinationArg := "myclass/myplan/myotherthing"
			err := cmd.Validate([]string{combinationArg})
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse class/plan name combination"))
			Expect(err.Error()).To(ContainSubstring(combinationArg))
		})
	})
	Describe("Run", func() {
		var (
			cmd          *GetCmd
			fakeApp      *svcat.App
			fakeSDK      *servicecatalogfakes.FakeSvcatClient
			outputBuffer *bytes.Buffer

			defaultNamespace    string
			bananaNamespace     string
			clusterServiceClass *v1beta1.ClusterServiceClass
			defaultServiceClass *v1beta1.ServiceClass
			bananaServiceClass  *v1beta1.ServiceClass
			clusterServicePlan  *v1beta1.ClusterServicePlan
			defaultServicePlan  *v1beta1.ServicePlan
			bananaServicePlan   *v1beta1.ServicePlan
		)
		BeforeEach(func() {
			fakeSDK = new(servicecatalogfakes.FakeSvcatClient)
			fakeApp, _ = svcat.NewApp(nil, nil, "default")
			fakeApp.SvcatClient = fakeSDK
			outputBuffer = &bytes.Buffer{}

			defaultNamespace = "default"
			bananaNamespace = "banana"
			cmd = &GetCmd{
				Namespaced: &command.Namespaced{
					Context:   svcattest.NewContext(outputBuffer, fakeApp),
					Namespace: defaultNamespace,
				},
				Scoped: &command.Scoped{
					Scope: servicecatalog.AllScope,
				},
				Formatted: command.NewFormatted(),
			}

			clusterServiceClass = &v1beta1.ClusterServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "csc-123",
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: "myclusterserviclass",
					},
				},
			}
			clusterServicePlan = &v1beta1.ClusterServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "csp-123",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: clusterServiceClass.Name,
					},
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "myclusterserviceplan",
					},
				},
			}
			defaultServiceClass = &v1beta1.ServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dsc-456",
					Namespace: defaultNamespace,
				},
				Spec: v1beta1.ServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: "mydefaultserviceclass",
					},
				},
			}
			defaultServicePlan = &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dsp-456",
					Namespace: defaultNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					ServiceClassRef: v1beta1.LocalObjectReference{
						Name: defaultServiceClass.Name,
					},
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "mydefaultserviceplan",
					},
				},
			}
			bananaServiceClass = &v1beta1.ServiceClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bsc-456",
					Namespace: bananaNamespace,
				},
				Spec: v1beta1.ServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: "mybananaserviceclass",
					},
				},
			}
			bananaServicePlan = &v1beta1.ServicePlan{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bsp-456",
					Namespace: bananaNamespace,
				},
				Spec: v1beta1.ServicePlanSpec{
					ServiceClassRef: v1beta1.LocalObjectReference{
						Name: bananaServiceClass.Name,
					},
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "mybananaserviceplan",
					},
				},
			}
		})
		Context("getAll()", func() {
			It("Calls the pkg/svcat libs RetrieveClasses/RetrievePlans with namespace scope and current namespace", func() {
				fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{clusterServiceClass, defaultServiceClass}, nil)
				fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{clusterServicePlan, defaultServicePlan}, nil)
				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))
				Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(1))
				classID, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
				Expect(classID).To(Equal(""))
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))

				output := outputBuffer.String()
				Expect(output).To(ContainSubstring(clusterServiceClass.Spec.ExternalName))
				Expect(output).To(ContainSubstring(clusterServicePlan.Spec.ExternalName))
				Expect(output).To(ContainSubstring(defaultServiceClass.Spec.ExternalName))
				Expect(output).To(ContainSubstring(defaultServicePlan.Spec.ExternalName))
			})
			It("Bubbles up errors from RetrieveClasses", func() {
				errMsg := "error: burnt toast"
				fakeSDK.RetrieveClassesReturns(nil, errors.New(errMsg))

				err := cmd.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(0))
			})
			It("Bubbles up errors from RetrieveClasses", func() {
				errMsg := "error: too many cookies"
				fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{clusterServiceClass, defaultServiceClass}, nil)
				fakeSDK.RetrievePlansReturns(nil, errors.New(errMsg))

				err := cmd.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(1))
			})
			Context("When --namespace is set", func() {
				BeforeEach(func() {
					cmd.Scope = servicecatalog.NamespaceScope
				})
				It("Calls the pkg/svcat libs RetrieveClasses/RetrievePlans with namespace scope and current namespace", func() {
					fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{defaultServiceClass}, nil)
					fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{defaultServicePlan}, nil)

					err := cmd.Run()

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
					scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.NamespaceScope,
						Namespace: defaultNamespace,
					}))
					Expect(fakeSDK.RetrievePlansCallCount()).To(Equal(1))
					classID, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
					Expect(classID).To(Equal(""))
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.NamespaceScope,
						Namespace: defaultNamespace,
					}))

					output := outputBuffer.String()
					Expect(output).To(ContainSubstring(defaultServiceClass.Spec.ExternalName))
					Expect(output).To(ContainSubstring(defaultServicePlan.Spec.ExternalName))
				})
			})
			// this is the only way to set Namespace to blank,
			// as it's normally populated from the kube config
			Context("When --all-namespaces is set", func() {
				BeforeEach(func() {
					cmd.Scope = servicecatalog.AllScope
					cmd.Namespace = ""
				})
				It("Calls the pkg/svcat libs RetrieveClasses/RetrievePlans with all scope and current namespace", func() {
					fakeSDK.RetrieveClassesReturns([]servicecatalog.Class{clusterServiceClass, defaultServiceClass, bananaServiceClass}, nil)
					fakeSDK.RetrievePlansReturns([]servicecatalog.Plan{clusterServicePlan, defaultServicePlan, bananaServicePlan}, nil)

					err := cmd.Run()

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeSDK.RetrieveClassesCallCount()).To(Equal(1))
					scopeArg := fakeSDK.RetrieveClassesArgsForCall(0)
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.AllScope,
						Namespace: "",
					}))
					classID, scopeArg := fakeSDK.RetrievePlansArgsForCall(0)
					Expect(classID).To(Equal(""))
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.AllScope,
						Namespace: "",
					}))

					output := outputBuffer.String()
					Expect(output).To(ContainSubstring(clusterServiceClass.Spec.ExternalName))
					Expect(output).To(ContainSubstring(clusterServicePlan.Spec.ExternalName))
					Expect(output).To(ContainSubstring(defaultServiceClass.Spec.ExternalName))
					Expect(output).To(ContainSubstring(defaultServicePlan.Spec.ExternalName))
					Expect(output).To(ContainSubstring(bananaServiceClass.Spec.ExternalName))
					Expect(output).To(ContainSubstring(bananaServicePlan.Spec.ExternalName))
				})
			})
		})
		Context("get(), when an external name is provided", func() {
			BeforeEach(func() {
				cmd.Name = clusterServicePlan.Spec.ExternalName
			})
			It("Calls the pkg/svcat libs RetrievePlanByName/RetrieveClassByID with all scope and current namespace", func() {
				fakeSDK.RetrievePlanByNameReturns(clusterServicePlan, nil)
				fakeSDK.RetrieveClassByIDReturns(clusterServiceClass, nil)

				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(0))
				Expect(fakeSDK.RetrievePlanByClassAndNameCallCount()).To(Equal(0))
				Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(1))
				planName, scopeArg := fakeSDK.RetrievePlanByNameArgsForCall(0)
				Expect(planName).To(Equal(clusterServicePlan.Spec.ExternalName))
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
				classID, scopeArg := fakeSDK.RetrieveClassByIDArgsForCall(0)
				Expect(classID).To(Equal(clusterServicePlan.Spec.ClusterServiceClassRef.Name))
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))
			})
			It("Bubbles up errors from RetrievePlanByName", func() {
				errMsg := "error: strawberry jam"
				fakeSDK.RetrievePlanByNameReturns(nil, errors.New(errMsg))

				err := cmd.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(0))
				Expect(fakeSDK.RetrievePlanByClassAndNameCallCount()).To(Equal(0))
				Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(0))
			})
			It("Bubbles up errors from RetrieveClassByID", func() {
				errMsg := "error: toast improperly buttered"
				fakeSDK.RetrievePlanByNameReturns(clusterServicePlan, nil)
				fakeSDK.RetrieveClassByIDReturns(nil, errors.New(errMsg))

				err := cmd.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
			})
			Context("When a class external name is also provided", func() {
				BeforeEach(func() {
					cmd.ClassName = clusterServiceClass.Spec.ExternalName
				})
				It("Calls the pkg/svcat libs RetrievePlanByClassAndName/RetrieveClassByID with all scope and current namespace and the passed in class name", func() {
					fakeSDK.RetrievePlanByClassAndNameReturns(clusterServicePlan, nil)
					fakeSDK.RetrieveClassByIDReturns(clusterServiceClass, nil)

					err := cmd.Run()

					Expect(err).NotTo(HaveOccurred())
					Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(0))
					Expect(fakeSDK.RetrievePlanByClassAndNameCallCount()).To(Equal(1))
					Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(0))
					className, planName, scopeArg := fakeSDK.RetrievePlanByClassAndNameArgsForCall(0)
					Expect(className).To(Equal(clusterServiceClass.Spec.ExternalName))
					Expect(planName).To(Equal(clusterServicePlan.Spec.ExternalName))
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.AllScope,
						Namespace: defaultNamespace,
					}))
					Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
					classID, scopeArg := fakeSDK.RetrieveClassByIDArgsForCall(0)
					Expect(classID).To(Equal(clusterServicePlan.Spec.ClusterServiceClassRef.Name))
					Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
						Scope:     servicecatalog.AllScope,
						Namespace: defaultNamespace,
					}))
				})
				It("Bubbles up errors from RetrievePlanByClassAndName", func() {
					errMsg := "error: too much sugar in coffee"
					fakeSDK.RetrievePlanByClassAndNameReturns(nil, errors.New(errMsg))

					err := cmd.Run()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(errMsg))
					Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(0))
					Expect(fakeSDK.RetrievePlanByClassAndNameCallCount()).To(Equal(1))
					Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(0))
					Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(0))
				})
			})
		})
		Context("When a kube name is provided", func() {
			BeforeEach(func() {
				cmd.LookupByKubeName = true
				cmd.KubeName = "csp-123"
			})
			It("Calls the pkg/svcat libs RetrievePlanByID/RetrieveClassByID with all scope and current namespace", func() {
				fakeSDK.RetrievePlanByIDReturns(clusterServicePlan, nil)
				fakeSDK.RetrieveClassByIDReturns(clusterServiceClass, nil)

				err := cmd.Run()

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrievePlanByClassAndNameCallCount()).To(Equal(0))
				Expect(fakeSDK.RetrievePlanByNameCallCount()).To(Equal(0))
				planName, scopeArg := fakeSDK.RetrievePlanByIDArgsForCall(0)
				Expect(planName).To(Equal(clusterServicePlan.Name))
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(1))
				classID, scopeArg := fakeSDK.RetrieveClassByIDArgsForCall(0)
				Expect(classID).To(Equal(clusterServicePlan.Spec.ClusterServiceClassRef.Name))
				Expect(scopeArg).To(Equal(servicecatalog.ScopeOptions{
					Scope:     servicecatalog.AllScope,
					Namespace: defaultNamespace,
				}))
			})
			It("Bubbles up errors from RetrievePlanByID", func() {
				errMsg := "error: too many pancakes"
				fakeSDK.RetrievePlanByIDReturns(nil, errors.New(errMsg))

				err := cmd.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(fakeSDK.RetrievePlanByIDCallCount()).To(Equal(1))
				Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(0))
			})
		})
	})
})
