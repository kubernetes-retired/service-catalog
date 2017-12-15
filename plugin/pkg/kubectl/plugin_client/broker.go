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

package plugin_client

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *PluginClient) GetBroker(brokerName string) (*v1beta1.ClusterServiceBroker, error) {
	broker, err := c.ScClient.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, v1.GetOptions{})
	return broker, err
}

func (c *PluginClient) ListBrokers() (*v1beta1.ClusterServiceBrokerList, error) {
	brokers, err := c.ScClient.ServicecatalogV1beta1().ClusterServiceBrokers().List(v1.ListOptions{})
	return brokers, err
}
