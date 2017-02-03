/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient/mem"
	kapi "k8s.io/kubernetes/pkg/api"
)

func TestGetServicePlanInfoEmpty(t *testing.T) {
	apiclient := mem.NewAPIClient()
	if _, _, _, err := GetServicePlanInfo(apiclient.ServiceClasses(), "service", ""); err == nil {
		t.Error("Expected err on GetServicePlanInfo with no plan with empty apiclient")
	}
	if _, _, _, err := GetServicePlanInfo(apiclient.ServiceClasses(), "service", "plan"); err == nil {
		t.Error("Expected err on GetServicePlanInfo with plan with empty apiclient")
	}
}

// test service classes with single plan
func TestGetServicePlanInfoOne(t *testing.T) {
	apiclient := mem.NewAPIClient()
	svcClss := createServiceClassesWithServicePlans(2, 1)
	apiclient.ServiceClasses().Create(&svcClss[0])

	// test that getting with empty plan name works
	if err := testGetServicePlanInfo(apiclient, &svcClss[0], &svcClss[0].Plans[0], ""); err != nil {
		t.Error(err)
	}

	// test that getting with correct plan name works
	if err := testGetServicePlanInfo(apiclient, &svcClss[0], &svcClss[0].Plans[0],
		svcClss[0].Plans[0].Name); err != nil {

		t.Error(err)
	}

	// test that using wrong plan name is wrong
	if err := testGetServicePlanInfo(apiclient, &svcClss[0], &svcClss[0].Plans[0],
		"wrong_name"); err == nil {

		t.Error("Expected err when using wrong plan name")
	}

	// test that using wrong service name is wrong
	if err := testGetServicePlanInfo(apiclient, &svcClss[1], &svcClss[0].Plans[0],
		svcClss[0].Plans[0].Name); err == nil {

		t.Error("Expected err when using wrong service name")
	}
}

// test service classes with multiple plans
func TestGetServicePlanInfoMany(t *testing.T) {
	apiclient := mem.NewAPIClient()
	svcClss := createServiceClassesWithServicePlans(3, 3)

	for i := range svcClss {
		apiclient.ServiceClasses().Create(&svcClss[i])
	}

	for _, svcCls := range svcClss {
		for _, plan := range svcCls.Plans {
			//test that getting empty plan name doesn't work
			if err := testGetServicePlanInfo(apiclient, &svcCls, &plan, ""); err == nil {
				t.Error("Expected err when using empty plan name")
			}

			//test that getting wrong plan name doesn't work
			if err := testGetServicePlanInfo(apiclient, &svcCls, &plan, "not here!"); err == nil {
				t.Error("Expected err when using wrong plan name")
			}

			//test that getting correct plan name works
			if err := testGetServicePlanInfo(apiclient, &svcCls, &plan, plan.Name); err != nil {

				t.Error("Unexpected err when trying to get service plan with multiple plans:", err)
			}
		}
	}
}

func TestGetBrokerByServiceClassName(t *testing.T) {
	apiclient := mem.NewAPIClient()

	brokers := make([]servicecatalog.Broker, 3)
	for i := range brokers {
		brokers[i] = servicecatalog.Broker{
			ObjectMeta: kapi.ObjectMeta{
				Name: "brokerName" + string(i),
			},
		}
		apiclient.Brokers().Create(&brokers[i])
	}

	svcClss := createServiceClassesWithServicePlans(2, 0)
	for _, svcCls := range svcClss {
		svcCls.BrokerName = brokers[0].Name
		apiclient.ServiceClasses().Create(&svcCls)
	}

	for _, svcCls := range svcClss {
		if _, err := GetBrokerByServiceClassName(apiclient.Brokers(), apiclient.ServiceClasses(),
			svcCls.Name); err != nil {

			t.Error("Unexpected err on GetBrokerByServiceClassName")
		}
	}

	if _, err := GetBrokerByServiceClassName(apiclient.Brokers(), apiclient.ServiceClasses(),
		"name_not_there"); err == nil {

		t.Error("No err from GetBrokerByServiceClassName on service class with wrong broker name")
	}
}

func TestGetBrokerByServiceClassNameEmpty(t *testing.T) {
	apiclient := mem.NewAPIClient()

	if _, err := GetBrokerByServiceClassName(apiclient.Brokers(), apiclient.ServiceClasses(),
		"some_name"); err == nil {
		t.Error("No err from GetBrokerByServiceClassName with empty apiclient")
	}
}

func createServiceClassesWithServicePlans(svcClsCount int, planCount int) []servicecatalog.ServiceClass {

	ret := make([]servicecatalog.ServiceClass, svcClsCount)
	for i := range ret {
		ret[i] = servicecatalog.ServiceClass{
			ObjectMeta: kapi.ObjectMeta{
				Name: "serviceName" + string(i),
			},
			OSBGUID: "serviceGuid" + string(i),
			Plans:   createServicePlans(planCount, planCount*i),
		}
	}
	return ret
}

func createServicePlans(n int, startCount int) []servicecatalog.ServicePlan {
	ret := make([]servicecatalog.ServicePlan, n)
	for i := range ret {
		ret[i] = servicecatalog.ServicePlan{
			Name:    "planName" + string(i+startCount),
			OSBGUID: "planGUID" + string(i+startCount),
		}
	}
	return ret
}

func testGetServicePlanInfo(apiclient apiclient.APIClient,
	svcCls *servicecatalog.ServiceClass, plan *servicecatalog.ServicePlan,
	queryPlanName string) error {

	svcGUID, planGUID, planName, err := GetServicePlanInfo(apiclient.ServiceClasses(),
		svcCls.Name, queryPlanName)
	if err != nil {
		return fmt.Errorf("Err from GetServicePlanInfo: %v", err)
	}
	if svcCls.OSBGUID != svcGUID {
		return fmt.Errorf("Service GUIDs don't match. Expected %s, got %s",
			svcCls.OSBGUID, svcGUID)
	}
	if plan.OSBGUID != planGUID {
		return fmt.Errorf("Plan GUIDs don't match. Expected %s, got %s",
			plan.OSBGUID, planGUID)
	}
	if plan.Name != planName {
		return fmt.Errorf("Plan names don't match. Expected %s, got %s",
			plan.Name, planName)
	}
	return nil
}
