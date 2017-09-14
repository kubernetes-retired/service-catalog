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

package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	scv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

const (
	// CrdServiceCatalogDomain is the Kubernetes domain for storing Service Catalog objects
	// Must be different from Service Catalog API server domain
	CrdServiceCatalogDomain = "crd.servicecatalog.k8s.io"
	// CrdServiceCatalogResourceGroup is the resource group for CRD storage
	CrdServiceCatalogResourceGroup = CrdServiceCatalogDomain
	// CrdResourceVersion is the CRD resource version
	CrdResourceVersion = "v1alpha1"
)

var (
	// SchemeGroupVersion is a group version for versioned CRD resources
	SchemeGroupVersion = schema.GroupVersion{
		Group:   CrdServiceCatalogResourceGroup,
		Version: CrdResourceVersion,
	}
	// SchemeBuilder is a scheme builder for versioned CRD resources
	SchemeBuilder = createSchemeBuilder()
	// AddToScheme registers versioned CRD resources in scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

func createSchemeBuilder() *runtime.SchemeBuilder {
	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	// Register conversions between versioned and internal types
	schemeBuilder.Register(scv1alpha1.RegisterConversions)
	return &schemeBuilder
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&scv1alpha1.ServiceBroker{},
		&scv1alpha1.ServiceBrokerList{},
		&scv1alpha1.ServiceClass{},
		&scv1alpha1.ServiceClassList{},
		&scv1alpha1.ServiceInstance{},
		&scv1alpha1.ServiceInstanceList{},
		&scv1alpha1.ServiceInstanceCredential{},
		&scv1alpha1.ServiceInstanceCredentialList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	scheme.AddKnownTypes(schema.GroupVersion{Version: "v1"}, &metav1.Status{})
	return nil
}
