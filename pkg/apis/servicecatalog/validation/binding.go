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
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	apivalidation "k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

// validateBindingName is the validation function for Binding names.
var validateBindingName = apivalidation.NameIsDNSSubdomain

// ValidateBinding validates a Binding and returns a list of errors.
func ValidateBinding(binding *sc.Binding) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = appendToErrListAndLog(
		allErrs,
		apivalidation.ValidateObjectMeta(
			&binding.ObjectMeta,
			true, /*namespace*/
			validateBindingName,
			field.NewPath("metadata"),
		)...,
	)
	allErrs = appendToErrListAndLog(
		allErrs,
		validateBindingSpec(&binding.Spec, field.NewPath("Spec"))...,
	)

	return allErrs
}

func validateBindingSpec(spec *sc.BindingSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, msg := range validateInstanceName(spec.InstanceRef.Name, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("instanceRef", "name"), spec.InstanceRef.Name, msg))
	}

	for _, msg := range apivalidation.ValidateSecretName(spec.SecretName, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("secretName"), spec.SecretName, msg))
	}

	return allErrs
}

// ValidateBindingUpdate checks that when changing from an older binding to a newer binding is okay.
func ValidateBindingUpdate(new *sc.Binding, old *sc.Binding) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateBinding(new)...)
	allErrs = append(allErrs, ValidateBinding(old)...)
	return allErrs
}

// ValidateBindingStatusUpdate checks that when changing from an older binding to a newer binding is okay.
func ValidateBindingStatusUpdate(new *sc.Binding, old *sc.Binding) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateBindingUpdate(new, old)...)
	return allErrs
}
