/*
Copyright 2017 The Kubernetes Authors.

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

package controller

import (
	"fmt"

	"github.com/golang/glog"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
)

// Conditions

// ConditionType can be translated to a ServiceInstanceCondition or
// ServiceBindingCondition value.
type ConditionType string

const (
	// ConditionReady represents that a given InstanceCondition is in
	// ready state.
	ConditionReady ConditionType = "Ready"

	// ConditionFailed represents information about a final failure
	// that should not be retried.
	ConditionFailed ConditionType = "Failed"
)

// setCondition sets a single condition on a type's status. This delegates to
// the correct type implementation.
//
// Note: objects coming from informers should never be mutated; always pass a
// deep copy as the instance parameter.
func setCondition(obj interface{},
	condition ConditionType,
	status v1beta1.ConditionStatus,
	reason,
	message string) {

	switch toUpdate := obj.(type) {
	case *v1beta1.ServiceInstance:
		var conditionType v1beta1.ServiceInstanceConditionType
		switch condition {
		case ConditionReady:
			conditionType = v1beta1.ServiceInstanceConditionReady
		case ConditionFailed:
			conditionType = v1beta1.ServiceInstanceConditionFailed
		default:
			glog.Errorf("Service Instance: unable to set set unknown condition: %+v", condition)
		}
		setServiceInstanceCondition(toUpdate, conditionType, status, reason, message)
		return
	case *v1beta1.ServiceBinding:
		var conditionType v1beta1.ServiceBindingConditionType
		switch condition {
		case ConditionReady:
			conditionType = v1beta1.ServiceBindingConditionReady
		case ConditionFailed:
			conditionType = v1beta1.ServiceBindingConditionFailed
		default:
			glog.Errorf("Service Binding: unable to set set unknown condition: %+v", condition)
		}
		setServiceBindingCondition(toUpdate, conditionType, status, reason, message)
		return
	default:
		glog.Errorf("attempting to set condition on object of unknown type: %+v", obj)
	}
}

// Status

// updateStatus inspects then updates status based on type
func (c *controller) updateStatus(obj interface{}, pcb *pretty.ContextBuilder) (interface{}, error) {
	switch toUpdate := obj.(type) {
	case *v1beta1.ServiceInstance:
		return c.updateServiceInstanceStatus(toUpdate)
	case *v1beta1.ServiceBinding:
		return c.updateServiceBindingStatus(toUpdate)
	default:
		s := fmt.Sprintf(
			"attempting to set status on object of unknown type: %+v",
			obj,
		)
		glog.Errorf(pcb.Message(s))
		return obj, fmt.Errorf(s)
	}
}
