package output

import (
	"fmt"
	"io"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func getBrokerStatusCondition(status v1beta1.ClusterServiceBrokerStatus) v1beta1.ServiceBrokerCondition {
	if len(status.Conditions) > 0 {
		return status.Conditions[len(status.Conditions)-1]
	}
	return v1beta1.ServiceBrokerCondition{}
}

func getBrokerStatusShort(status v1beta1.ClusterServiceBrokerStatus) string {
	lastCond := getBrokerStatusCondition(status)
	return formatStatusShort(string(lastCond.Type), lastCond.Status, lastCond.Reason)
}

func getBrokerStatusFull(status v1beta1.ClusterServiceBrokerStatus) string {
	lastCond := getBrokerStatusCondition(status)
	return formatStatusFull(string(lastCond.Type), lastCond.Status, lastCond.Reason, lastCond.Message, lastCond.LastTransitionTime)
}

// WriteBrokerList prints a list of brokers.
func WriteBrokerList(w io.Writer, brokers ...v1beta1.ClusterServiceBroker) {
	t := NewListTable(w)
	t.SetHeader([]string{
		"Name",
		"URL",
		"Status",
	})
	for _, broker := range brokers {
		t.Append([]string{
			broker.Name,
			broker.Spec.URL,
			getBrokerStatusShort(broker.Status),
		})
	}
	t.Render()
}

// WriteParentBroker prints identifying information for a parent broker.
func WriteParentBroker(w io.Writer, broker *v1beta1.ClusterServiceBroker) {
	fmt.Fprintln(w, "\nBroker:")
	t := NewDetailsTable(w)
	t.AppendBulk([][]string{
		{"Name:", broker.Name},
		{"Status:", getBrokerStatusShort(broker.Status)},
	})
	t.Render()
}

// WriteBrokerDetails prints details for a single broker.
func WriteBrokerDetails(w io.Writer, broker *v1beta1.ClusterServiceBroker) {
	t := NewDetailsTable(w)

	t.AppendBulk([][]string{
		{"Name:", broker.Name},
		{"URL:", broker.Spec.URL},
		{"Status:", getBrokerStatusFull(broker.Status)},
	})

	t.Render()
}
