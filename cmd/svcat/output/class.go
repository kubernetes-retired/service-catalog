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
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func getClassStatusText(status v1beta1.ClusterServiceClassStatus) string {
	if status.RemovedFromBrokerCatalog {
		return statusDeprecated
	}
	return statusActive
}

func writeClassListTable(w io.Writer, classes []v1beta1.ClusterServiceClass) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Broker",
		"Description",
	})
	for _, class := range classes {
		t.Append([]string{
			class.Spec.ExternalName,
			class.Spec.ClusterServiceBrokerName,
			class.Spec.Description,
		})
	}
	t.Render()
}

// WriteClassList prints a list of classes in the specified output format.
func WriteClassList(w io.Writer, outputFormat string, classes ...v1beta1.ClusterServiceClass) {
	classList := v1beta1.ClusterServiceClassList{
		Items: classes,
	}
	switch outputFormat {
	case formatJSON:
		writeJSON(w, classList)
	case formatYAML:
		writeYAML(w, classList, 0)
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
		writeClassListTable(w, []v1beta1.ClusterServiceClass{class})
	}
}

// WriteParentClass prints identifying information for a parent class.
func WriteParentClass(w io.Writer, class *v1beta1.ClusterServiceClass) {
	fmt.Fprintln(w, "\nClass:")
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", class.Spec.ExternalName},
		{"UUID:", string(class.Name)},
		{"Status:", getClassStatusText(class.Status)},
	})
	t.Render()
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
