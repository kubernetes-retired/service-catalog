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
	"sync"
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

var _ = Describe("Binding", func() {
	var (
		sdk          *SDK
		svcCatClient *fake.Clientset
		sb           *v1beta1.ServiceBinding
		sb2          *v1beta1.ServiceBinding
	)

	BeforeEach(func() {
		sb = &v1beta1.ServiceBinding{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "foobar_namespace"}}
		sb2 = &v1beta1.ServiceBinding{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "foobar_namespace"}}
		svcCatClient = fake.NewSimpleClientset(sb, sb2)
		sdk = &SDK{
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("BindingHasStatus", func() {
		It("Returns false when conditions is empty", func() {
			binding := &v1beta1.ServiceBinding{}
			result := sdk.BindingHasStatus(binding, v1beta1.ServiceBindingConditionReady)
			Expect(result).Should(BeFalse())
		})
		It("Returns false when condition status is false", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionFailed,
							Status: v1beta1.ConditionFalse,
						},
					},
				},
			}
			result := sdk.BindingHasStatus(binding, v1beta1.ServiceBindingConditionFailed)
			Expect(result).Should(BeFalse())
		})
		It("Returns true when conditions contain ready", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionReady,
							Status: v1beta1.ConditionTrue,
						},
					},
				},
			}
			result := sdk.BindingHasStatus(binding, v1beta1.ServiceBindingConditionReady)
			Expect(result).Should(BeTrue())
		})
	})

	Describe("IsBindingReady", func() {
		It("Returns false when conditions is empty", func() {
			binding := &v1beta1.ServiceBinding{}
			result := sdk.IsBindingReady(binding)
			Expect(result).Should(BeFalse())
		})
		It("Returns false when ready condition status is false", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionReady,
							Status: v1beta1.ConditionFalse,
						},
					},
				},
			}
			result := sdk.IsBindingReady(binding)
			Expect(result).Should(BeFalse())
		})
		It("Returns true when ready condition status is true", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionReady,
							Status: v1beta1.ConditionTrue,
						},
					},
				},
			}
			result := sdk.IsBindingReady(binding)
			Expect(result).Should(BeTrue())
		})
	})

	Describe("IsBindingFailed", func() {
		It("Returns false when conditions is empty", func() {
			binding := &v1beta1.ServiceBinding{}
			result := sdk.IsBindingFailed(binding)
			Expect(result).Should(BeFalse())
		})
		It("Returns false when failed condition status is false", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionFailed,
							Status: v1beta1.ConditionFalse,
						},
					},
				},
			}
			result := sdk.IsBindingFailed(binding)
			Expect(result).Should(BeFalse())
		})
		It("Returns true when failed condition status is true", func() {
			binding := &v1beta1.ServiceBinding{
				Status: v1beta1.ServiceBindingStatus{
					Conditions: []v1beta1.ServiceBindingCondition{
						{
							Type:   v1beta1.ServiceBindingConditionFailed,
							Status: v1beta1.ConditionTrue,
						},
					},
				},
			}
			result := sdk.IsBindingFailed(binding)
			Expect(result).Should(BeTrue())
		})
	})

	Describe("WaitForBinding", func() {
		It("Polls until the binding is ready", func() {
			timeout := 1 * time.Second

			readyClient := &fake.Clientset{}
			readyClient.AddReactor("get", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, sb, nil
			})
			sdk.ServiceCatalogClient = readyClient

			var wg sync.WaitGroup
			wg.Add(1)
			var binding *v1beta1.ServiceBinding
			var err error
			go func() {
				binding, err = sdk.WaitForBinding(sb.Namespace, sb.Name, time.Millisecond, &timeout)
				wg.Done()
			}()

			time.Sleep(500 * time.Millisecond)
			sb.Status = v1beta1.ServiceBindingStatus{
				Conditions: []v1beta1.ServiceBindingCondition{
					{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: v1beta1.ConditionTrue,
					},
				},
			}

			wg.Wait()

			condition := binding.Status.Conditions[0]
			Expect(condition.Type).To(Equal(v1beta1.ServiceBindingConditionReady))
			Expect(condition.Status).To(Equal(v1beta1.ConditionTrue))
		})
		It("Polls until the binding is failed", func() {
			timeout := 1 * time.Second

			readyClient := &fake.Clientset{}
			readyClient.AddReactor("get", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, sb, nil
			})
			sdk.ServiceCatalogClient = readyClient

			var wg sync.WaitGroup
			wg.Add(1)
			var binding *v1beta1.ServiceBinding
			var err error
			go func() {
				binding, err = sdk.WaitForBinding(sb.Namespace, sb.Name, time.Millisecond, &timeout)
				wg.Done()
			}()

			time.Sleep(500 * time.Millisecond)
			sb.Status = v1beta1.ServiceBindingStatus{
				Conditions: []v1beta1.ServiceBindingCondition{
					{
						Type:   v1beta1.ServiceBindingConditionFailed,
						Status: v1beta1.ConditionTrue,
					},
				},
			}

			wg.Wait()
			Expect(err).To(BeNil())
			condition := binding.Status.Conditions[0]
			Expect(condition.Type).To(Equal(v1beta1.ServiceBindingConditionFailed))
			Expect(condition.Status).To(Equal(v1beta1.ConditionTrue))
		})
		It("Polls until the async operation is complete", func() {
			timeout := 1 * time.Second

			sb.Status.AsyncOpInProgress = true
			sb.Status = v1beta1.ServiceBindingStatus{
				Conditions: []v1beta1.ServiceBindingCondition{
					{
						Type:   v1beta1.ServiceBindingConditionFailed,
						Status: v1beta1.ConditionTrue,
					},
				},
			}

			readyClient := &fake.Clientset{}
			readyClient.AddReactor("get", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, sb, nil
			})
			sdk.ServiceCatalogClient = readyClient

			var wg sync.WaitGroup
			wg.Add(1)
			var binding *v1beta1.ServiceBinding
			var err error
			go func() {
				binding, err = sdk.WaitForBinding(sb.Namespace, sb.Name, time.Millisecond, &timeout)
				wg.Done()
			}()

			time.Sleep(500 * time.Millisecond)
			sb.Status.AsyncOpInProgress = false

			wg.Wait()

			Expect(binding.Status.AsyncOpInProgress).To(BeFalse())
		})
	})

	Describe("RetrieveBinding", func() {
		It("Calls the generated v1beta1 Get method with the passed in binding and namespace", func() {
			binding, err := sdk.RetrieveBinding(sb.Namespace, sb.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(binding).To(Equal(sb))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(sb.Name))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(sb.Namespace))
		})
		It("Bubbles up errors", func() {
			fakeName := "not_a_real_binding"

			_, err := sdk.RetrieveBinding(sb.Namespace, fakeName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(fakeName))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(sb.Namespace))
		})
	})

	Describe("RetrieveBindings", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			bindings, err := sdk.RetrieveBindings(sb.Namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(bindings.Items).Should(ConsistOf(*sb, *sb2))
			Expect(svcCatClient.Actions()[0].Matches("list", "servicebindings")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			bindings, err := sdk.RetrieveBindings(sb.Namespace)

			Expect(bindings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "servicebindings")).To(BeTrue())
		})
	})

	Describe("RetrieveBindingsByInstance", func() {
		It("Calls the generated v1beta1 List method on the provided instance's namespace", func() {
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "apple_instance", Namespace: sb.Namespace}}
			sb.Spec.ServiceInstanceRef.Name = si.Name
			svcCatClient = fake.NewSimpleClientset(sb, sb2)
			sdk = &SDK{
				ServiceCatalogClient: svcCatClient,
			}

			bindings, err := sdk.RetrieveBindingsByInstance(si)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindings).To(ConsistOf(*sb))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("list", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.ListActionImpl).Namespace).To(Equal(si.Namespace))
		})

		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "apple_instance", Namespace: "not_real_namespace"}}
			bindings, err := sdk.RetrieveBindingsByInstance(si)

			Expect(bindings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "servicebindings")).To(BeTrue())
		})
	})

	Describe("Bind", func() {
		It("Calls the generated v1beta1 method to create a binding", func() {
			bindingNamespace := "banana_namespace"
			bindingName := "banana_binding"
			externalID := "banana_external_id"
			instanceName := "banana_instance"
			secret := "banana_secret"
			binding, err := sdk.Bind(bindingNamespace, bindingName, externalID, instanceName, secret, map[string]string{}, map[string]string{})

			Expect(err).NotTo(HaveOccurred())
			Expect(binding).NotTo(BeNil())
			Expect(binding.ObjectMeta.Namespace).To(Equal(bindingNamespace))
			Expect(binding.ObjectMeta.Name).To(Equal(bindingName))
			Expect(binding.Spec.ServiceInstanceRef.Name).To(Equal(instanceName))
			Expect(binding.Spec.SecretName).To(Equal(secret))
			Expect(binding.Spec.ExternalID).To(Equal(externalID))
			Expect(svcCatClient.Actions()[0].Matches("create", "servicebindings")).To(BeTrue())
		})

		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("create", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			bindingNamespace := "banana_namespace"
			bindingName := "banana_binding"
			instanceName := "banana_instance"
			binding, err := sdk.Bind(bindingNamespace, bindingName, "", instanceName, "banana_secret", map[string]string{}, map[string]string{})

			Expect(binding).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("create", "servicebindings")).To(BeTrue())
		})
	})

	Describe("Unbind", func() {
		It("Calls the generated v1beta1 method to delete a binding", func() {
			instanceNamespace := sb.Namespace
			instanceName := "apple_instance"
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace}}
			sb.Spec.ServiceInstanceRef.Name = si.Name
			linkedClient := fake.NewSimpleClientset(sb, sb2, si)
			sdk = &SDK{
				ServiceCatalogClient: linkedClient,
			}

			deleted, err := sdk.Unbind(instanceNamespace, instanceName)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(deleted)).To(Equal(1))
			Expect(linkedClient.Actions()[0].Matches("get", "serviceinstances")).To(BeTrue())
			Expect(linkedClient.Actions()[1].Matches("list", "servicebindings")).To(BeTrue())
			Expect(linkedClient.Actions()[2].Matches("delete", "servicebindings")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			instanceNamespace := sb.Namespace
			instanceName := "apple_instance"
			errorMessage := "error deleting binding"
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace}}
			sb.Spec.ServiceInstanceRef.Name = si.Name

			badClient := &fake.Clientset{}
			badClient.AddReactor("get", "serviceinstances", func(action testing.Action) (bool, runtime.Object, error) {
				return true, si, nil
			})
			badClient.AddReactor("list", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceBindingList{Items: []v1beta1.ServiceBinding{*sb}}, nil
			})
			badClient.AddReactor("delete", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk = &SDK{
				ServiceCatalogClient: badClient,
			}

			deleted, err := sdk.Unbind(instanceNamespace, instanceName)

			Expect(err).To(HaveOccurred())
			Expect(len(deleted)).To(Equal(0))
			Expect(err.Error()).To(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("get", "serviceinstances")).To(BeTrue())
			Expect(badClient.Actions()[1].Matches("list", "servicebindings")).To(BeTrue())
			Expect(badClient.Actions()[2].Matches("delete", "servicebindings")).To(BeTrue())
		})
		It("Checks to see if the binding's instance exists before attempting to delete the binding", func() {
			instanceNamespace := sb.Namespace
			instanceName := "apple_instance"
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace}}
			sb.Spec.ServiceInstanceRef.Name = si.Name
			noInstanceClient := fake.NewSimpleClientset(sb, sb2)
			sdk = &SDK{
				ServiceCatalogClient: noInstanceClient,
			}

			deleted, err := sdk.Unbind(instanceNamespace, instanceName)

			Expect(err).To(HaveOccurred())
			Expect(len(deleted)).To(Equal(0))
			Expect(err.Error()).To(ContainSubstring("unable to get instance"))
			Expect(noInstanceClient.Actions()[0].Matches("get", "serviceinstances")).To(BeTrue())
		})
		It("Returns only successfully deleted bindings", func() {
			instanceNamespace := sb.Namespace
			instanceName := "apple_instance"
			errorMessage := "error deleting binding"
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace}}
			sb.Spec.ServiceInstanceRef.Name = si.Name
			sb2.Spec.ServiceInstanceRef.Name = si.Name
			badClient := &fake.Clientset{}
			badClient.AddReactor("get", "serviceinstances", func(action testing.Action) (bool, runtime.Object, error) {
				return true, si, nil
			})
			badClient.AddReactor("list", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, &v1beta1.ServiceBindingList{Items: []v1beta1.ServiceBinding{*sb, *sb2}}, nil
			})
			badClient.AddReactor("delete", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				da, ok := action.(testing.DeleteAction)
				if !ok {
					return true, nil, fmt.Errorf("internal error occurred")
				}
				switch da.GetName() {
				case sb2.Name:
					return true, nil, fmt.Errorf(errorMessage)
				default:
					return true, sb, nil
				}
			})
			sdk = &SDK{
				ServiceCatalogClient: badClient,
			}

			deleted, err := sdk.Unbind(instanceNamespace, instanceName)

			Expect(err).To(HaveOccurred())
			Expect(len(deleted)).To(Equal(1))
			Expect(err.Error()).To(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("get", "serviceinstances")).To(BeTrue())
			Expect(badClient.Actions()[1].Matches("list", "servicebindings")).To(BeTrue())
			Expect(badClient.Actions()[2].Matches("delete", "servicebindings")).To(BeTrue())
		})
	})
})
