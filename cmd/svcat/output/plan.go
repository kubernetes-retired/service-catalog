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

package output

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
)

func getPlanStatusShort(status v1beta1.ClusterServicePlanStatus) string {
	if status.RemovedFromBrokerCatalog {
		return statusDeprecated
	}
	return statusActive
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type byClass []servicecatalog.Plan

func (a byClass) Len() int {
	return len(a)
}
func (a byClass) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byClass) Less(i, j int) bool {
	return a[i].GetClassID() < a[j].GetClassID()
}

func writePlanListTable(w io.Writer, plans []servicecatalog.Plan, classNames map[string]string) {

	sort.Sort(byClass(plans))

	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Class",
		"Description",
	})
	for _, plan := range plans {
		t.Append([]string{
			plan.GetExternalName(),
			plan.GetNamespace(),
			classNames[plan.GetClassID()],
			plan.GetDescription(),
		})
	}
	t.SetVariableColumn(4)

	t.Render()
}

// WritePlanList prints a list of plans in the specified output format.
func WritePlanList(w io.Writer, outputFormat string, plans []servicecatalog.Plan, classes []servicecatalog.Class) {
	classNames := map[string]string{}
	for _, class := range classes {
		classNames[class.GetName()] = class.GetExternalName()
	}
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, plans)
	case FormatYAML:
		writeYAML(w, plans, 0)
	case FormatTable:
		writePlanListTable(w, plans, classNames)
	}
}

// WritePlan prints a single plan in the specified output format.
func WritePlan(w io.Writer, outputFormat string, plan servicecatalog.Plan, class v1beta1.ClusterServiceClass) {

	switch outputFormat {
	case FormatJSON:
		writeJSON(w, plan)
	case FormatYAML:
		writeYAML(w, plan, 0)
	case FormatTable:
		classNames := map[string]string{}
		classNames[class.Name] = class.Spec.ExternalName
		writePlanListTable(w, []servicecatalog.Plan{plan}, classNames)
	}
}

// WriteAssociatedPlans prints a list of plans associated with a class.
func WriteAssociatedPlans(w io.Writer, plans []servicecatalog.Plan) {
	fmt.Fprintln(w, "\nPlans:")
	if len(plans) == 0 {
		fmt.Fprintln(w, "No plans defined")
		return
	}

	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Description",
	})
	for _, plan := range plans {
		t.Append([]string{
			plan.GetExternalName(),
			plan.GetDescription(),
		})
	}
	t.Render()
}

// WriteParentPlan prints identifying information for a parent class.
func WriteParentPlan(w io.Writer, plan *v1beta1.ClusterServicePlan) {
	fmt.Fprintln(w, "\nPlan:")
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", plan.Spec.ExternalName},
		{"UUID:", string(plan.Name)},
		{"Status:", getPlanStatusShort(plan.Status)},
	})
	t.Render()
}

// WritePlanDetails prints details for a single plan.
func WritePlanDetails(w io.Writer, plan servicecatalog.Plan, class *v1beta1.ClusterServiceClass) {
	t := NewDetailsTable(w)

	t.AppendBulk([][]string{
		{"Name:", plan.GetExternalName()},
		{"Description:", plan.GetDescription()},
		{"UUID:", string(plan.GetName())},
		{"Status:", plan.GetShortStatus()},
		{"Free:", strconv.FormatBool(plan.GetFree())},
		{"Class:", class.Spec.ExternalName},
	})

	t.Render()
}

// WriteDefaultProvisionParameters prints the default provision parameters for a single plan.
func WriteDefaultProvisionParameters(w io.Writer, plan servicecatalog.Plan) {
	defaultProvisionParameters := plan.GetDefaultProvisionParameters()

	if defaultProvisionParameters != nil {
		fmt.Fprintln(w, "\nDefault Provision Parameters:")
		writeYAML(w, defaultProvisionParameters, 2)
	}
}

// WritePlanSchemas prints the schemas for a single plan.
func WritePlanSchemas(w io.Writer, plan servicecatalog.Plan) {
	instanceCreateSchema := plan.GetInstanceCreateSchema()
	instanceUpdateSchema := plan.GetInstanceUpdateSchema()
	bindingCreateSchema := plan.GetBindingCreateSchema()

	if instanceCreateSchema != nil {
		fmt.Fprintln(w, "\nInstance Create Parameter Schema:")
		writeYAML(w, instanceCreateSchema, 2)
	}

	if instanceUpdateSchema != nil {
		fmt.Fprintln(w, "\nInstance Update Parameter Schema:")
		writeYAML(w, instanceUpdateSchema, 2)
	}

	if bindingCreateSchema != nil {
		fmt.Fprintln(w, "\nBinding Create Parameter Schema:")
		writeYAML(w, bindingCreateSchema, 2)
	}
}
