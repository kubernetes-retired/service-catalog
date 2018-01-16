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

var _ = Describe("Plan", func() {
	var (
		client       *PluginClient
		err          error
		svcCatClient *fake.Clientset
		sp           *v1beta1.ClusterServicePlan
		sp2          *v1beta1.ClusterServicePlan
	)

	BeforeEach(func() {
		client, err = NewClient()
		Expect(err).NotTo(HaveOccurred())

		sp = &v1beta1.ClusterServicePlan{ObjectMeta: metav1.ObjectMeta{Name: "foobar"}}
		sp2 = &v1beta1.ClusterServicePlan{ObjectMeta: metav1.ObjectMeta{Name: "barbaz"}}
		svcCatClient = fake.NewSimpleClientset(sp, sp2)
		client.ScClient = svcCatClient
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in plan", func() {
			planName := "foobar"

			plan, err := client.GetPlan(planName)

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.Name).To(Equal(planName))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(planName))

		})
		It("Bubbles up errors", func() {
			planName := "not_real"

			plan, err := client.GetPlan(planName)

			Expect(plan).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("not found"))
			actions := svcCatClient.Actions()
			Expect(actions[0].Matches("get", "clusterserviceplans")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(planName))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			plans, err := client.ListPlans()

			Expect(err).NotTo(HaveOccurred())
			Expect(plans.Items).Should(ConsistOf(*sp, *sp2))
			Expect(svcCatClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
		It("Bubbles up errors", func() {
			badClient := &fake.Clientset{}
			errorMessage := "error retrieving list"
			badClient.AddReactor("list", "clusterserviceplans", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			client.ScClient = badClient
			_, err := client.ListPlans()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(badClient.Actions()[0].Matches("list", "clusterserviceplans")).To(BeTrue())
		})
	})
})
