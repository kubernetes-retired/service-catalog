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

// ValidateInstanceName is the validation function for Instance names.
var ValidateInstanceName = apivalidation.NameIsDNSSubdomain

// ValidateInstance checks the fields of a Instance.
func ValidateInstance(instance *sc.Instance) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(
			&instance.ObjectMeta,
			false, /* namespace required */
			ValidateInstanceName,
			field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateInstanceSpec(&instance.Spec, field.NewPath("Spec"))...)
	// validate the status array
	// allErrs = append(allErrs, validateInstanceStatus(&instance.Spec, field.NewPath("Status"))...)
	return allErrs
}

func validateInstanceSpec(spec *sc.InstanceSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateInstanceUpdate checks that when changing from an older instance to a newer instance is okay.
func ValidateInstanceUpdate(new *sc.Instance, old *sc.Instance) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateInstance(new)...)
	allErrs = append(allErrs, ValidateInstance(old)...)
	// allErrs = append(allErrs, validateObjectMetaUpdate(new, old)...)
	// allErrs = append(allErrs, validateInstanceSpecUpdate(new, old)...)
	// allErrs = append(allErrs, validateInstanceStatusUpdate(new, old)...)
	return allErrs
}
