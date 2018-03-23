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
	"fmt"
	"regexp"

	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

const commonServicePlanNameFmt string = `[-.a-zA-Z0-9]+`
const commonServicePlanNameMaxLength int = 63

var servicePlanNameRegexp = regexp.MustCompile("^" + commonServicePlanNameFmt + "$")

// validateCommonServicePlanName is the common validation function for
// service plan types.
func validateCommonServicePlanName(value string, prefix bool) []string {
	var errs []string
	if len(value) > commonServicePlanNameMaxLength {
		errs = append(errs, utilvalidation.MaxLenError(commonServicePlanNameMaxLength))
	}
	if !servicePlanNameRegexp.MatchString(value) {
		errs = append(errs, utilvalidation.RegexError(commonServicePlanNameFmt, "plan-name-40d-0983-1b89"))
	}

	return errs
}

// ValidateClusterServicePlan validates a ClusterServicePlan and returns a list of errors.
func ValidateClusterServicePlan(clusterServicePlan *sc.ClusterServicePlan) field.ErrorList {
	return validateClusterServicePlan(clusterServicePlan)
}

func validateClusterServicePlan(clusterServicePlan *sc.ClusterServicePlan) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(
			&clusterServicePlan.ObjectMeta,
			false, /* namespace required */
			validateCommonServicePlanName,
			field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateClusterServicePlanSpec(&clusterServicePlan.Spec, field.NewPath("spec"))...)
	return allErrs
}

func validateCommonServicePlanSpec(spec sc.CommonServicePlanSpec, fldPath *field.Path) field.ErrorList {

	allErrs := field.ErrorList{}

	if "" == spec.ExternalID {
		allErrs = append(allErrs, field.Required(fldPath.Child("externalID"), "externalID is required"))
	}

	if "" == spec.Description {
		allErrs = append(allErrs, field.Required(fldPath.Child("description"), "description is required"))
	}

	for _, msg := range validateExternalID(spec.ExternalID) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("externalID"), spec.ExternalID, msg))
	}

	for _, msg := range validateCommonServicePlanName(spec.ExternalName, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("externalName"), spec.ExternalName, msg))
	}

	return allErrs

}

func validateClusterServicePlanSpec(spec *sc.ClusterServicePlanSpec, fldPath *field.Path) field.ErrorList {
	allErrs := validateCommonServicePlanSpec(spec.CommonServicePlanSpec, fldPath)

	if "" == spec.ClusterServiceBrokerName {
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServiceBrokerName"), "clusterServiceBrokerName is required"))
	}

	if "" == spec.ClusterServiceClassRef.Name {
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServiceClassRef"), "an owning serviceclass is required"))
	}

	for _, msg := range validateCommonServiceClassName(spec.ClusterServiceClassRef.Name, false /* prefix */) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServiceClassRef", "name"), spec.ClusterServiceClassRef.Name, msg))
	}

	return allErrs
}

// ValidateClusterServicePlanUpdate checks that when changing from an older
// ClusterServicePlan to a newer ClusterServicePlan is okay.
func ValidateClusterServicePlanUpdate(new *sc.ClusterServicePlan, old *sc.ClusterServicePlan) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateClusterServicePlan(new)...)
	allErrs = append(allErrs, validateCommonServicePlanUpdate(new.Spec.CommonServicePlanSpec, old.Spec.CommonServicePlanSpec, "ClusterServicePlan")...)
	return allErrs
}

func validateCommonServicePlanUpdate(new sc.CommonServicePlanSpec, old sc.CommonServicePlanSpec, resourceType string) field.ErrorList {
	allErrs := field.ErrorList{}
	if new.ExternalID != old.ExternalID {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("externalID"), new.ExternalID, fmt.Sprintf("externalID cannot change when updating a %s", resourceType)))
	}
	return allErrs
}
