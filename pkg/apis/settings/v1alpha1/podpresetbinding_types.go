package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement the PodPresetBinding resource schema definition
// as a go struct.
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodPresetBindingSpec defines the desired state of PodPresetBinding
type PodPresetBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	BindingRef        *v1.ObjectReference `json:"bindingRef,omitempty"`
	PodPresetTemplate PodPreset           `json:"podPresetTemplate"`
}

// A new pod preset binding CRD is created which watches for service bindings to be ready. The pod preset bindings contains a reference to the formerly mentioned binding as well as a pod preset template. Once the service binding is ready, the pod preset is created which will contain an owner reference back to the pod preset binding.

// PodPresetBindingStatus defines the observed state of PodPresetBinding
type PodPresetBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodPresetBinding
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=podpresetbindings
type PodPresetBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodPresetBindingSpec   `json:"spec,omitempty"`
	Status PodPresetBindingStatus `json:"status,omitempty"`
}
