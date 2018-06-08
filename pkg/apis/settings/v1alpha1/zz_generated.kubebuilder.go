package v1alpha1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "settings.servicecatalog.k8s.io", Version: "v1alpha1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&PodPreset{},
		&PodPresetList{},
		&PodPresetBinding{},
		&PodPresetBindingList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PodPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodPreset `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PodPresetBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodPresetBinding `json:"items"`
}

// CRD Generation
func getFloat(f float64) *float64 {
	return &f
}

var (
	// Define CRDs for resources
	PodPresetCRD = v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "podpresets.settings.servicecatalog.k8s.io",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "settings.servicecatalog.k8s.io",
			Version: "v1alpha1",
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:   "PodPreset",
				Plural: "podpresets",
			},
			Scope: "Namespaced",
			Validation: &v1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]v1beta1.JSONSchemaProps{
						"apiVersion": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"kind": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"metadata": v1beta1.JSONSchemaProps{
							Type: "object",
						},
						"spec": v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"env": v1beta1.JSONSchemaProps{
									Type: "array",
									Items: &v1beta1.JSONSchemaPropsOrArray{
										Schema: &v1beta1.JSONSchemaProps{
											Type:       "object",
											Properties: map[string]v1beta1.JSONSchemaProps{},
										},
									},
								},
								"envFrom": v1beta1.JSONSchemaProps{
									Type: "array",
									Items: &v1beta1.JSONSchemaPropsOrArray{
										Schema: &v1beta1.JSONSchemaProps{
											Type:       "object",
											Properties: map[string]v1beta1.JSONSchemaProps{},
										},
									},
								},
								"selector": v1beta1.JSONSchemaProps{
									Type:       "object",
									Properties: map[string]v1beta1.JSONSchemaProps{},
								},
								"volumeMounts": v1beta1.JSONSchemaProps{
									Type: "array",
									Items: &v1beta1.JSONSchemaPropsOrArray{
										Schema: &v1beta1.JSONSchemaProps{
											Type:       "object",
											Properties: map[string]v1beta1.JSONSchemaProps{},
										},
									},
								},
								"volumes": v1beta1.JSONSchemaProps{
									Type: "array",
									Items: &v1beta1.JSONSchemaPropsOrArray{
										Schema: &v1beta1.JSONSchemaProps{
											Type:       "object",
											Properties: map[string]v1beta1.JSONSchemaProps{},
										},
									},
								},
							},
						},
						"status": v1beta1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1beta1.JSONSchemaProps{},
						},
					},
				},
			},
		},
	}
	// Define CRDs for resources
	PodPresetBindingCRD = v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "podpresetbindings.settings.servicecatalog.k8s.io",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "settings.servicecatalog.k8s.io",
			Version: "v1alpha1",
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:   "PodPresetBinding",
				Plural: "podpresetbindings",
			},
			Scope: "Namespaced",
			Validation: &v1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]v1beta1.JSONSchemaProps{
						"apiVersion": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"kind": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"metadata": v1beta1.JSONSchemaProps{
							Type: "object",
						},
						"spec": v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"apiVersion": v1beta1.JSONSchemaProps{
									Type: "string",
								},
								"bindingRef": v1beta1.JSONSchemaProps{
									Type:       "object",
									Properties: map[string]v1beta1.JSONSchemaProps{},
								},
								"kind": v1beta1.JSONSchemaProps{
									Type: "string",
								},
								"metadata": v1beta1.JSONSchemaProps{
									Type: "object",
								},
								"podPresetTemplate": v1beta1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"apiVersion": v1beta1.JSONSchemaProps{
											Type: "string",
										},
										"kind": v1beta1.JSONSchemaProps{
											Type: "string",
										},
										"metadata": v1beta1.JSONSchemaProps{
											Type: "object",
										},
										"spec": v1beta1.JSONSchemaProps{
											Type: "object",
											Properties: map[string]v1beta1.JSONSchemaProps{
												"env": v1beta1.JSONSchemaProps{
													Type: "array",
													Items: &v1beta1.JSONSchemaPropsOrArray{
														Schema: &v1beta1.JSONSchemaProps{
															Type:       "object",
															Properties: map[string]v1beta1.JSONSchemaProps{},
														},
													},
												},
												"envFrom": v1beta1.JSONSchemaProps{
													Type: "array",
													Items: &v1beta1.JSONSchemaPropsOrArray{
														Schema: &v1beta1.JSONSchemaProps{
															Type:       "object",
															Properties: map[string]v1beta1.JSONSchemaProps{},
														},
													},
												},
												"selector": v1beta1.JSONSchemaProps{
													Type:       "object",
													Properties: map[string]v1beta1.JSONSchemaProps{},
												},
												"volumeMounts": v1beta1.JSONSchemaProps{
													Type: "array",
													Items: &v1beta1.JSONSchemaPropsOrArray{
														Schema: &v1beta1.JSONSchemaProps{
															Type:       "object",
															Properties: map[string]v1beta1.JSONSchemaProps{},
														},
													},
												},
												"volumes": v1beta1.JSONSchemaProps{
													Type: "array",
													Items: &v1beta1.JSONSchemaPropsOrArray{
														Schema: &v1beta1.JSONSchemaProps{
															Type:       "object",
															Properties: map[string]v1beta1.JSONSchemaProps{},
														},
													},
												},
											},
										},
										"status": v1beta1.JSONSchemaProps{
											Type:       "object",
											Properties: map[string]v1beta1.JSONSchemaProps{},
										},
									},
								},
							},
						},
						"status": v1beta1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1beta1.JSONSchemaProps{},
						},
					},
				},
			},
		},
	}
)
