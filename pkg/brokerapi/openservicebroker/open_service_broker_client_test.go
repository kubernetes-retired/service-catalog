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

package openservicebroker

import (
	"testing"

	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func TestUpdateServiceInstance(t *testing.T) {
	cli := NewClient(&servicecatalog.Broker{})

	_, err := cli.UpdateServiceInstance("foo", &model.ServiceInstanceRequest{})
	if err == nil {
		t.Fatalf("Expected not implemented")
	}
	if err.Error() != "Not implemented" {
		t.Errorf("Expected not implemented, got %v", err)
	}

	// TODO: test against fake/test broker.
}
