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

package controller

import (
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testServiceClassGUID  = "SCGUID"
	testServicePlanGUID   = "SPGUID"
	testServiceBrokerName = "test-servicebroker"
	testServiceClassName  = "test-serviceclass"
	testServicePlanName   = "test-serviceplan"
)

func getTestCommonServiceBrokerSpec() v1beta1.CommonServiceBrokerSpec {
	return v1beta1.CommonServiceBrokerSpec{
		URL:            "https://example.com",
		RelistBehavior: v1beta1.ServiceBrokerRelistBehaviorDuration,
		RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
	}
}

func getTestServiceBroker() *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServiceBrokerName,
			Namespace: testNamespace,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: getTestCommonServiceBrokerSpec(),
		},
	}
}

func getTestCommonServiceClassSpec() v1beta1.CommonServiceClassSpec {
	return v1beta1.CommonServiceClassSpec{
		Description:  "a test service",
		ExternalName: testServiceClassName,
		ExternalID:   testServiceClassGUID,
		Bindable:     true,
	}
}

func getTestServiceClass() *v1beta1.ServiceClass {
	return &v1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServiceClassGUID,
			Namespace: testNamespace,
		},
		Spec: v1beta1.ServiceClassSpec{
			ServiceBrokerName:      testServiceBrokerName,
			CommonServiceClassSpec: getTestCommonServiceClassSpec(),
		},
	}
}

func getTestCommonServicePlanSpec() v1beta1.CommonServicePlanSpec {
	return v1beta1.CommonServicePlanSpec{
		ExternalID:   testServicePlanGUID,
		ExternalName: testServicePlanName,
		Bindable:     truePtr(),
	}
}

func getTestServicePlan() *v1beta1.ServicePlan {
	return &v1beta1.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testServicePlanGUID,
			Namespace: testNamespace,
		},
		Spec: v1beta1.ServicePlanSpec{
			ServiceBrokerName:     testServiceBrokerName,
			CommonServicePlanSpec: getTestCommonServicePlanSpec(),
			ServiceClassRef: v1beta1.LocalObjectReference{
				Name: testServiceClassGUID,
			},
		},
		Status: v1beta1.ServicePlanStatus{},
	}
}

func getTestServiceInstanceWithNamespacedPlanReference() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testServiceInstanceName,
			Namespace:  testNamespace,
			Generation: 1,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: testServiceClassName,
				ServicePlanExternalName:  testServicePlanName,
			},
			ExternalID: testServiceInstanceGUID,
		},
		Status: v1beta1.ServiceInstanceStatus{
			DeprovisionStatus: v1beta1.ServiceInstanceDeprovisionStatusRequired,
		},
	}
}

func getTestServiceInstanceAsyncProvisioningWithNamespacedRefs(operation string) *v1beta1.ServiceInstance {
	instance := getTestServiceInstanceWithNamespacedRefs()

	operationStartTime := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	instance.Status = v1beta1.ServiceInstanceStatus{
		Conditions: []v1beta1.ServiceInstanceCondition{{
			Type:               v1beta1.ServiceInstanceConditionReady,
			Status:             v1beta1.ConditionFalse,
			Message:            "Provisioning",
			LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
		}},
		AsyncOpInProgress:  true,
		OperationStartTime: &operationStartTime,
		CurrentOperation:   v1beta1.ServiceInstanceOperationProvision,
		InProgressProperties: &v1beta1.ServiceInstancePropertiesState{
			ServicePlanExternalName: testServicePlanName,
			ServicePlanExternalID:   testServicePlanGUID,
		},
		ObservedGeneration: instance.Generation,
		DeprovisionStatus:  v1beta1.ServiceInstanceDeprovisionStatusRequired,
	}
	if operation != "" {
		instance.Status.LastOperation = &operation
	}

	return instance
}

func getTestServiceInstanceAsyncDeprovisioningWithNamespacedRefs(operation string) *v1beta1.ServiceInstance {
	instance := getTestServiceInstanceWithNamespacedRefs()
	instance.Generation = 2

	operationStartTime := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	instance.Status = v1beta1.ServiceInstanceStatus{
		Conditions: []v1beta1.ServiceInstanceCondition{{
			Type:               v1beta1.ServiceInstanceConditionReady,
			Status:             v1beta1.ConditionFalse,
			Message:            "Deprovisioning",
			LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
		}},
		AsyncOpInProgress:  true,
		OperationStartTime: &operationStartTime,
		CurrentOperation:   v1beta1.ServiceInstanceOperationDeprovision,
		InProgressProperties: &v1beta1.ServiceInstancePropertiesState{
			ServicePlanExternalName: testServicePlanName,
			ServicePlanExternalID:   testServicePlanGUID,
		},

		ReconciledGeneration: 1,
		ObservedGeneration:   2,
		ExternalProperties: &v1beta1.ServiceInstancePropertiesState{
			ServicePlanExternalName: testServicePlanName,
			ServicePlanExternalID:   testServicePlanGUID,
		},
		ProvisionStatus:   v1beta1.ServiceInstanceProvisionStatusProvisioned,
		DeprovisionStatus: v1beta1.ServiceInstanceDeprovisionStatusRequired,
	}
	if operation != "" {
		instance.Status.LastOperation = &operation
	}

	// Set the deleted timestamp to simulate deletion
	ts := metav1.NewTime(time.Now().Add(-5 * time.Minute))
	instance.DeletionTimestamp = &ts
	return instance
}

func getTestServiceInstanceWithNamespacedRefsAndStatus(status v1beta1.ConditionStatus) *v1beta1.ServiceInstance {
	instance := getTestServiceInstanceWithNamespacedRefsAndExternalProperties()
	instance.Status.Conditions = []v1beta1.ServiceInstanceCondition{{
		Type:               v1beta1.ServiceInstanceConditionReady,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
	}}
	return instance
}

func getTestServiceInstanceWithNamespacedRefsAndExternalProperties() *v1beta1.ServiceInstance {
	sc := getTestServiceInstanceWithNamespacedRefs()
	sc.Status.ExternalProperties = &v1beta1.ServiceInstancePropertiesState{
		ServicePlanExternalID:   testServicePlanGUID,
		ServicePlanExternalName: testServicePlanName,
	}
	return sc
}

func getTestBindingRetrievableServiceClass() *v1beta1.ServiceClass {
	return &v1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{Name: testServiceClassGUID},
		Spec: v1beta1.ServiceClassSpec{
			ServiceBrokerName: testServiceBrokerName,
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				Description:        "a test service",
				ExternalName:       testServiceClassName,
				ExternalID:         testServiceClassGUID,
				BindingRetrievable: true,
				Bindable:           true,
			},
		},
	}
}
