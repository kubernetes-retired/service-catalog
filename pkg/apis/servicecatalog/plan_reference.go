/*
Copyright 2018 The Kubernetes Authors.

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

package servicecatalog

import (
	"fmt"
	"strings"
)

// ClusterServiceClassSpecified checks that at least one class field is set.
func (pr PlanReference) ClusterServiceClassSpecified() bool {
	return pr.ClusterServiceClassExternalName != "" ||
		pr.ClusterServiceClassExternalID != "" ||
		pr.ClusterServiceClassName != ""
}

// ClusterServicePlanSpecified checks that at least one plan field is set.
func (pr PlanReference) ClusterServicePlanSpecified() bool {
	return pr.ClusterServicePlanExternalName != "" ||
		pr.ClusterServicePlanExternalID != "" ||
		pr.ClusterServicePlanName != ""
}

// ServiceClassSpecified checks that at least one serviceclass field is set.
func (pr PlanReference) ServiceClassSpecified() bool {
	return pr.ServiceClassExternalName != "" ||
		pr.ServiceClassExternalID != "" ||
		pr.ServiceClassName != ""
}

// ServicePlanSpecified checks that at least one serviceplan field is set.
func (pr PlanReference) ServicePlanSpecified() bool {
	return pr.ServicePlanExternalName != "" ||
		pr.ServicePlanExternalID != "" ||
		pr.ServicePlanName != ""
}

// GetSpecifiedClusterServiceClass returns the user-specified class value from one of:
// * ClusterServiceClassExternalName
// * ClusterServiceClassExternalID
// * ClusterServiceClassName
// This method is intended for presentation purposes only.
func (pr PlanReference) GetSpecifiedClusterServiceClass() string {
	if pr.ClusterServiceClassExternalName != "" {
		return pr.ClusterServiceClassExternalName
	}

	if pr.ClusterServiceClassExternalID != "" {
		return pr.ClusterServiceClassExternalID
	}

	if pr.ClusterServiceClassName != "" {
		return pr.ClusterServiceClassName
	}

	return ""
}

// GetSpecifiedServiceClass returns the user-specified class value from either:
// * ServiceClassExternalName
// * ServiceClassExternalID
// * ServiceClassName
func (pr PlanReference) GetSpecifiedServiceClass() string {
	if pr.ServiceClassExternalName != "" {
		return pr.ServiceClassExternalName
	}

	if pr.ServiceClassExternalID != "" {
		return pr.ServiceClassExternalID
	}

	if pr.ServiceClassName != "" {
		return pr.ServiceClassName
	}

	return ""
}

// GetSpecifiedClusterServicePlan returns the user-specified plan value from one of:
// * ClusterServicePlanExternalName
// * ClusterServicePlanExternalID
// * ClusterServicePlanName
// This method is intended for presentation purposes only.
func (pr PlanReference) GetSpecifiedClusterServicePlan() string {
	if pr.ClusterServicePlanExternalName != "" {
		return pr.ClusterServicePlanExternalName
	}

	if pr.ClusterServicePlanExternalID != "" {
		return pr.ClusterServicePlanExternalID
	}

	if pr.ClusterServicePlanName != "" {
		return pr.ClusterServicePlanName
	}

	return ""
}

// GetClassFilterFieldName returns the appropriate field name for filtering
// a list of service catalog classes by the PlanReference.
func (pr PlanReference) GetClassFilterFieldName() string {
	if pr.ClusterServiceClassExternalName != "" {
		return "spec.externalName"
	}

	if pr.ClusterServiceClassExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// GetPlanFilterFieldName returns the appropriate field name for filtering
// a list of service catalog plans by the PlanReference.
func (pr PlanReference) GetPlanFilterFieldName() string {
	if pr.ClusterServicePlanExternalName != "" {
		return "spec.externalName"
	}

	if pr.ClusterServicePlanExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// String representation of a PlanReference
// Example: class_name/plan_name, class_id/plan_id
func (pr PlanReference) String() string {
	return fmt.Sprintf("%s/%s", pr.GetSpecifiedClusterServiceClass(), pr.GetSpecifiedClusterServicePlan())
}

// Format the PlanReference
// %c - Print specified class fields only
//    Examples:
//     {ClassExternalName:"foo"}
//     {ClassExternalID:"foo123"}
//     {ClassName:"k8s-foo123"}
// %b - Print specified plan fields only
//    NOTE: %p is a reserved verb so we can't use it, and go vet fails for non-standard verbs
//    Examples:
//     {PlanExternalName:"bar"}
//     {PlanExternalID:"bar456"}
//     {PlanName:"k8s-bar456"}
// %s - Print a short form of the plan and class
//    Examples:
//     foo/bar
//     foo123/bar456
//     k8s-foo123/k8s-bar456
// %v - Print all specified fields
//    Examples:
//     {ClassExternalName:"foo", PlanExternalName:"bar"}
//     {ClassExternalID:"foo123", PlanExternalID:"bar456"}
//     {ClassName:"k8s-foo123", PlanName:"k8s-bar456"}
func (pr PlanReference) Format(s fmt.State, verb rune) {
	var classFields []string
	if pr.ClusterServiceClassExternalName != "" {
		classFields = append(classFields, fmt.Sprintf("ClassExternalName:%q", pr.ClusterServiceClassExternalName))
	}
	if pr.ClusterServiceClassExternalID != "" {
		classFields = append(classFields, fmt.Sprintf("ClassExternalID:%q", pr.ClusterServiceClassExternalID))
	}
	if pr.ClusterServiceClassName != "" {
		classFields = append(classFields, fmt.Sprintf("ClassName:%q", pr.ClusterServiceClassName))
	}

	var planFields []string
	if pr.ClusterServicePlanExternalName != "" {
		planFields = append(planFields, fmt.Sprintf("PlanExternalName:%q", pr.ClusterServicePlanExternalName))
	}
	if pr.ClusterServicePlanExternalID != "" {
		planFields = append(planFields, fmt.Sprintf("PlanExternalID:%q", pr.ClusterServicePlanExternalID))
	}
	if pr.ClusterServicePlanName != "" {
		planFields = append(planFields, fmt.Sprintf("PlanName:%q", pr.ClusterServicePlanName))
	}

	switch verb {
	case 'c':
		fmt.Fprintf(s, "{%s}", strings.Join(classFields, ", "))
	case 'b':
		fmt.Fprintf(s, "{%s}", strings.Join(planFields, ", "))
	case 'v':
		fmt.Fprintf(s, "{%s}", strings.Join(append(classFields, planFields...), ", "))
	}
}
