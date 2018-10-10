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
	"io"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
)

func getScope(class servicecatalog.Class) string {
	if class.GetNamespace() != "" {
		return servicecatalog.NamespaceScope
	}
	return servicecatalog.ClusterScope
}

func writeClassListTable(w io.Writer, classes []servicecatalog.Class) {
	t := NewListTable(w)

	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Description",
	})
	t.SetVariableColumn(3)

	for _, class := range classes {
		t.Append([]string{
			class.GetExternalName(),
			class.GetNamespace(),
			class.GetDescription(),
		})
	}

	t.Render()
}

// WriteClassList prints a list of classes in the specified output format.
func WriteClassList(w io.Writer, outputFormat string, classes ...servicecatalog.Class) {
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, classes)
	case FormatYAML:
		writeYAML(w, classes, 0)
	case FormatTable:
		writeClassListTable(w, classes)
	}
}

// WriteClass prints a single class in the specified output format.
func WriteClass(w io.Writer, outputFormat string, class servicecatalog.Class) {
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, class)
	case FormatYAML:
		writeYAML(w, class, 0)
	case FormatTable:
		writeClassListTable(w, []servicecatalog.Class{class})
	}
}

// WriteClassDetails prints details for a single class.
func WriteClassDetails(w io.Writer, class servicecatalog.Class) {
	scope := getScope(class)
	spec := class.GetSpec()
	t := NewDetailsTable(w)
	t.Append([]string{"Name:", spec.ExternalName})
	if class.GetNamespace() != "" {
		t.Append([]string{"Namespace:", class.GetNamespace()})
	}
	t.AppendBulk([][]string{
		{"Scope:", scope},
		{"Description:", spec.Description},
		{"UUID:", class.GetName()},
		{"Status:", class.GetStatusText()},
		{"Tags:", strings.Join(spec.Tags, ", ")},
		{"Broker:", class.GetServiceBrokerName()},
	})
	t.Render()
}

// WriteClassAndPlanDetails prints details for multiple classes and plans
func WriteClassAndPlanDetails(w io.Writer, classes []servicecatalog.Class, plans [][]servicecatalog.Plan) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Class",
		"Plans",
		"Description",
	})
	for i, class := range classes {
		for i, plan := range plans[i] {
			if i == 0 {
				t.Append([]string{
					class.GetExternalName(),
					plan.GetName(),
					class.GetSpec().Description,
				})
			} else {
				t.Append([]string{
					"",
					plan.GetName(),
					"",
				})
			}
		}
	}
	t.table.SetAutoWrapText(true)
	t.SetVariableColumn(3)
	t.Render()
}
