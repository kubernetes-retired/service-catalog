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

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"

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
			sb.Spec.InstanceRef.Name = si.Name
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
			Expect(binding.Spec.InstanceRef.Name).To(Equal(instanceName))
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
			sb.Spec.InstanceRef.Name = si.Name
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
			sb.Spec.InstanceRef.Name = si.Name

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
			sb.Spec.InstanceRef.Name = si.Name
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
			sb.Spec.InstanceRef.Name = si.Name
			sb2.Spec.InstanceRef.Name = si.Name
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

	Describe("DeleteBindings", func() {
		It("Calls the generated v1beta1 delete method for every binding", func() {
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "myinstance", Namespace: sb.Namespace}}
			sb.Spec.InstanceRef.Name = si.Name
			sb2.Spec.InstanceRef.Name = si.Name
			client := fake.NewSimpleClientset(sb, sb2, si)
			sdk = &SDK{
				ServiceCatalogClient: client,
			}

			bindingsToDelete := []types.NamespacedName{
				{Namespace: sb.Namespace, Name: sb.Name},
				{Namespace: sb2.Namespace, Name: sb2.Name},
			}

			deleted, err := sdk.DeleteBindings(bindingsToDelete)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(deleted)).To(Equal(2))

			deletedBindings := make(map[string]string) // maps name to namespace
			for i := 0; i < 2; i++ {
				action := client.Actions()[i]
				Expect(action.Matches("delete", "servicebindings")).To(BeTrue())
				deleteAction := action.(testing.DeleteAction)
				deletedBindings[deleteAction.GetName()] = deleteAction.GetNamespace()
			}

			Expect(deletedBindings[sb.Name]).To(Equal(sb.Namespace))
			Expect(deletedBindings[sb2.Name]).To(Equal(sb2.Namespace))
		})
	})

	Describe("RemoveFinalizerForBinding", func() {
		It("Calls the generated v1beta1 put method with the passed in binding", func() {
			err := sdk.RemoveFinalizerForBinding(types.NamespacedName{
				Namespace: sb.Namespace,
				Name:      sb.Name,
			})
			Expect(err).NotTo(HaveOccurred())

			actions := svcCatClient.Actions()
			// The first action should be a get call because RemoveFinalizerForBinding calls RetrieveBinding to retrieve the ServiceBinding object
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(sb.Name))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(sb.Namespace))
			Expect(actions[1].Matches("update", "servicebindings")).To(BeTrue())
			Expect(actions[1].(testing.UpdateActionImpl).Object.(*v1beta1.ServiceBinding).ObjectMeta.Name).To(Equal(sb.ObjectMeta.Name))
			Expect(actions[1].(testing.UpdateActionImpl).Object.(*v1beta1.ServiceBinding).ObjectMeta.Namespace).To(Equal(sb.ObjectMeta.Namespace))
		})
		It("Bubbles up errors", func() {
			fakeNamespacedName := types.NamespacedName{
				Namespace: "not_a_real_namespace",
				Name:      "not_a_real_binding",
			}
			err := sdk.RemoveFinalizerForBinding(fakeNamespacedName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			// The only expected action is the get call to retrive binding, because if the binding is not found, no delete call will be made
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(fakeNamespacedName.Name))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(fakeNamespacedName.Namespace))
		})
	})

	Describe("RemoveFinalizerForBindings", func() {
		It("Calls the generated v1beta1 put method with all bindings", func() {
			bindings, err := sdk.RemoveFinalizerForBindings([]types.NamespacedName{
				{Namespace: sb.Namespace, Name: sb.Name},
				{Namespace: sb2.Namespace, Name: sb2.Name},
			})

			expectedBindingsToDelete := []types.NamespacedName{
				{Namespace: sb.Namespace, Name: sb.Name},
				{Namespace: sb2.Namespace, Name: sb2.Name},
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(bindings).Should(ConsistOf(expectedBindingsToDelete[0], expectedBindingsToDelete[1]))
			// Two of the actions should be get calls and the other two should be update calls
			numGetCalls := 0
			numUpdateCalls := 0
			for _, action := range svcCatClient.Actions() {
				if action.Matches("get", "servicebindings") {
					numGetCalls++
				} else if action.Matches("update", "servicebindings") {
					numUpdateCalls++
				}
			}
			Expect(numGetCalls).To(Equal(2))
			Expect(numUpdateCalls).To(Equal(2))
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error updating bindings"
			badClient.AddReactor("update", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = badClient

			bindings, err := sdk.RemoveFinalizerForBindings([]types.NamespacedName{
				{Namespace: sb.Namespace, Name: sb.Name},
				{Namespace: sb2.Namespace, Name: sb2.Name},
			})

			Expect(bindings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
		})
	})

	Describe("RemoveBindingFinalizerByInstance", func() {
		It("Deletes all bindings of the instance", func() {
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "myinstance", Namespace: sb.Namespace}}
			sb.Spec.InstanceRef.Name = si.Name
			sb2.Spec.InstanceRef.Name = si.Name
			client := fake.NewSimpleClientset(sb, sb2, si)
			sdk.ServiceCatalogClient = client

			bindings, err := sdk.RemoveBindingFinalizerByInstance(si.Namespace, si.Name)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindings).To(ConsistOf(
				types.NamespacedName{Namespace: sb.Namespace, Name: sb.Name},
				types.NamespacedName{Namespace: sb2.Namespace, Name: sb2.Name},
			))
		})

		It("Bubbles up errors", func() {
			si := &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "myinstance", Namespace: sb.Namespace}}
			sb.Spec.InstanceRef.Name = si.Name
			sb2.Spec.InstanceRef.Name = si.Name
			client := fake.NewSimpleClientset(sb, sb2, si)
			errorMessage := "error updating bindings"
			client.PrependReactor("update", "servicebindings", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.ServiceCatalogClient = client

			bindings, err := sdk.RemoveBindingFinalizerByInstance(si.Namespace, si.Name)

			Expect(bindings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
		})
	})
})
