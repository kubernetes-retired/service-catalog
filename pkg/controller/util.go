package controller

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func checkBindingCondition(sb *servicecatalog.Binding, bc servicecatalog.BindingCondition, ct servicecatalog.ConditionStatus) bool {
	return false
}
