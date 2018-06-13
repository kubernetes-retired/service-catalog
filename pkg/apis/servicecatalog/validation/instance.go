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
	"github.com/ghodss/yaml"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

// validateServiceInstanceName is the validation function for Instance names.
var validateServiceInstanceName = apivalidation.NameIsDNSSubdomain

var validServiceInstanceOperations = map[sc.ServiceInstanceOperation]bool{
	sc.ServiceInstanceOperation(""):        true,
	sc.ServiceInstanceOperationProvision:   true,
	sc.ServiceInstanceOperationUpdate:      true,
	sc.ServiceInstanceOperationDeprovision: true,
}

var validServiceInstanceOperationValues = func() []string {
	validValues := make([]string, len(validServiceInstanceOperations))
	i := 0
	for operation := range validServiceInstanceOperations {
		validValues[i] = string(operation)
		i++
	}
	return validValues
}()

var validServiceInstanceDeprovisionStatuses = map[sc.ServiceInstanceDeprovisionStatus]bool{
	sc.ServiceInstanceDeprovisionStatusNotRequired: true,
	sc.ServiceInstanceDeprovisionStatusRequired:    true,
	sc.ServiceInstanceDeprovisionStatusSucceeded:   true,
	sc.ServiceInstanceDeprovisionStatusFailed:      true,
}

var validServiceInstanceDeprovisionStatusValues = func() []string {
	validValues := make([]string, len(validServiceInstanceDeprovisionStatuses))
	i := 0
	for operation := range validServiceInstanceDeprovisionStatuses {
		validValues[i] = string(operation)
		i++
	}
	return validValues
}()

// ValidateServiceInstance validates an Instance and returns a list of errors.
func ValidateServiceInstance(instance *sc.ServiceInstance) field.ErrorList {
	return internalValidateServiceInstance(instance, true)
}

func internalValidateServiceInstance(instance *sc.ServiceInstance, create bool) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, apivalidation.ValidateObjectMeta(&instance.ObjectMeta, true, /*namespace*/
		validateServiceInstanceName,
		field.NewPath("metadata"))...)
	allErrs = append(allErrs, validateServiceInstanceSpec(&instance.Spec, field.NewPath("spec"), create)...)
	allErrs = append(allErrs, validateServiceInstanceStatus(&instance.Status, field.NewPath("status"), create)...)
	if create {
		allErrs = append(allErrs, validateServiceInstanceCreate(instance)...)
	} else {
		allErrs = append(allErrs, validateServiceInstanceUpdate(instance)...)
	}
	return allErrs
}

func validateServiceInstanceSpec(spec *sc.ServiceInstanceSpec, fldPath *field.Path, create bool) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateObjectReferences(spec, fldPath)...)
	allErrs = append(allErrs, validatePlanReference(&spec.PlanReference, fldPath)...)

	if spec.ParametersFrom != nil {
		allErrs = append(allErrs, validateParametersFromSource(spec.ParametersFrom, fldPath)...)
	}
	if spec.Parameters != nil {
		if len(spec.Parameters.Raw) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("parameters"), "inline parameters must not be empty if present"))
		}
		if _, err := controller.UnmarshalRawParameters(spec.Parameters.Raw); err != nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("parameters"), "invalid inline parameters"))
		}
	}

	allErrs = append(allErrs, apivalidation.ValidateNonnegativeField(spec.UpdateRequests, fldPath.Child("updateRequests"))...)

	return allErrs
}

