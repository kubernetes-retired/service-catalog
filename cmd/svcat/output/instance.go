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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func getInstanceStatusCondition(status v1beta1.ServiceInstanceStatus) v1beta1.ServiceInstanceCondition {
	if len(status.Conditions) > 0 {
		return status.Conditions[len(status.Conditions)-1]
	}
	return v1beta1.ServiceInstanceCondition{}
}

func getInstanceStatusFull(status v1beta1.ServiceInstanceStatus) string {
	lastCond := getInstanceStatusCondition(status)
	return formatStatusFull(string(lastCond.Type), lastCond.Status, lastCond.Reason, lastCond.Message, lastCond.LastTransitionTime)
}

func getInstanceStatusShort(status v1beta1.ServiceInstanceStatus) string {
	lastCond := getInstanceStatusCondition(status)
	return formatStatusShort(string(lastCond.Type), lastCond.Status, lastCond.Reason)
}

func writeInstanceListTable(w io.Writer, instanceList *v1beta1.ServiceInstanceList) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Class",
		"Plan",
		"Status",
	})

	for _, instance := range instanceList.Items {
		t.Append([]string{
			instance.Name,
			instance.Namespace,
			instance.Spec.GetSpecifiedClass(),
			instance.Spec.GetSpecifiedPlan(),
			getInstanceStatusShort(instance.Status),
		})
	}

	t.Render()
}

// WriteInstanceList prints a list of instances.
func WriteInstanceList(w io.Writer, outputFormat string, instanceList *v1beta1.ServiceInstanceList) {
	switch outputFormat {
	case formatJSON:
		writeJSON(w, instanceList)
	case formatYAML:
		writeYAML(w, instanceList, 0)
	case formatTable:
		writeInstanceListTable(w, instanceList)
	}
}

// WriteInstance prints a single instance
func WriteInstance(w io.Writer, outputFormat string, instance v1beta1.ServiceInstance) {
	switch outputFormat {
	case "json":
		writeJSON(w, instance)
	case "yaml":
		writeYAML(w, instance, 0)
	case "table":
		p := v1beta1.ServiceInstanceList{
			Items: []v1beta1.ServiceInstance{instance},
		}
		writeInstanceListTable(w, &p)
	}
}

// WriteParentInstance prints identifying information for a parent instance.
func WriteParentInstance(w io.Writer, instance *v1beta1.ServiceInstance) {
	fmt.Fprintln(w, "\nInstance:")
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", instance.Name},
		{"Namespace:", instance.Namespace},
		{"Status:", getInstanceStatusShort(instance.Status)},
	})
	t.Render()
}

// WriteAssociatedInstances prints a list of instances associated with a plan.
func WriteAssociatedInstances(w io.Writer, instances []v1beta1.ServiceInstance) {
	fmt.Fprintln(w, "\nInstances:")
	if len(instances) == 0 {
		fmt.Fprintln(w, "No instances defined")
		return
	}

	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Status",
	})
	for _, instance := range instances {
		t.Append([]string{
			instance.Name,
			instance.Namespace,
			getInstanceStatusShort(instance.Status),
		})
	}
	t.Render()
}

// WriteInstanceDetails prints an instance.
func WriteInstanceDetails(w io.Writer, instance *v1beta1.ServiceInstance) {
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", instance.Name},
		{"Namespace:", instance.Namespace},
		{"Status:", getInstanceStatusFull(instance.Status)},
		{"Class:", instance.Spec.GetSpecifiedClass()},
		{"Plan:", instance.Spec.GetSpecifiedPlan()},
	})
	t.Render()

	writeParameters(w, instance.Spec.Parameters)
	writeParametersFrom(w, instance.Spec.ParametersFrom)
}
