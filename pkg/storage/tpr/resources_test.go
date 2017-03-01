package tpr

import (
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

//make sure each of TPR kinds are built with the correct structure
func TestTPRKinds(t *testing.T) {
	var serviceInstanceSample = v1beta1.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ThirdPartyResource",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: withGroupName("instance"),
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1alpha1"},
		},
	}

	var serviceBrokerSample = v1beta1.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ThirdPartyResource",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: withGroupName("broker"),
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1alpha1"},
		},
	}

	var serviceServiceClassSample = v1beta1.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ThirdPartyResource",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: withGroupName("service-class"),
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1alpha1"},
		},
	}

	var serviceBindingSample = v1beta1.ThirdPartyResource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ThirdPartyResource",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: withGroupName("binding"),
		},
		Versions: []v1beta1.APIVersion{
			{Name: "v1alpha1"},
		},
	}

	if !reflect.DeepEqual(serviceInstanceSample, serviceInstanceTPR) {
		t.Errorf("Unexpected Instance TPR structure")
	}

	if !reflect.DeepEqual(serviceBindingSample, serviceBindingTPR) {
		t.Errorf("Unexpected Broker TPR structure")
	}

	if !reflect.DeepEqual(serviceBrokerSample, serviceBrokerTPR) {
		t.Errorf("Unexpected Binding TPR structure")
	}

	if !reflect.DeepEqual(serviceServiceClassSample, serviceClassTPR) {
		t.Errorf("Unexpected Service Class TPR structure")
	}
}
