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

// WriteInstanceList prints a list of instances.
func WriteInstanceList(w io.Writer, instances ...v1beta1.ServiceInstance) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"Namespace",
		"Class",
		"Plan",
		"Status",
	})

	for _, instance := range instances {
		t.Append([]string{
			instance.Name,
			instance.Namespace,
			instance.Spec.ClusterServiceClassExternalName,
			instance.Spec.ClusterServicePlanExternalName,
			getInstanceStatusShort(instance.Status),
		})
	}

	t.Render()
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
		{"Class:", instance.Spec.ClusterServiceClassExternalName},
		{"Plan:", instance.Spec.ClusterServicePlanExternalName},
	})

	t.Render()
}
