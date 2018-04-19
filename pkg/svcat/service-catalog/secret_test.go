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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"

	. "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Secret", func() {
	var (
		sdk            *SDK
		k8sClient      *k8sfake.Clientset
		svcCatClient   *fake.Clientset
		readyBinding   *v1beta1.ServiceBinding
		unreadyBinding *v1beta1.ServiceBinding
		boundSecret    *corev1.Secret
	)

	BeforeEach(func() {
		readyBinding = &v1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foobar",
				Namespace: "foobar_namespace"},
			Spec: v1beta1.ServiceBindingSpec{
				SecretName: "mysecret",
			},
			Status: v1beta1.ServiceBindingStatus{
				Conditions: []v1beta1.ServiceBindingCondition{
					{
						Type:   v1beta1.ServiceBindingConditionReady,
						Status: v1beta1.ConditionTrue,
					},
				},
			},
		}
		unreadyBinding = &v1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "barbaz",
				Namespace: "foobar_namespace",
			},
			Spec: v1beta1.ServiceBindingSpec{
				SecretName: "missing-secret",
			},
		}
		boundSecret = &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "mysecret", Namespace: "foobar_namespace"}}
		svcCatClient = fake.NewSimpleClientset(readyBinding, unreadyBinding)
		k8sClient = k8sfake.NewSimpleClientset(boundSecret)
		sdk = &SDK{
			K8sClient:            k8sClient,
			ServiceCatalogClient: svcCatClient,
		}
	})

	Describe("RetrieveSecretByBinding", func() {
		It("Gets the secret", func() {
			secret, err := sdk.RetrieveSecretByBinding(readyBinding)

			Expect(err).NotTo(HaveOccurred())
			Expect(secret).To(Equal(boundSecret))

			actions := k8sClient.Actions()
			Expect(actions[0].Matches("get", "secrets")).To(BeTrue())
			Expect(actions[0].(testing.GetActionImpl).Name).To(Equal(boundSecret.Name))
			Expect(actions[0].(testing.GetActionImpl).Namespace).To(Equal(boundSecret.Namespace))
		})
		It("Ignores missing secrets when the binding is not ready", func() {
			secret, err := sdk.RetrieveSecretByBinding(unreadyBinding)

			Expect(err).NotTo(HaveOccurred())
			Expect(secret).To(BeNil())
		})
		It("Bubbles up errors", func() {
			badClient := &k8sfake.Clientset{}
			errorMessage := "resource not found"
			badClient.AddReactor("get", "secrets", func(action testing.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf(errorMessage)
			})
			sdk.K8sClient = badClient

			secret, err := sdk.RetrieveSecretByBinding(readyBinding)

			Expect(err).To(HaveOccurred())
			Expect(secret).To(BeNil())
			Expect(err.Error()).Should(ContainSubstring("not found"))
		})
	})

})
