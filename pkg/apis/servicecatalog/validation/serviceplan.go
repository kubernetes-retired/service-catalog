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
	"regexp"

	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

const servicePlanNameFmt string = `[-a-z0-9]+`
const servicePlanNameMaxLength int = 63

var servicePlanNameRegexp = regexp.MustCompile("^" + servicePlanNameFmt + "$")

// validateServicePlanName is the validation function for ServicePlan names.
func validateServicePlanName(value string, prefix bool) []string {
	var errs []string
	if len(value) > servicePlanNameMaxLength {
		errs = append(errs, utilvalidation.MaxLenError(servicePlanNameMaxLength))
	}
	if !servicePlanNameRegexp.MatchString(value) {
		errs = append(errs, utilvalidation.RegexError(servicePlanNameFmt, "plan-name-40d-0983-1b89"))
	}

	return errs
}

// ValidateServicePlan validates a ServicePlan and returns a list of errors.
func ValidateServicePlan(serviceplan *sc.ServicePlan) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(
			&serviceplan.ObjectMeta,
			false, /* namespace required */
			validateServicePlanName,
			field.NewPath("metadata"))...)

	if "" == serviceplan.ExternalID {
		allErrs = append(allErrs, field.Required(field.NewPath("externalID"), "externalID is required"))
	}

	if "" == serviceplan.Description {
		allErrs = append(allErrs, field.Required(field.NewPath("description"), "description is required"))
	}

	if "" == serviceplan.ServiceClassRef.Name {
		allErrs = append(allErrs, field.Required(field.NewPath("serviceClassRef"), "an owning serviceclass is required"))
	}

	for _, msg := range validateExternalID(serviceplan.ExternalID) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("externalID"), serviceplan.ExternalID, msg))
	}

	for _, msg := range validateServiceClassName(serviceplan.ServiceClassRef.Name, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("serviceClassRef", "name"), serviceplan.ServiceClassRef.Name, msg))
	}

	return allErrs
}

// ValidateServicePlanUpdate checks that when changing from an older
// ServicePlan to a newer ServicePlan is okay.
func ValidateServicePlanUpdate(new *sc.ServicePlan, old *sc.ServicePlan) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateServicePlan(new)...)
	allErrs = append(allErrs, ValidateServicePlan(old)...)
	if new.ExternalID != old.ExternalID {
		allErrs = append(allErrs, field.Invalid(field.NewPath("externalID"), new.ExternalID, "externalID cannot change when updating a ServicePlan"))
	}
	return allErrs
}
