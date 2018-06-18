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
