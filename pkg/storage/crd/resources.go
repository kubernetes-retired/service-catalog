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
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func getCrd(kind Kind, resourcePlural ResourcePlural, hasNamespace bool) *crdv1beta1.CustomResourceDefinition {
	scope := crdv1beta1.NamespaceScoped
	if !hasNamespace {
		scope = crdv1beta1.ClusterScoped
	}
	return &crdv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourcePlural.String() + "." + CrdServiceCatalogDomain,
		},
		Spec: crdv1beta1.CustomResourceDefinitionSpec{
			Group:   CrdServiceCatalogResourceGroup,
			Version: CrdResourceVersion,
			Scope:   scope,
			Names: crdv1beta1.CustomResourceDefinitionNames{
				Plural:   resourcePlural.String(),
				Singular: strings.ToLower(kind.String()),
				Kind:     kind.String(),
			},
		},
	}
}

var serviceInstanceCRD = getCrd(ServiceInstanceKind, ServiceInstanceResourcePlural, true)
var serviceInstanceCredentialCRD = getCrd(ServiceInstanceCredentialKind, ServiceInstanceCredentialResourcePlural, true)
var serviceBrokerCRD = getCrd(ServiceBrokerKind, ServiceBrokerResourcePlural, false)
var serviceClassCRD = getCrd(ServiceClassKind, ServiceClassResourcePlural, false)
