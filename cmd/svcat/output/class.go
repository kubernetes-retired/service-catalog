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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
)

func getClassStatusText(status v1beta1.ClusterServiceClassStatus) string {
	if status.RemovedFromBrokerCatalog {
		return statusDeprecated
	}
	return statusActive
}

func writeClassListTable(w io.Writer, classes []servicecatalog.Class) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Description",
	})
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
	case formatJSON:
		writeJSON(w, classes)
	case formatYAML:
		writeYAML(w, classes, 0)
	case formatTable:
		writeClassListTable(w, classes)
	}
}

// WriteClass prints a single class in the specified output format.
func WriteClass(w io.Writer, outputFormat string, class v1beta1.ClusterServiceClass) {
	switch outputFormat {
	case formatJSON:
		writeJSON(w, class)
	case formatYAML:
		writeYAML(w, class, 0)
	case formatTable:
		writeClassListTable(w, []servicecatalog.Class{&class})
	}
}

// WriteClassDetails prints details for a single class.
func WriteClassDetails(w io.Writer, class *v1beta1.ClusterServiceClass) {
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", class.Spec.ExternalName},
		{"Description:", class.Spec.Description},
		{"UUID:", string(class.Name)},
		{"Status:", getClassStatusText(class.Status)},
		{"Tags:", strings.Join(class.Spec.Tags, ", ")},
		{"Broker:", class.Spec.ClusterServiceBrokerName},
	})
	t.Render()
}
