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
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/util/validation/field"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// validateServiceClassName is the validation function for ServiceClass names.
var validateServiceClassName = apivalidation.NameIsDNSSubdomain

// validateServicePlanName is the validation function for ServicePlan names.
var validateServicePlanName = apivalidation.NameIsDNSLabel

// validateOSBGuid is the validation function for OSB GUIDs.  We generate
// GUIDs for Instances and Bindings, but for ServiceClass and ServicePlan,
// they are part of the payload returned from the Broker.
//
// TODO: This might be looser than it should be, but it seems like a
// reasonable approximation for now.  The OSB spec does not provide specifics
// about the format of ID fields in the API.
var validateOSBGuid = apivalidation.NameIsDNSLabel

// ValidateServiceClass validates a ServiceClass and returns a list of errors.
func ValidateServiceClass(serviceclass *sc.ServiceClass) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(
			&serviceclass.ObjectMeta,
			false, /* namespace required */
			validateServiceClassName,
			field.NewPath("metadata"))...)

	if "" == serviceclass.BrokerName {
		allErrs = append(allErrs, field.Required(field.NewPath("brokerName"), "brokerName is required"))
	}

	if "" == serviceclass.OSBGUID {
		allErrs = append(allErrs, field.Required(field.NewPath("osbGuid"), "osbGuid is required"))
	}

	for _, msg := range validateOSBGuid(serviceclass.OSBGUID, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("osbGuid"), serviceclass.OSBGUID, msg))
	}

	planNames := sets.NewString()
	for i, plan := range serviceclass.Plans {
		planPath := field.NewPath("plans").Index(i)
		allErrs = append(allErrs, validateServicePlan(plan, planPath)...)

		if planNames.Has(plan.Name) {
			allErrs = append(allErrs, field.Invalid(planPath.Child("name"), plan.Name, "each plan must have a unique name"))
		}
	}

	return allErrs
}

// validateServicePlan validates the fields of a single ServicePlan and
// returns a list of errors.
func validateServicePlan(plan sc.ServicePlan, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, msg := range validateServicePlanName(plan.Name, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), plan.Name, msg))
	}

	if "" == plan.OSBGUID {
		allErrs = append(allErrs, field.Required(fldPath.Child("osbGuid"), "osbGuid is required"))
	}

	for _, msg := range validateOSBGuid(plan.OSBGUID, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("osbGuid"), plan.OSBGUID, msg))
	}

	return allErrs
}

// ValidateServiceClassUpdate checks that when changing from an older
// ServiceClass to a newer ServiceClass is okay.
func ValidateServiceClassUpdate(new *sc.ServiceClass, old *sc.ServiceClass) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateServiceClass(new)...)
	allErrs = append(allErrs, ValidateServiceClass(old)...)

	return allErrs
}
