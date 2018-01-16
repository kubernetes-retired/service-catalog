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

var _ = Describe("Binding", func() {
	var (
		client       *PluginClient
		err          error
		svcCatClient *fake.Clientset
		sb           *v1beta1.ServiceBinding
		sb2          *v1beta1.ServiceBinding
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())

		sb = &v1beta1.ServiceBinding{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "foobar_namespace"}}
		sb2 = &v1beta1.ServiceBinding{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "foobar_namespace"}}
		svcCatClient = fake.NewSimpleClientset(sb, sb2)
		client.ScClient = svcCatClient
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in binding and namespace", func() {
			bindingName := "foobar"
			namespace := "foobar_namespace"

			binding, err := client.GetBinding(bindingName, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(binding.Name).To(Equal(bindingName))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(bindingName))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(namespace))
		})
		It("Bubbles up errors", func() {
			bindingName := "not_a_real_binding"
			namespace := "foobar_namespace"

			_, err := client.GetBinding(bindingName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "servicebindings")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(bindingName))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(namespace))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			namespace := "foobar_namespace"

			bindings, err := client.ListBindings(namespace)

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
			client.ScClient = badClient
			namespace := "foobar_namespace"

			_, err := client.ListBindings(namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "servicebindings")).To(BeTrue())
		})
	})
})