func validateServiceInstanceStatus(status *sc.ServiceInstanceStatus, fldPath *field.Path, create bool) field.ErrorList {
	allErrs := field.ErrorList{}

	if create {
		if status.CurrentOperation != "" {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("currentOperation"), status.CurrentOperation, "currentOperation must be empty on create"))
		}
	} else {
		if !validServiceInstanceOperations[status.CurrentOperation] {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("currentOperation"), status.CurrentOperation, validServiceInstanceOperationValues))
		}
	}

	if status.CurrentOperation == "" {
		if status.OperationStartTime != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("operationStartTime"), "operationStartTime must not be present when currentOperation is not present"))
		}
		if status.AsyncOpInProgress {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("asyncOpInProgress"), "asyncOpInProgress cannot be true when there is no currentOperation"))
		}
		if status.LastOperation != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("lastOperation"), "lastOperation cannot be true when currentOperation is not present"))
		}
	} else {
		if status.OperationStartTime == nil && !status.OrphanMitigationInProgress {
			allErrs = append(allErrs, field.Required(fldPath.Child("operationStartTime"), "operationStartTime is required when currentOperation is present and no orphan mitigation in progress"))
		}
		// Do not allow the instance to be ready if there is an on-going operation
		for i, c := range status.Conditions {
			if c.Type == sc.ServiceInstanceConditionReady && c.Status == sc.ConditionTrue {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("conditions").Index(i), "Can not set ServiceInstanceConditionReady to true when there is an operation in progress"))
			}
		}
	}

	switch status.CurrentOperation {
	case sc.ServiceInstanceOperationProvision, sc.ServiceInstanceOperationUpdate, sc.ServiceInstanceOperationDeprovision:
		if status.InProgressProperties == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("inProgressProperties"), `inProgressProperties is required when currentOperation is "Provision", "Update" or "Deprovision"`))
		}
	default:
		if status.InProgressProperties != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("inProgressProperties"), `inProgressProperties must not be present when currentOperation is not "Provision", "Update" or "Deprovision"`))
		}
	}

	if status.InProgressProperties != nil {
		allErrs = append(allErrs, validateServiceInstancePropertiesState(status.InProgressProperties, fldPath.Child("inProgressProperties"), create)...)
	}

	if status.ExternalProperties != nil {
		allErrs = append(allErrs, validateServiceInstancePropertiesState(status.ExternalProperties, fldPath.Child("externalProperties"), create)...)
	}

	if create {
		if status.DeprovisionStatus != sc.ServiceInstanceDeprovisionStatusNotRequired {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("deprovisionStatus"), status.DeprovisionStatus, `deprovisionStatus must be "NotRequired" on create`))
		}
	} else {
		if !validServiceInstanceDeprovisionStatuses[status.DeprovisionStatus] {
			allErrs = append(allErrs, field.NotSupported(fldPath.Child("deprovisionStatus"), status.DeprovisionStatus, validServiceInstanceDeprovisionStatusValues))
		}
	}

	return allErrs
}

func validateServiceInstancePropertiesState(propertiesState *sc.ServiceInstancePropertiesState, fldPath *field.Path, create bool) field.ErrorList {
	var errMsg string
	allErrs := field.ErrorList{}

	if propertiesState.ClusterServicePlanExternalName == "" && propertiesState.ServicePlanExternalName == "" {
		errMsg = "clusterServicePlanExternalName or servicePlanExternalName is required"
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServicePlanExternalName"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("servicePlanExternalName"), errMsg))
	}

	if propertiesState.ClusterServicePlanExternalName != "" && propertiesState.ServicePlanExternalName != "" {
		errMsg = "clusterServicePlanExternalName and servicePlanExternalName cannot both be set"
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanExternalName"), propertiesState.ClusterServicePlanExternalName, errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("servicePlanExternalName"), propertiesState.ServicePlanExternalName, errMsg))
	}

	if propertiesState.ClusterServicePlanExternalID == "" && propertiesState.ServicePlanExternalID == "" {
		errMsg = "clusterServicePlanExternalID or servicePlanExternalID is required"
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServicePlanExternalID"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("servicePlanExternalID"), errMsg))
	}

	if propertiesState.ClusterServicePlanExternalID != "" && propertiesState.ServicePlanExternalID != "" {
		errMsg = "clusterServicePlanExternalID and servicePlanExternalID cannot both be set"
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanExternalID"), propertiesState.ClusterServicePlanExternalID, errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("servicePlanExternalID"), propertiesState.ServicePlanExternalID, errMsg))
	}

	if propertiesState.Parameters == nil {
		if propertiesState.ParametersChecksum != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("parametersChecksum"), "parametersChecksum must be empty when there are no parameters"))
		}
	} else {
		if len(propertiesState.Parameters.Raw) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("parameters").Child("raw"), "raw must not be empty"))
		} else {
			unmarshalled := make(map[string]interface{})
			if err := yaml.Unmarshal(propertiesState.Parameters.Raw, &unmarshalled); err != nil {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("parameters").Child("raw"), propertiesState.Parameters.Raw, "raw must be valid yaml"))
			}
		}
		if propertiesState.ParametersChecksum == "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("parametersChecksum"), "parametersChecksum must not be empty when there are parameters"))
		}
	}

	if propertiesState.ParametersChecksum != "" {
		if len(propertiesState.ParametersChecksum) != 64 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("parametersChecksum"), propertiesState.ParametersChecksum, "parametersChecksum must be exactly 64 digits"))
		}
		if !stringIsHexadecimal(propertiesState.ParametersChecksum) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("parametersChecksum"), propertiesState.ParametersChecksum, "parametersChecksum must be a hexadecimal number"))
		}
	}

	return allErrs
}

