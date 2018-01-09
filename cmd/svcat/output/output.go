package output

import (
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	statusActive     = "Active"
	statusDeprecated = "Deprecated"
)

func formatStatusShort(condition string, conditionStatus v1beta1.ConditionStatus, reason string) string {
	if conditionStatus == v1beta1.ConditionTrue {
		return condition
	}
	return reason
}

func formatStatusFull(condition string, conditionStatus v1beta1.ConditionStatus, reason string, message string, timestamp v1.Time) string {
	status := formatStatusShort(condition, conditionStatus, reason)
	if status == "" {
		return ""
	}
	message = strings.TrimRight(message, ".")
	return fmt.Sprintf("%s - %s @ %s", status, message, timestamp)
}
