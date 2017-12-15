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

package validation

import (
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// ValidateClusterID validates a single clusterid
func ValidateClusterID(id *sc.ClusterID) field.ErrorList {
	allErrs := field.ErrorList{}

	// standard metadata validation
	metadataField := field.NewPath("metadata")
	allErrs = append(allErrs,
		validation.ValidateObjectMeta(&id.ObjectMeta,
			false, /* namespace required */
			validation.NameIsDNSSubdomain,
			metadataField)...)

	// only one with a specific name allowed.
	if id.Name != "cluster-id" {
		allErrs = append(allErrs, field.Invalid(metadataField.Child("name"), id.Name, "cluster-id name must be cluster-id"))
	}
	if id.ID == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("ID"), id.ID, "cluster-id ID must have content when setting"))
	}
	return allErrs
}

// ValidateClusterIDUpdate validates an update of a clusterid
func ValidateClusterIDUpdate(new *sc.ClusterID, old *sc.ClusterID) field.ErrorList {
	allErrs := field.ErrorList{}
	if new.ID != old.ID {
		allErrs = append(allErrs, field.Required(field.NewPath("spec"),
			"ID cannot change. New ID must equal Old ID."))
	}
	return allErrs
}
