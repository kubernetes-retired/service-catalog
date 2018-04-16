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

	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func getBindingStatusCondition(status v1beta1.ServiceBindingStatus) v1beta1.ServiceBindingCondition {
	if len(status.Conditions) > 0 {
		return status.Conditions[len(status.Conditions)-1]
	}
	return v1beta1.ServiceBindingCondition{}
}

func getBindingStatusShort(status v1beta1.ServiceBindingStatus) string {
	lastCond := getBindingStatusCondition(status)
	return formatStatusShort(string(lastCond.Type), lastCond.Status, lastCond.Reason)
}

func getBindingStatusFull(status v1beta1.ServiceBindingStatus) string {
	lastCond := getBindingStatusCondition(status)
	return formatStatusFull(string(lastCond.Type), lastCond.Status, lastCond.Reason, lastCond.Message, lastCond.LastTransitionTime)
}

func writeBindingListJSON(w io.Writer, bindings []v1beta1.ServiceBinding) {
	l := v1beta1.ServiceBindingList{
		Items: bindings,
	}
	j, _ := json.MarshalIndent(l, "", "   ")

	w.Write(j)
}

func writeBindingListYAML(w io.Writer, bindings []v1beta1.ServiceBinding) {
	l := v1beta1.ServiceBindingList{
		Items: bindings,
	}
	y, _ := yaml.Marshal(l)

	w.Write(y)
}

func writeBindingListTable(w io.Writer, bindings []v1beta1.ServiceBinding) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Instance",
		"Status",
	})

	for _, binding := range bindings {
		t.Append([]string{
			binding.Name,
			binding.Namespace,
			binding.Spec.ServiceInstanceRef.Name,
			getBindingStatusShort(binding.Status),
		})
	}
	t.Render()
}

// WriteBindingList prints a list of bindings in the specified output format.
func WriteBindingList(w io.Writer, outputFormat string, bindings ...v1beta1.ServiceBinding) {
	switch outputFormat {
	case "json":
		writeBindingListJSON(w, bindings)
	case "yaml":
		writeBindingListYAML(w, bindings)
	case "table":
		writeBindingListTable(w, bindings)
	}
}

func writeBindingJSON(w io.Writer, binding v1beta1.ServiceBinding) {
	j, _ := json.MarshalIndent(binding, "", "   ")
	w.Write(j)
}

func writeBindingYAML(w io.Writer, binding v1beta1.ServiceBinding) {
	y, _ := yaml.Marshal(binding)
	w.Write(y)
}

// WriteBinding prints a single bindings in the specified output format.
func WriteBinding(w io.Writer, outputFormat string, binding v1beta1.ServiceBinding) {
	switch outputFormat {
	case "json":
		writeBindingJSON(w, binding)
	case "yaml":
		writeBindingYAML(w, binding)
	case "table":
		writeBindingListTable(w, []v1beta1.ServiceBinding{binding})
	}
}

// WriteBindingDetails prints details for a single binding.
func WriteBindingDetails(w io.Writer, binding *v1beta1.ServiceBinding) {
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", binding.Name},
		{"Namespace:", binding.Namespace},
		{"Status:", getBindingStatusFull(binding.Status)},
		{"Instance:", binding.Spec.ServiceInstanceRef.Name},
	})
	t.Render()

	writeParameters(w, binding.Spec.Parameters)
}

// WriteAssociatedBindings prints a list of bindings associated with an instance.
func WriteAssociatedBindings(w io.Writer, bindings []v1beta1.ServiceBinding) {
	fmt.Fprintln(w, "\nBindings:")
	if len(bindings) == 0 {
		fmt.Fprintln(w, "No bindings defined")
		return
	}

	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Status",
	})
	for _, binding := range bindings {
		t.Append([]string{
			binding.Name,
			getBindingStatusShort(binding.Status),
		})
	}
	t.Render()
}

// WriteDeletedBindingNames prints the names of a list of bindings
func WriteDeletedBindingNames(w io.Writer, bindings []v1beta1.ServiceBinding) {
	for _, binding := range bindings {
		WriteDeletedResourceName(w, binding.Name)
	}
}