func validateServiceInstanceCreate(instance *sc.ServiceInstance) field.ErrorList {
	allErrs := field.ErrorList{}
	if instance.Status.ReconciledGeneration >= instance.Generation {
		allErrs = append(allErrs, field.Invalid(field.NewPath("status").Child("reconciledGeneration"), instance.Status.ReconciledGeneration, "reconciledGeneration must be less than generation on create"))
	}
	if instance.Spec.ClusterServiceClassRef != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("clusterServiceClassRef"), "clusterServiceClassRef must not be present on create"))
	}
	if instance.Spec.ClusterServicePlanRef != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("clusterServicePlanRef"), "clusterServicePlanRef must not be present on create"))
	}
	if instance.Spec.ServiceClassRef != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("serviceClassRef"), "serviceClassRef must not be present on create"))
	}
	if instance.Spec.ServicePlanRef != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec").Child("servicePlanRef"), "servicePlanRef must not be present on create"))
	}
	return allErrs
}

func validateServiceInstanceUpdate(instance *sc.ServiceInstance) field.ErrorList {
	var errMsg string
	allErrs := field.ErrorList{}

	if instance.Status.ReconciledGeneration == instance.Generation {
		if instance.Status.CurrentOperation != "" {
			allErrs = append(allErrs, field.Forbidden(field.NewPath("status").Child("currentOperation"), "currentOperation must not be present when reconciledGeneration and generation are equal"))
		}
	} else if instance.Status.ReconciledGeneration > instance.Generation {
		allErrs = append(allErrs, field.Invalid(field.NewPath("status").Child("reconciledGeneration"), instance.Status.ReconciledGeneration, "reconciledGeneration must not be greater than generation"))
	}
	if instance.Status.CurrentOperation != "" {
		if instance.Spec.ClusterServiceClassRef == nil && instance.Spec.ServiceClassRef == nil {
			errMsg = "clusterServiceClassRef or serviceClassRef is required when currentOperation is present"
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("clusterServiceClassRef"), errMsg))
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("serviceClassRef"), errMsg))
		}
		if instance.Status.CurrentOperation != sc.ServiceInstanceOperationDeprovision {
			if instance.Spec.ClusterServicePlanRef == nil && instance.Spec.ServicePlanRef == nil {
				errMsg = "clusterServicePlanRef or servicePlanRef is required when currentOperation is present"
				allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("clusterServicePlanRef"), errMsg))
				allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("servicePlanRef"), errMsg))
			}
		} else {
			clusterUnset := instance.Spec.ClusterServicePlanRef == nil &&
				(instance.Status.ExternalProperties == nil || instance.Status.ExternalProperties.ClusterServicePlanExternalID == "")
			nsUnset := instance.Spec.ServicePlanRef == nil &&
				(instance.Status.ExternalProperties == nil || instance.Status.ExternalProperties.ServicePlanExternalID == "")
			if clusterUnset && nsUnset {
				errMsg = "spec.clusterServicePlanRef, status.externalProperties.clusterServicePlanExternalID, spec.servicePlanRef, or status.externalProperties.servicePlanExternalID is required when currentOperation is Deprovision"
				allErrs = append(allErrs, field.Invalid(field.NewPath("status").Child("currentOperation"), instance.Status.CurrentOperation, errMsg))
			}
		}
	}
	return allErrs
}

