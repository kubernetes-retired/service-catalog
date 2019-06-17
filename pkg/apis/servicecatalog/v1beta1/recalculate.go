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

package v1beta1

import "fmt"

// RecalculatePrinterColumnStatusFields sets column status fields using status conditions
func (in *ServiceBroker) RecalculatePrinterColumnStatusFields() {
	in.Status.LastConditionState = serviceBrokerLastConditionState(&in.Status.CommonServiceBrokerStatus)
}

// RecalculatePrinterColumnStatusFields sets column status fields using status conditions
func (in *ClusterServiceBroker) RecalculatePrinterColumnStatusFields() {
	in.Status.LastConditionState = serviceBrokerLastConditionState(&in.Status.CommonServiceBrokerStatus)
}

// RecalculatePrinterColumnStatusFields sets column status fields using status conditions
func (in *ServiceInstance) RecalculatePrinterColumnStatusFields() {
	var class, plan string
	if in.Spec.ClusterServiceClassSpecified() && in.Spec.ClusterServicePlanSpecified() {
		class = fmt.Sprintf("ClusterServiceClass/%s", in.Spec.GetSpecifiedClusterServiceClass())
		plan = in.Spec.GetSpecifiedClusterServicePlan()
	} else {
		class = fmt.Sprintf("ServiceClass/%s", in.Spec.GetSpecifiedServiceClass())
		plan = in.Spec.GetSpecifiedServicePlan()
	}
	in.Status.UserSpecifiedClassName = class
	in.Status.UserSpecifiedPlanName = plan

	in.Status.LastConditionState = getServiceInstanceLastConditionState(&in.Status)
}

// RecalculatePrinterColumnStatusFields sets column status fields using status conditions
func (in *ServiceBinding) RecalculatePrinterColumnStatusFields() {
	in.Status.LastConditionState = getServiceBindingLastConditionState(in.Status)
}

// IsUserSpecifiedClassOrPlan returns true if user specified class or plan is not empty
func (in *ServiceInstance) IsUserSpecifiedClassOrPlan() bool {
	return in.Status.UserSpecifiedPlanName != "" ||
		in.Status.UserSpecifiedClassName != ""
}

func getServiceInstanceLastConditionState(status *ServiceInstanceStatus) string {
	if len(status.Conditions) > 0 {
		condition := status.Conditions[len(status.Conditions)-1]
		if condition.Status == ConditionTrue {
			return string(condition.Type)
		}
		return condition.Reason
	}
	return ""
}

func serviceBrokerLastConditionState(status *CommonServiceBrokerStatus) string {
	if len(status.Conditions) > 0 {
		condition := status.Conditions[len(status.Conditions)-1]
		if condition.Status == ConditionTrue {
			return string(condition.Type)
		}
		return condition.Reason
	}
	return ""
}

func getServiceBindingLastConditionState(status ServiceBindingStatus) string {
	if len(status.Conditions) > 0 {
		condition := status.Conditions[len(status.Conditions)-1]
		if condition.Status == ConditionTrue {
			return string(condition.Type)
		}
		return condition.Reason
	}
	return ""
}
