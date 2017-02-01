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

// ValidateBindingName is the validation function for Binding names.
var ValidateBindingName = apivalidation.NameIsDNSSubdomain

// ValidateBinding checks the fields of a Binding.
func ValidateBinding(binding *sc.Binding) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(&binding.ObjectMeta,
			false, /* namespace required */
			ValidateBindingName,
			field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateBindingSpec(&binding.Spec, field.NewPath("Spec"))...)

	// validate the status array
	// allErrs = append(allErrs, validateBindingStatus(&binding.Spec, field.NewPath("Status"))...)
	return allErrs
}

func validateBindingSpec(spec *sc.BindingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateBindingUpdate checks that when changing from an older binding to a newer binding is okay.
func ValidateBindingUpdate(new *sc.Binding, old *sc.Binding) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateBinding(new)...)
	allErrs = append(allErrs, ValidateBinding(old)...)
	// allErrs = append(allErrs, validateObjectMetaUpdate(new, old)...)
	// allErrs = append(allErrs, validateBindingSpecUpdate(new, old)...)
	// allErrs = append(allErrs, validateBindingStatusUpdate(new, old)...)
	return allErrs
}
