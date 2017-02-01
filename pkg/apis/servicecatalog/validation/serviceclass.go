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
	apivalidation "k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/util/validation/field"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// ValidateServiceClassName is the validation function for ServiceClass names.
var ValidateServiceClassName = apivalidation.NameIsDNSSubdomain

// ValidateServiceclass makes sure a serviceclass object is okay.
func ValidateServiceclass(serviceclass *sc.ServiceClass) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(
			&serviceclass.ObjectMeta,
			false, /* namespace required */
			ValidateServiceClassName,
			field.NewPath("metadata"))...)

	return allErrs
}

// ValidateServiceclassUpdate checks that when changing from an older
// serviceclass to a newer serviceclass is okay.
func ValidateServiceclassUpdate(new *sc.ServiceClass, old *sc.ServiceClass) field.ErrorList {
	allErrs := field.ErrorList{}
	// should each individual serviceclass validate successfully before validating changes?
	allErrs = append(allErrs, ValidateServiceclass(new)...)
	allErrs = append(allErrs, ValidateServiceclass(old)...)
	// allErrs = append(allErrs, validateObjectMetaUpdate(new, old)...)
	// allErrs = append(allErrs, validateServiceclassFieldsUpdate(new, old)...)
	return allErrs
}
