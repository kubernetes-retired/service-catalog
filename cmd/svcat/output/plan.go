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
type byClass []v1beta1.ClusterServicePlan

func (a byClass) Len() int {
	return len(a)
}
func (a byClass) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a byClass) Less(i, j int) bool {
	return a[i].Spec.ClusterServiceClassRef.Name < a[j].Spec.ClusterServiceClassRef.Name
}

func writePlanListTable(w io.Writer, plans []v1beta1.ClusterServicePlan, classNames map[string]string) {

	sort.Sort(byClass(plans))

	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Class",
		"Description",
	})

	// Pre-parse the data so we allow the last column to be really wide.
	// We can't set the MinWidth after data has been loaded
	maxNameWidth := len("Name")
	maxClassWidth := len("Class")
	maxDescWidth := len("Description")

	for _, plan := range plans {
		if tmp := len(plan.Spec.ExternalName); tmp > maxNameWidth {
			maxNameWidth = tmp
		}
		tmp := len(classNames[plan.Spec.ClusterServiceClassRef.Name])
		if tmp > maxClassWidth {
			maxClassWidth = tmp
		}
		if tmp := len(plan.Spec.Description); tmp > maxDescWidth {
			maxDescWidth = tmp
		}
	}
	if tmp := (80 - (maxNameWidth + maxClassWidth + 11)); tmp > maxDescWidth {
		t.SetColMinWidth(2, tmp)
	}

	for _, plan := range plans {
		t.Append([]string{
			plan.Spec.ExternalName,
			classNames[plan.Spec.ClusterServiceClassRef.Name],
			plan.Spec.Description,
		})
	}

	t.Render()
}

// WritePlanList prints a list of plans in the specified output format.
func WritePlanList(w io.Writer, outputFormat string, plans []v1beta1.ClusterServicePlan, classes []servicecatalog.Class) {
	classNames := map[string]string{}
	for _, class := range classes {
		classNames[class.GetName()] = class.GetExternalName()
	}
	list := v1beta1.ClusterServicePlanList{
		Items: plans,
	}
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, list)
	case FormatYAML:
		writeYAML(w, list, 0)
	case FormatTable:
		writePlanListTable(w, plans, classNames)
	}
}

// WritePlan prints a single plan in the specified output format.
func WritePlan(w io.Writer, outputFormat string, plan v1beta1.ClusterServicePlan, class v1beta1.ClusterServiceClass) {

	switch outputFormat {
	case FormatJSON:
		writeJSON(w, plan)
	case FormatYAML:
		writeYAML(w, plan, 0)
	case FormatTable:
		classNames := map[string]string{}
		classNames[class.Name] = class.Spec.ExternalName
		writePlanListTable(w, []v1beta1.ClusterServicePlan{plan}, classNames)
	}
}

// WriteAssociatedPlans prints a list of plans associated with a class.
func WriteAssociatedPlans(w io.Writer, plans []v1beta1.ClusterServicePlan) {
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
			plan.Spec.ExternalName,
			plan.Spec.Description,
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
func WritePlanDetails(w io.Writer, plan *v1beta1.ClusterServicePlan, class *v1beta1.ClusterServiceClass) {
	t := NewDetailsTable(w)

	t.AppendBulk([][]string{
		{"Name:", plan.Spec.ExternalName},
		{"Description:", plan.Spec.Description},
		{"UUID:", string(plan.Name)},
		{"Status:", getPlanStatusShort(plan.Status)},
		{"Free:", strconv.FormatBool(plan.Spec.Free)},
		{"Class:", class.Spec.ExternalName},
	})

	t.Render()
}

// WritePlanSchemas prints the schemas for a single plan.
func WritePlanSchemas(w io.Writer, plan *v1beta1.ClusterServicePlan) {
	instanceCreateSchema := plan.Spec.ServiceInstanceCreateParameterSchema
	instanceUpdateSchema := plan.Spec.ServiceInstanceUpdateParameterSchema
	bindingCreateSchema := plan.Spec.ServiceBindingCreateParameterSchema

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
