/*
Copyright 2019 The Kubernetes Authors.

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

package cleaner

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apimachinaryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestFinalizerCleaner_RemoveFinalizers(t *testing.T) {
	// Given
	fakeClisc := scfake.NewSimpleClientset(newTestCR()...)
	clr := NewFinalizerCleaner(fakeClisc)

	// When
	assert.NoError(t, clr.RemoveFinalizers())

	// Then
	listCsb, err := fakeClisc.ServicecatalogV1beta1().ClusterServiceBrokers().List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listCsb.Items {
		assert.Empty(t, item.Finalizers)
	}

	listSb, err := fakeClisc.ServicecatalogV1beta1().ServiceBrokers(v1.NamespaceAll).List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listSb.Items {
		assert.Empty(t, item.Finalizers)
	}

	listCsc, err := fakeClisc.ServicecatalogV1beta1().ClusterServiceClasses().List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listCsc.Items {
		assert.Empty(t, item.Finalizers)
	}

	listSc, err := fakeClisc.ServicecatalogV1beta1().ServiceClasses(v1.NamespaceAll).List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listSc.Items {
		assert.Empty(t, item.Finalizers)
	}

	listCsp, err := fakeClisc.ServicecatalogV1beta1().ClusterServicePlans().List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listCsp.Items {
		assert.Empty(t, item.Finalizers)
	}

	listSp, err := fakeClisc.ServicecatalogV1beta1().ServicePlans(v1.NamespaceAll).List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listSp.Items {
		assert.Empty(t, item.Finalizers)
	}

	listI, err := fakeClisc.ServicecatalogV1beta1().ServiceInstances(v1.NamespaceAll).List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listI.Items {
		assert.Empty(t, item.Finalizers)
	}

	listB, err := fakeClisc.ServicecatalogV1beta1().ServiceBindings(v1.NamespaceAll).List(apimachinaryv1.ListOptions{})
	assert.NoError(t, err)
	for _, item := range listB.Items {
		assert.Empty(t, item.Finalizers)
	}
}

func newTestCR() []runtime.Object {
	return []runtime.Object{
		&v1beta1.ServiceBroker{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ClusterServiceBroker{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ServiceClass{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ClusterServiceClass{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ServicePlan{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ClusterServicePlan{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
		&v1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{v1beta1.FinalizerServiceCatalog},
			},
		},
	}
}
