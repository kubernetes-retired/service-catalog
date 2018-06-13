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

// GetSpecifiedServicePlan returns the user-specified class value from either:
// * ServicePlanExternalName
// * ServicePlanExternalID
// * ServicePlanName
func (pr PlanReference) GetSpecifiedServicePlan() string {
	if pr.ServicePlanExternalName != "" {
		return pr.ServicePlanExternalName
	}

	if pr.ServicePlanExternalID != "" {
		return pr.ServicePlanExternalID
	}

	if pr.ServicePlanName != "" {
		return pr.ServicePlanName
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

// GetClusterServiceClassFilterFieldName returns the appropriate field name for filtering
// a list of service catalog classes by the PlanReference.
func (pr PlanReference) GetClusterServiceClassFilterFieldName() string {
	if pr.ClusterServiceClassExternalName != "" {
		return "spec.externalName"
	}

	if pr.ClusterServiceClassExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// GetClusterServicePlanFilterFieldName returns the appropriate field name for filtering
// a list of service catalog plans by the PlanReference.
func (pr PlanReference) GetClusterServicePlanFilterFieldName() string {
	if pr.ClusterServicePlanExternalName != "" {
		return "spec.externalName"
	}

	if pr.ClusterServicePlanExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// GetServiceClassFilterFieldName returns the appropriate field name for filtering
// a list of service catalog classes by the PlanReference.
func (pr PlanReference) GetServiceClassFilterFieldName() string {
	if pr.ServiceClassExternalName != "" {
		return "spec.externalName"
	}

	if pr.ServiceClassExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// GetServicePlanFilterFieldName returns the appropriate field name for filtering
// a list of service catalog plans by the PlanReference.
func (pr PlanReference) GetServicePlanFilterFieldName() string {
	if pr.ServicePlanExternalName != "" {
		return "spec.externalName"
	}

	if pr.ServicePlanExternalID != "" {
		return "spec.externalID"
	}

	return ""
}

// String representation of a PlanReference
// Example: class_name/plan_name, class_id/plan_id
func (pr PlanReference) String() string {
	var rep string
	if pr.ClusterServiceClassSpecified() && pr.ClusterServicePlanSpecified() {
		rep = fmt.Sprintf("%s/%s", pr.GetSpecifiedClusterServiceClass(), pr.GetSpecifiedClusterServicePlan())
	} else {
		// Namespace scoped
		rep = fmt.Sprintf("%s/%s", pr.GetSpecifiedServiceClass(), pr.GetSpecifiedServicePlan())
	}
	return rep
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
	var planFields []string

	if pr.ClusterServiceClassExternalName != "" {
		classFields = append(classFields, fmt.Sprintf("ClusterServiceClassExternalName:%q", pr.ClusterServiceClassExternalName))
	}
	if pr.ClusterServiceClassExternalID != "" {
		classFields = append(classFields, fmt.Sprintf("ClusterServiceClassExternalID:%q", pr.ClusterServiceClassExternalID))
	}
	if pr.ClusterServiceClassName != "" {
		classFields = append(classFields, fmt.Sprintf("ClusterServiceClassName:%q", pr.ClusterServiceClassName))
	}

	if pr.ClusterServicePlanExternalName != "" {
		planFields = append(planFields, fmt.Sprintf("ClusterServicePlanExternalName:%q", pr.ClusterServicePlanExternalName))
	}
	if pr.ClusterServicePlanExternalID != "" {
		planFields = append(planFields, fmt.Sprintf("ClusterServicePlanExternalID:%q", pr.ClusterServicePlanExternalID))
	}
	if pr.ClusterServicePlanName != "" {
		planFields = append(planFields, fmt.Sprintf("ClusterServicePlanName:%q", pr.ClusterServicePlanName))
	}

	if pr.ServiceClassExternalName != "" {
		classFields = append(classFields, fmt.Sprintf("ServiceClassExternalName:%q", pr.ServiceClassExternalName))
	}
	if pr.ServiceClassExternalID != "" {
		classFields = append(classFields, fmt.Sprintf("ServiceClassExternalID:%q", pr.ServiceClassExternalID))
	}
	if pr.ServiceClassName != "" {
		classFields = append(classFields, fmt.Sprintf("ServiceClassName:%q", pr.ServiceClassName))
	}

	if pr.ServicePlanExternalName != "" {
		planFields = append(planFields, fmt.Sprintf("ServicePlanExternalName:%q", pr.ServicePlanExternalName))
	}
	if pr.ServicePlanExternalID != "" {
		planFields = append(planFields, fmt.Sprintf("ServicePlanExternalID:%q", pr.ServicePlanExternalID))
	}
	if pr.ServicePlanName != "" {
		planFields = append(planFields, fmt.Sprintf("ServicePlanName:%q", pr.ServicePlanName))
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