// internalValidateServiceInstanceUpdateAllowed ensures there is not a
// pending update on-going with the spec of the instance before allowing an update
// to the spec to go through.
func internalValidateServiceInstanceUpdateAllowed(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	errors := field.ErrorList{}

	// If the OriginatingIdentityLocking feature is set then don't allow spec updates
	// if processing of the current generation hasn't finished yet
	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.OriginatingIdentityLocking) {
		// TODO nilebox: The condition for locking should not be based on whether
		// there is an operation in progress. It should be based on whether controller
		// has finished processing the current generation (i.e. either succeeded, or failed and won't retry).
		// In other words, check for ObservedGeneration + conditions instead of CurrentOperation
		if old.Generation != new.Generation && old.Status.CurrentOperation != "" {
			errors = append(errors, field.Forbidden(field.NewPath("spec"), "Another update for this service instance is in progress"))
		}
	}

	clusterPlanUpdated := old.Spec.ClusterServicePlanExternalName != new.Spec.ClusterServicePlanExternalName
	clusterPlanUpdated = clusterPlanUpdated || old.Spec.ClusterServicePlanExternalID != new.Spec.ClusterServicePlanExternalID
	clusterPlanUpdated = clusterPlanUpdated || old.Spec.ClusterServicePlanName != new.Spec.ClusterServicePlanName

	nsPlanUpdated := old.Spec.ServicePlanExternalName != new.Spec.ServicePlanExternalName
	nsPlanUpdated = nsPlanUpdated || old.Spec.ServicePlanExternalID != new.Spec.ServicePlanExternalID
	nsPlanUpdated = nsPlanUpdated || old.Spec.ServicePlanName != new.Spec.ServicePlanName

	if clusterPlanUpdated && new.Spec.ClusterServicePlanRef != nil {
		errors = append(errors, field.Forbidden(field.NewPath("spec").Child("clusterServicePlanRef"), "clusterServicePlanRef must not be present when the plan is being changed"))
	} else if nsPlanUpdated && new.Spec.ServicePlanRef != nil {
		errors = append(errors, field.Forbidden(field.NewPath("spec").Child("servicePlanRef"), "servicePlanRef must not be present when the plan is being changed"))
	}

	return errors
}

// ValidateServiceInstanceUpdate validates a change to the Instance's spec.
func ValidateServiceInstanceUpdate(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	allErrs := field.ErrorList{}

	specFieldPath := field.NewPath("spec")

	allErrs = append(allErrs, validatePlanReferenceUpdate(&new.Spec.PlanReference, &old.Spec.PlanReference, specFieldPath)...)
	allErrs = append(allErrs, internalValidateServiceInstanceUpdateAllowed(new, old)...)
	allErrs = append(allErrs, internalValidateServiceInstance(new, false)...)

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.Spec.ExternalID, old.Spec.ExternalID, specFieldPath.Child("externalID"))...)

	if new.Spec.UpdateRequests < old.Spec.UpdateRequests {
		allErrs = append(allErrs, field.Invalid(specFieldPath.Child("updateRequests"), new.Spec.UpdateRequests, "new updateRequests value must not be less than the old one"))
	}

	return allErrs
}

func internalValidateServiceInstanceStatusUpdateAllowed(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	errors := field.ErrorList{}
	// TODO(vaikas): Are there any cases where we do not allow updates to
	// Status during Async updates in progress?
	return errors
}

