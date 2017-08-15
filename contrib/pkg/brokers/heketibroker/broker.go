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

package heketibroker

import (
	//"fmt"
	//"sync"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokers/broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

// CreateBroker initializes the service broker.  This function is called by server.Start()
func CreateBroker() broker.Broker {
	var instanceMap = make(map[string]*heketiServiceInstance)
	return &heketiBroker{
		instanceMap: instanceMap,
	}
}

type heketiServiceInstance struct{}

type heketiBroker struct {
	instanceMap map[string]*heketiServiceInstance
}

func (b *heketiBroker) Catalog() (*brokerapi.Catalog, error) {
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name: "GlusterFS Object Storage",
				ID: "gfs-object-storage",
				Description: "A bucket of GlusterFS Object Storage.",
				Bindable: true,
			},
		},
	}, nil
}

func (b *heketiBroker) GetServiceInstanceLastOperation(instanceID, serviceID, planID, operation string) (*brokerapi.LastOperationResponse, error) {
	glog.Info("GetServiceInstanceLastOperation not yet implemented.")
	return nil, nil
}
func (b *heketiBroker) CreateServiceInstance(instanceID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Info("CreateServiceInstance not yet implemented.")
	return nil, nil
}
func (b *heketiBroker) RemoveServiceInstance(instanceID, serviceID, planID string, acceptsIncomplete bool) (*brokerapi.DeleteServiceInstanceResponse, error) {
	glog.Info("RemoveServiceInstance not yet implemented.")
	return nil, nil
}

func (b *heketiBroker) Bind(instanceID, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	glog.Info("Bind not yet implemented.")
	return nil, nil
}
func (b *heketiBroker) UnBind(instanceID, bindingID, serviceID, planID string) error {
	glog.Info("UnBind not yet implemented.")
	return nil
}
