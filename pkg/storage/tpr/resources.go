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

package tpr

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	tprKind    = "ThirdPartyResource"
	tprVersion = "v1alpha1"
)

// ServiceCatalogInstanceResource represents the API resource for the service instance third
// party resource
var ServiceCatalogInstanceResource = metav1.APIResource{
	Name:       ServiceCatalogInstanceKind.TPRName(),
	Namespaced: true,
}

// ServiceCatalogBindingResource represents the API resource for the service binding third
// party resource
var ServiceCatalogBindingResource = metav1.APIResource{
	Name:       ServiceCatalogBindingKind.TPRName(),
	Namespaced: true,
}

// ServiceCatalogBrokerResource represents the API resource for the service broker third
// party resource
var ServiceCatalogBrokerResource = metav1.APIResource{
	Name:       ServiceCatalogBrokerKind.TPRName(),
	Namespaced: true,
}

// ServiceCatalogServiceClassResource represents the API resource for the service class third
// party resource
var ServiceCatalogServiceClassResource = metav1.APIResource{
	// ServiceCatalogServiceClass is the kind, but TPRName converts it to 'serviceclass'. For now, just hard-code
	// it here
	Name:       "service-catalog-service-class",
	Namespaced: true,
}

// ServiceCatalogInstanceResource represents the API resource for the service instance third
// party resource
var serviceCatalogInstanceTPR = v1beta1.ThirdPartyResource{
	TypeMeta: metav1.TypeMeta{
		Kind:       tprKind,
		APIVersion: tprVersion,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: withGroupName(ServiceCatalogInstanceKind.TPRName()),
	},
	Versions: []v1beta1.APIVersion{
		{Name: tprVersion},
	},
}

// ServiceCatalogBindingResource represents the API resource for the service binding third
// party resource
var serviceCatalogBindingTPR = v1beta1.ThirdPartyResource{
	TypeMeta: metav1.TypeMeta{
		Kind:       tprKind,
		APIVersion: tprVersion,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: withGroupName(ServiceCatalogBindingKind.TPRName()),
	},
	Versions: []v1beta1.APIVersion{
		{Name: tprVersion},
	},
}

// ServiceCatalogBrokerResource represents the API resource for the service broker third
// party resource
var serviceCatalogBrokerTPR = v1beta1.ThirdPartyResource{
	TypeMeta: metav1.TypeMeta{
		Kind:       tprKind,
		APIVersion: tprVersion,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: withGroupName(ServiceCatalogBrokerKind.TPRName()),
	},
	Versions: []v1beta1.APIVersion{
		{Name: tprVersion},
	},
}

// ServiceCatalogServiceClassResource represents the API resource for the service class third
// party resource
var serviceCatalogServiceClassTPR = v1beta1.ThirdPartyResource{
	TypeMeta: metav1.TypeMeta{
		Kind:       tprKind,
		APIVersion: tprVersion,
	},
	// ServiceCatalogServiceClass is the kind, but TPRName converts it to 'serviceclass'. For now, just hard-code
	// it here
	ObjectMeta: metav1.ObjectMeta{
		Name: withGroupName(ServiceCatalogServiceClassKind.TPRName()),
	},
	Versions: []v1beta1.APIVersion{
		{Name: tprVersion},
	},
}
