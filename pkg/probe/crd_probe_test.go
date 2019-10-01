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

package probe

import (
	"github.com/stretchr/testify/assert"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestCRDProbe_Check(t *testing.T) {
	// Given
	fakeCliext := apiextfake.NewSimpleClientset(newTestCRD()...)
	probe := NewCRDProbe(fakeCliext, 0)

	// Then
	assert.NoError(t, probe.Check(nil))
}

func TestCRDProbe_IsReady(t *testing.T) {
	// Given
	fakeCliext := apiextfake.NewSimpleClientset(newTestCRD()...)
	probe := NewCRDProbe(fakeCliext, 0)

	// Then
	ready, err := probe.IsReady()
	assert.NoError(t, err)
	assert.True(t, ready)
}

func TestCRDProbe_IsReadyWithDelay(t *testing.T) {
	// Given
	fakeCliext := apiextfake.NewSimpleClientset(newTestCRDNotReady()...)
	probe := NewCRDProbe(fakeCliext, 2)

	// Then
	assert.NoError(t, probe.Check(nil))
	assert.NoError(t, probe.Check(nil))
	assert.EqualError(t, probe.Check(nil), "CRDs are not ready")
}

func TestCRDProbe_CheckFailed(t *testing.T) {
	// Given
	fakeCliext := apiextfake.NewSimpleClientset(newTestCRDNotReady()...)
	probe := NewCRDProbe(fakeCliext, 0)

	// Then
	assert.EqualError(t, probe.Check(nil), "CRDs are not ready")
}

func newTestCRD() []runtime.Object {
	return []runtime.Object{
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "NotServiceCatalogCRD",
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceBroker,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServiceBroker,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceClass,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServiceClass,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServicePlan,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServicePlan,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceInstance,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceBinding,
				Labels: map[string]string{"svcat": "true"},
			},
			Status: extv1beta1.CustomResourceDefinitionStatus{
				Conditions: []extv1beta1.CustomResourceDefinitionCondition{
					{
						Type:   extv1beta1.Established,
						Status: "True",
					},
				},
			},
		},
	}
}

func newTestCRDNotReady() []runtime.Object {
	return []runtime.Object{
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceBroker,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServiceBroker,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceClass,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServiceClass,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServicePlan,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServicePlan,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceInstance,
				Labels: map[string]string{"svcat": "true"},
			},
		},
		&extv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ServiceBinding,
				Labels: map[string]string{"svcat": "true"},
			},
		},
	}
}
