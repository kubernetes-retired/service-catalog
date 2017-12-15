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

func (c *PluginClient) GetPlan(planName string) (*v1beta1.ClusterServicePlan, error) {
	plan, err := c.ScClient.ServicecatalogV1beta1().ClusterServicePlans().Get(planName, v1.GetOptions{})
	return plan, err
}

func (c *PluginClient) ListPlans() (*v1beta1.ClusterServicePlanList, error) {
	plans, err := c.ScClient.ServicecatalogV1beta1().ClusterServicePlans().List(v1.ListOptions{})
	return plans, err
}