func internalValidateServiceInstanceReferencesUpdateAllowed(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	var errMsg string
	allErrs := field.ErrorList{}

	if new.Status.CurrentOperation != "" {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("status").Child("currentOperation"), "cannot update references when currentOperation is present"))
	}

	if new.Spec.ClusterServiceClassRef == nil && new.Spec.ServiceClassRef == nil {
		errMsg = "clusterServiceClassRef or serviceClassRef is required when updating references"
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("clusterServiceClassRef"), errMsg))
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("serviceClassRef"), errMsg))
	}
	if new.Spec.ClusterServicePlanRef == nil && new.Spec.ServicePlanRef == nil {
		errMsg = "clusterServicePlanRef or servicePlanRef is required when updating references"
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("clusterServicePlanRef"), errMsg))
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("servicePlanRef"), errMsg))
	}
	if new.Spec.ClusterServiceClassRef != nil && new.Spec.ServiceClassRef != nil {
		errMsg = "clusterServiceClassRef and serviceClassRef cannot both be set when updating references"
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("clusterServiceClassRef"), new.Spec.ClusterServiceClassRef, errMsg)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("serviceClassRef"), new.Spec.ServiceClassRef, errMsg)
	}
	if new.Spec.ClusterServicePlanRef != nil && new.Spec.ServicePlanRef != nil {
		errMsg = "clusterServicePlanRef and servicePlanRef cannot both be set when updating references"
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("clusterServicePlanRef"), new.Spec.ClusterServicePlanRef, errMsg)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("servicePlanRef"), new.Spec.ServicePlanRef, errMsg)
	}

	if old.Spec.ClusterServiceClassRef != nil {
		allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.Spec.ClusterServiceClassRef, old.Spec.ClusterServiceClassRef, field.NewPath("spec").Child("clusterServiceClassRef"))...)
	}
	if old.Spec.ClusterServicePlanRef != nil {
		allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.Spec.ClusterServicePlanRef, old.Spec.ClusterServicePlanRef, field.NewPath("spec").Child("clusterServicePlanRef"))...)
	}
	if old.Spec.ServiceClassRef != nil {
		allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.Spec.ServiceClassRef, old.Spec.ServiceClassRef, field.NewPath("spec").Child("serviceClassRef"))...)
	}
	if old.Spec.ServicePlanRef != nil {
		allErrs = append(allErrs, apivalidation.ValidateImmutableField(new.Spec.ServicePlanRef, old.Spec.ServicePlanRef, field.NewPath("spec").Child("servicePlanRef"))...)
	}
	return allErrs
}

// ValidateServiceInstanceStatusUpdate checks that when changing from an older
// instance to a newer instance is okay. This only checks the instance.Status field.
func ValidateServiceInstanceStatusUpdate(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, internalValidateServiceInstanceStatusUpdateAllowed(new, old)...)
	allErrs = append(allErrs, internalValidateServiceInstance(new, false)...)
	return allErrs
}

// ValidateServiceInstanceReferencesUpdate checks that when changing from an older
// instance to a newer instance is okay.
func ValidateServiceInstanceReferencesUpdate(new *sc.ServiceInstance, old *sc.ServiceInstance) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, internalValidateServiceInstanceReferencesUpdateAllowed(new, old)...)
	allErrs = append(allErrs, internalValidateServiceInstance(new, false)...)
	return allErrs
}

func validateObjectReferences(spec *sc.ServiceInstanceSpec, fldPath *field.Path) field.ErrorList {
	var errMsg string
	allErrs := field.ErrorList{}

	if spec.ClusterServiceClassRef != nil && spec.ServiceClassRef != nil {
		errMsg = "ClusterServiceClassRef and ServiceClassRef should never be set simultaneously"
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServiceClassRef"), spec.ClusterServiceClassRef, errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceClassRef"), spec.ServiceClassRef, errMsg))
	}

	if spec.ClusterServicePlanRef != nil && spec.ServicePlanRef != nil {
		errMsg = "ClusterServicePlanRef and ServicePlanRef should never be set simultaneously"
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanRef"), spec.ClusterServicePlanRef, errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceClassRef"), spec.ServicePlanRef, errMsg))
	}

	return allErrs
}

const (
	clusterScopedPlanReference   = "clusterScoped"
	namespaceScopedPlanReference = "namespaceScoped"
)

type scopedRefHelper struct {
	externalClassName string
	externalPlanName  string
	externalClassID   string
	externalPlanID    string
	k8sClass          string
	k8sPlan           string
	classField        func(string) string
	planField         func(string) string
}

