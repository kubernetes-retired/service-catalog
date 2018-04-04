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

package controller

import (
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// Cluster service plan handlers and control-loop

func (c *controller) clusterServicePlanAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("ClusterServicePlan: Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.clusterServicePlanQueue.Add(key)
}

func (c *controller) clusterServicePlanUpdate(oldObj, newObj interface{}) {
	c.clusterServicePlanAdd(newObj)
}

func (c *controller) clusterServicePlanDelete(obj interface{}) {
	clusterServicePlan, ok := obj.(*v1beta1.ClusterServicePlan)
	if clusterServicePlan == nil || !ok {
		return
	}

	glog.V(4).Infof("ClusterServicePlan: Received delete event for %v; no further processing will occur", clusterServicePlan.Name)
}

// reconcileClusterServicePlanKey reconciles a ClusterServicePlan due to resync
// or an event on the ClusterServicePlan.  Note that this is NOT the main
// reconciliation loop for ClusterServicePlans. ClusterServicePlans are
// primarily reconciled in a separate flow when a ClusterServiceBroker is
// reconciled.
func (c *controller) reconcileClusterServicePlanKey(key string) error {
	plan, err := c.clusterServicePlanLister.Get(key)
	if errors.IsNotFound(err) {
		glog.Infof("ClusterServicePlan %q: Not doing work because it has been deleted", key)
		return nil
	}
	if err != nil {
		glog.Infof("ClusterServicePlan %q: Unable to retrieve object from store: %v", key, err)
		return err
	}

	return c.reconcileClusterServicePlan(plan)
}

func (c *controller) reconcileClusterServicePlan(clusterServicePlan *v1beta1.ClusterServicePlan) error {
	glog.Infof("ClusterServicePlan %q (ExternalName: %q): processing", clusterServicePlan.Name, clusterServicePlan.Spec.ExternalName)

	if !clusterServicePlan.Status.RemovedFromBrokerCatalog {
		return nil
	}

	glog.Infof("ClusterServicePlan %q (ExternalName: %q): has been removed from broker catalog; determining whether there are instances remaining", clusterServicePlan.Name, clusterServicePlan.Spec.ExternalName)

	serviceInstances, err := c.findServiceInstancesOnClusterServicePlan(clusterServicePlan)
	if err != nil {
		return err
	}

	if len(serviceInstances.Items) != 0 {
		return nil
	}

	glog.Infof("ClusterServicePlan %q (ExternalName: %q): has been removed from broker catalog and has zero instances remaining; deleting", clusterServicePlan.Name, clusterServicePlan.Spec.ExternalName)
	return c.serviceCatalogClient.ClusterServicePlans().Delete(clusterServicePlan.Name, &metav1.DeleteOptions{})
}

func (c *controller) findServiceInstancesOnClusterServicePlan(clusterServicePlan *v1beta1.ClusterServicePlan) (*v1beta1.ServiceInstanceList, error) {
	fieldSet := fields.Set{
		"spec.clusterServicePlanRef.name": clusterServicePlan.Name,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := metav1.ListOptions{FieldSelector: fieldSelector}

	return c.serviceCatalogClient.ServiceInstances(metav1.NamespaceAll).List(listOpts)
}
