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

var _ = Describe("Instances", func() {
	var (
		client       *PluginClient
		err          error
		svcCatClient *fake.Clientset
		si           *v1beta1.ServiceInstance
		si2          *v1beta1.ServiceInstance
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())
		si = &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "foobar", Namespace: "foobar_namespace"}}
		si2 = &v1beta1.ServiceInstance{ObjectMeta: metav1.ObjectMeta{Name: "barbaz", Namespace: "foobar_namespace"}}
		svcCatClient = fake.NewSimpleClientset(si, si2)
		client.ScClient = svcCatClient
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in instance and namespace", func() {
			instanceName := "foobar"
			namespace := "foobar_namespace"

			instance, err := client.GetInstance(instanceName, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Name).To(Equal(instanceName))

			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "serviceinstances")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(instanceName))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(namespace))
		})
		It("Bubbles up errors", func() {
			instanceName := "not_real"
			namespace := "foobar_namespace"

			_, err := client.GetInstance(instanceName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "serviceinstances")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(instanceName))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(namespace))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			namespace := "foobar_namespace"

			instances, err := client.ListInstances(namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(instances.Items).Should(ConsistOf(*si, *si2))
			Expect(svcCatClient.Actions()[0].Matches("list", "serviceinstances")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "serviceinstances", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			client.ScClient = badClient
			namespace := "foobar_namespace"

			_, err := client.ListInstances(namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "serviceinstances")).To(BeTrue())
		})
	})
})
