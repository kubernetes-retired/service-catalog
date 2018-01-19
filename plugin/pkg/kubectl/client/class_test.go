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

var _ = Describe("Class", func() {
	var (
		client       *PluginClient
		err          error
		svcCatClient *fake.Clientset
		sc           *v1beta1.ClusterServiceClass
		sc2          *v1beta1.ClusterServiceClass
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())

		sc = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "foobar"}}
		sc2 = &v1beta1.ClusterServiceClass{ObjectMeta: metav1.ObjectMeta{Name: "barbaz"}}
		svcCatClient = fake.NewSimpleClientset(sc, sc2)
		client.ScClient = svcCatClient
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in class", func() {
			className := "foobar"
			class, err := client.GetClass(className)

			Expect(err).NotTo(HaveOccurred())
			Expect(class.Name).To(Equal(className))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(className))
		})
		It("Bubbles up errors", func() {
			className := "banana"

			class, err := client.GetClass(className)

			Expect(class).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceclasses")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(className))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			classes, err := client.ListClasses()

			Expect(err).NotTo(HaveOccurred())
			Expect(classes.Items).Should(ConsistOf(*sc, *sc2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "clusterserviceclasses", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			client.ScClient = badClient

			_, err := client.ListClasses()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "clusterserviceclasses")).To(BeTrue())
		})
	})
})
