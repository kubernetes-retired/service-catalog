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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatsdk "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"k8s.io/api/core/v1"
)

func getBindingStatusShort(status v1beta1.ServiceBindingStatus) string {
	lastCond := svcatsdk.GetBindingStatusCondition(status)
	return formatStatusShort(string(lastCond.Type), lastCond.Status, lastCond.Reason)
}

func getBindingStatusFull(status v1beta1.ServiceBindingStatus) string {
	lastCond := svcatsdk.GetBindingStatusCondition(status)
	return formatStatusFull(string(lastCond.Type), lastCond.Status, lastCond.Reason, lastCond.Message, lastCond.LastTransitionTime)
}

func writeBindingListTable(w io.Writer, bindingList *v1beta1.ServiceBindingList) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Instance",
		"Status",
	})

	for _, binding := range bindingList.Items {
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
func WriteBindingList(w io.Writer, outputFormat string, bindingList *v1beta1.ServiceBindingList) {
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, bindingList)
	case FormatYAML:
		writeYAML(w, bindingList, 0)
	case FormatTable:
		writeBindingListTable(w, bindingList)
	}
}

// WriteBinding prints a single bindings in the specified output format.
func WriteBinding(w io.Writer, outputFormat string, binding v1beta1.ServiceBinding) {
	switch outputFormat {
	case FormatJSON:
		writeJSON(w, binding)
	case FormatYAML:
		writeYAML(w, binding, 0)
	case FormatTable:
		l := v1beta1.ServiceBindingList{
			Items: []v1beta1.ServiceBinding{binding},
		}
		writeBindingListTable(w, &l)
	}
}

// WriteBindingDetails prints details for a single binding.
func WriteBindingDetails(w io.Writer, binding *v1beta1.ServiceBinding) {
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", binding.Name},
		{"Namespace:", binding.Namespace},
		{"Status:", getBindingStatusFull(binding.Status)},
		{"Secret:", binding.Spec.SecretName},
		{"Instance:", binding.Spec.ServiceInstanceRef.Name},
	})
	t.Render()

	writeParameters(w, binding.Spec.Parameters)
	writeParametersFrom(w, binding.Spec.ParametersFrom)
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

// WriteAssociatedSecret prints the secret data associated with a binding.
func WriteAssociatedSecret(w io.Writer, secret *v1.Secret, err error, showSecrets bool) {
	// Don't print anything the secret isn't ready yet
	if secret == nil && err == nil {
		return
	}

	fmt.Fprintln(w, "\nSecret Data:")
	if err != nil {
		// We should have been able to find a secret but couldn't for some reason,
		// warn about it without blowing up the entire command.
		fmt.Fprintf(w, "  %s", err.Error())
		return
	}

	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	t := NewDetailsTable(w)
	for _, key := range keys {
		value := secret.Data[key]
		if showSecrets {
			t.Append([]string{key, string(value)})
		} else {
			t.Append([]string{key, fmt.Sprintf("%d bytes", len(value))})
		}
	}
	t.Render()
}

// WriteDeletedBindingNames prints the names of a list of bindings
func WriteDeletedBindingNames(w io.Writer, bindings []v1beta1.ServiceBinding) {
	for _, binding := range bindings {
		WriteDeletedResourceName(w, binding.Name)
	}
}
