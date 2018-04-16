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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
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
		"UUID"})
	for _, plan := range plans {
		t.Append([]string{
			plan.Spec.ExternalName,
			classNames[plan.Spec.ClusterServiceClassRef.Name],
			plan.Spec.Description,
			plan.Name})
	}
	t.Render()
}

func writePlanListJSON(w io.Writer, plans []v1beta1.ClusterServicePlan) {
	list := v1beta1.ClusterServicePlanList{
		Items: plans,
	}
	j, _ := json.MarshalIndent(list, "", "   ")
	w.Write(j)
}

func writePlanListYAML(w io.Writer, plans []v1beta1.ClusterServicePlan) {
	list := v1beta1.ClusterServicePlanList{
		Items: plans,
	}
	y, _ := yaml.Marshal(list)
	w.Write(y)
}

// WritePlanList prints a list of plans in the specified output format.
func WritePlanList(w io.Writer, outputFormat string, plans []v1beta1.ClusterServicePlan, classes []v1beta1.ClusterServiceClass) {
	classNames := map[string]string{}
	for _, class := range classes {
		classNames[class.Name] = class.Spec.ExternalName
	}
	switch outputFormat {
	case "json":
		writePlanListJSON(w, plans)
	case "yaml":
		writePlanListJSON(w, plans)
	case "table":
		writePlanListTable(w, plans, classNames)
	}
}

func writePlanJSON(w io.Writer, plan v1beta1.ClusterServicePlan) {
	j, _ := json.MarshalIndent(plan, "", "   ")
	w.Write(j)
}

func writePlanYAML(w io.Writer, plan v1beta1.ClusterServicePlan) {
	y, _ := yaml.Marshal(plan)
	w.Write(y)
}

// WritePlan prints a single plan in the specified output format.
func WritePlan(w io.Writer, outputFormat string, plan v1beta1.ClusterServicePlan, class v1beta1.ClusterServiceClass) {

	switch outputFormat {
	case "json":
		writePlanJSON(w, plan)
	case "yaml":
		writePlanJSON(w, plan)
	case "table":
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
