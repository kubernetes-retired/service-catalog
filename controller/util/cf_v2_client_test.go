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
	"net/http"
	"net/http/httptest"
	"testing"

	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

func TestCreateServiceInstance(t *testing.T) {
	fakeBroker := newFakeBroker()
	defer fakeBroker.Close()
	const id = "testID"
	req := &model.ServiceInstanceRequest{
		OrgID:             "testOrgID",
		PlanID:            "testPlanID",
		ServiceID:         "testSvcID",
		SpaceID:           "testSpaceID",
		Parameters:        map[string]interface{}{},
		AcceptsIncomplete: false,
	}

	cl := CreateCFV2BrokerClient(&scmodel.ServiceBroker{
		BrokerURL: fakeBroker.URLStr(),
	})

	_, err := cl.CreateServiceInstance(id, req)
	if err != nil {
		t.Fatalf("error in create service instance (%s)", err)
	}
}

func TestUpdateServiceInstance(t *testing.T) {
	cl := CreateCFV2BrokerClient(&scmodel.ServiceBroker{})
	_, err := cl.UpdateServiceInstance("foo", &model.ServiceInstanceRequest{})
	if err == nil {
		t.Fatalf("Expected not implemented")
	}
	if err.Error() != "Not implemented" {
		t.Errorf("Expected not implemented, got %v", err)
	}
}