func validatePlanReference(p *sc.PlanReference, fldPath *field.Path) field.ErrorList {
	var errMsg string
	allErrs := field.ErrorList{}

	// Verify an instance refs either cluster *or* namespaced types, but not both.
	cases := []struct {
		cluster string
		ns      string
		field   string
	}{
		{p.ClusterServiceClassExternalName, p.ServiceClassExternalName, "serviceClassExternalName"},
		{p.ClusterServiceClassExternalID, p.ServiceClassExternalID, "serviceClassExternalID"},
		{p.ClusterServiceClassName, p.ServiceClassName, "serviceClassName"},
		{p.ClusterServicePlanExternalName, p.ServicePlanExternalName, "servicePlanExternalName"},
		{p.ClusterServicePlanExternalID, p.ServicePlanExternalID, "servicePlanExternalID"},
		{p.ClusterServicePlanName, p.ServicePlanName, "servicePlanName"},
	}

	var clusterCount, nsCount uint8
	for _, test := range cases {
		if test.ns != "" {
			nsCount++
		}
		if test.cluster != "" {
			clusterCount++
		}
	}

	if clusterCount > 0 && nsCount > 0 {
		errMsg = "instances can only refer to a cluster or namespaced class or plan type, but not both"
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServiceClassExternalName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServiceClassExternalID"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServiceClassName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanExternalName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanExternalID"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("clusterServicePlanName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceClassExternalName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceClassExternalID"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("serviceClassName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("servicePlanExternalName"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("servicePlanExternalID"), "", errMsg))
		allErrs = append(allErrs, field.Invalid(fldPath.Child("servicePlanName"), "", errMsg))
		return allErrs
	}

	if clusterCount == 0 && nsCount == 0 {
		errMsg = "plan references must have a class reference set"
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServiceClassExternalName"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServiceClassExternalID"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("clusterServiceClassName"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("serviceClassExternalName"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("serviceClassExternalID"), errMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child("serviceClassName"), errMsg))
		return allErrs
	}

	// Clue in if we're cluster or ns typed
	var isClusterScoped bool
	if clusterCount > 0 {
		isClusterScoped = true
	} else if nsCount > 0 {
		isClusterScoped = false
	}

	var refHelper scopedRefHelper
	if isClusterScoped {
		refHelper.externalClassName = p.ClusterServiceClassExternalName
		refHelper.externalPlanName = p.ClusterServicePlanExternalName
		refHelper.externalClassID = p.ClusterServiceClassExternalID
		refHelper.externalPlanID = p.ClusterServicePlanExternalID
		refHelper.k8sClass = p.ClusterServiceClassName
		refHelper.k8sPlan = p.ClusterServicePlanName
		refHelper.classField = func(f string) string {
			return fmt.Sprintf("clusterServiceClass%s", f)
		}
		refHelper.planField = func(f string) string {
			return fmt.Sprintf("clusterServicePlan%s", f)
		}
	} else {
		refHelper.externalClassName = p.ServiceClassExternalName
		refHelper.externalPlanName = p.ServicePlanExternalName
		refHelper.externalClassID = p.ServiceClassExternalID
		refHelper.externalPlanID = p.ServicePlanExternalID
		refHelper.k8sClass = p.ServiceClassName
		refHelper.k8sPlan = p.ServicePlanName
		refHelper.classField = func(f string) string {
			return fmt.Sprintf("serviceClass%s", f)
		}
		refHelper.planField = func(f string) string {
			return fmt.Sprintf("servicePlan%s", f)
		}
	}

	return append(allErrs, validateScopedPlanRef(refHelper, p, fldPath)...)
}

func validateScopedPlanRef(h scopedRefHelper, p *sc.PlanReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// helper function to test that exactly one set of plan references are set
	b2i := func(b bool) int8 {
		if b {
			return 1
		}
		return 0
	}
	// Just to make reading of the conditionals in the code easier.
	externalClassNameSet := h.externalClassName != ""
	externalPlanNameSet := h.externalPlanName != ""
	externalClassIDSet := h.externalClassID != ""
	externalPlanIDSet := h.externalPlanID != ""
	k8sClassSet := h.k8sClass != ""
	k8sPlanSet := h.k8sPlan != ""

	// Must specify exactly one source of the class: external id, external name, k8s name.
	if (b2i(externalClassNameSet) + b2i(externalClassIDSet) + b2i(k8sClassSet)) != 1 {
		classSetErrMsg := fmt.Sprintf("exactly one of %s, %s, or %s required",
			h.classField("ExternalName"), h.classField("ExternalID"), h.classField("Name"))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.classField("ExternalName")), classSetErrMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.classField("ExternalID")), classSetErrMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.classField("Name")), classSetErrMsg))
	}

	// Must specify zero or one source of the plan: external id, external name, k8s name.
	// If Zero, assume there is a "default plan" and the defaultserviceplan admission controller
	// will set it up or error out
	// Must specify exactly one source of the plan: external id, external name, k8s name.
	if (b2i(externalPlanNameSet) + b2i(externalPlanIDSet) + b2i(k8sPlanSet)) > 1 {
		planSetErrMsg := fmt.Sprintf("exactly one of %s, %s, or %s required",
			h.planField("ExternalName"), h.planField("ExternalID"), h.planField("Name"))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("ExternalName")), planSetErrMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("ExternalID")), planSetErrMsg))
		allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("Name")), planSetErrMsg))
	}

	var errMsg string
	if externalClassNameSet {
		for _, msg := range validateCommonServiceClassName(h.externalClassName, false /* prefix */) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child(h.classField("ExternalName")), h.externalClassName, msg))
		}

		// If ClassExternalName given, must use PlanExternalName or not specify the plan
		if !externalPlanNameSet {
			if externalPlanIDSet || k8sPlanSet {
				errMsg = fmt.Sprintf("must specify %s with %s", h.planField("ExternalName"), h.classField("ExternalName"))
				allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("ExternalName")), errMsg))
			}
		} else {
			for _, msg := range validateCommonServicePlanName(h.externalPlanName, false /* prefix */) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child(h.planField("ExternalName")), h.externalPlanName, msg))
			}
		}
	} else if externalClassIDSet {
		for _, msg := range validateExternalID(h.externalClassID) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child(h.classField("ExternalID")), h.externalClassID, msg))
		}

		// If ClassExternalID given, must use PlanExternalID or not specify the plan
		if !externalPlanIDSet {
			if externalPlanNameSet || k8sPlanSet {
				errMsg = fmt.Sprintf("must specify %s with %s", h.planField("ExternalID"), h.classField("ExternalID"))
				allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("ExternalID")), errMsg))
			}
		} else {
			for _, msg := range validateExternalID(h.externalPlanID) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child(h.planField("ExternalID")), h.externalPlanID, msg))
			}
		}
	} else if k8sClassSet {
		for _, msg := range validateCommonServiceClassName(h.k8sClass, false /* prefix */) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child(h.classField("Name")), h.k8sClass, msg))
		}

		// If ClassName given, must use PlanName or not specify the plan
		if !k8sPlanSet {
			if externalPlanNameSet || externalPlanIDSet {
				errMsg = fmt.Sprintf("must specify %s with %s", h.planField("Name"), h.classField("Name"))
				allErrs = append(allErrs, field.Required(fldPath.Child(h.planField("Name")), errMsg))
			}
		} else {
			for _, msg := range validateCommonServicePlanName(h.k8sPlan, false /* prefix */) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child(h.planField("Name")), h.k8sPlan, msg))
			}
		}
	}

	return allErrs
}

func validatePlanReferenceUpdate(pOld *sc.PlanReference, pNew *sc.PlanReference, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validatePlanReference(pOld, fldPath)...)
	allErrs = append(allErrs, validatePlanReference(pNew, fldPath)...)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ClusterServiceClassExternalName, pOld.ClusterServiceClassExternalName, field.NewPath("spec").Child("clusterServiceClassExternalName"))...)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ClusterServiceClassExternalID, pOld.ClusterServiceClassExternalID, field.NewPath("spec").Child("clusterServiceClassExternalID"))...)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ClusterServiceClassName, pOld.ClusterServiceClassName, field.NewPath("spec").Child("clusterServiceClassName"))...)

	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ServiceClassExternalName, pOld.ServiceClassExternalName, field.NewPath("spec").Child("serviceClassExternalName"))...)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ServiceClassExternalID, pOld.ServiceClassExternalID, field.NewPath("spec").Child("serviceClassExternalID"))...)
	allErrs = append(allErrs, apivalidation.ValidateImmutableField(pNew.ServiceClassName, pOld.ServiceClassName, field.NewPath("spec").Child("serviceClassName"))...)
	return allErrs
}
