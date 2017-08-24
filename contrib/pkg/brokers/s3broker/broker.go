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

package s3broker

import (
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokers/broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/minio/minio-go"
)

// CreateBroker initializes the service broker.  This function is called by server.Start()
func CreateBroker() broker.Broker {
	// TODO temp use of minio import so it can be vendored in.
	_, _ = minio.NewV2("", "", "", false)

	var instanceMap = make(map[string]*s3ServiceInstance)
	return &s3Broker{
		instanceMap: instanceMap,
	}
}

type s3ServiceInstance struct{}

type s3Broker struct {
	instanceMap map[string]*s3ServiceInstance
}

func (b *s3Broker) Catalog() (*brokerapi.Catalog, error) {
	isBindable := true
	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "gluster-object-store",
				ID:          "lkdgkf2napdwedom",
				Description: "An object bucket backed by GlusterFS Object Storage.",
				Bindable:    true,
				Plans: []brokerapi.ServicePlan {
					{
						Name: "default",
						ID: "0",
						Description: "The best plan, and the only one.",
						Free: true,
						Bindable: &isBindable,
					},
				},
			},
		},
	}, nil
}

func (b *s3Broker) GetServiceInstanceLastOperation(instanceID, serviceID, planID, operation string) (*brokerapi.LastOperationResponse, error) {
	glog.Info("GetServiceInstanceLastOperation not yet implemented.")
	return nil, nil
}
func (b *s3Broker) CreateServiceInstance(instanceID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Info("CreateServiceInstance not yet implemented.")
	return nil, nil
}
func (b *s3Broker) RemoveServiceInstance(instanceID, serviceID, planID string, acceptsIncomplete bool) (*brokerapi.DeleteServiceInstanceResponse, error) {
	glog.Info("RemoveServiceInstance not yet implemented.")
	return nil, nil
}

func (b *s3Broker) Bind(instanceID, bindingID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	glog.Info("Bind not yet implemented.")
	return nil, nil
}
func (b *s3Broker) UnBind(instanceID, bindingID, serviceID, planID string) error {
	glog.Info("UnBind not yet implemented.")
	return nil
}
