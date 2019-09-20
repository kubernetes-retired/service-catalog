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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

// Service class handlers and control-loop

func (c *controller) clusterServiceClassAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.clusterServiceClassQueue.Add(key)
}

func (c *controller) clusterServiceClassUpdate(oldObj, newObj interface{}) {
	c.clusterServiceClassAdd(newObj)
}

func (c *controller) clusterServiceClassDelete(obj interface{}) {
	serviceClass, ok := obj.(*v1beta1.ClusterServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	klog.V(4).Infof("Received delete event for ServiceClass %v; no further processing will occur", serviceClass.Name)
}

// reconcileServiceClassKey reconciles a ClusterServiceClass due to controller resync
// or an event on the ClusterServiceClass.  Note that this is NOT the main
// reconciliation loop for ClusterServiceClass. ClusterServiceClasses are primarily
// reconciled in a separate flow when a ClusterServiceBroker is reconciled.
func (c *controller) reconcileClusterServiceClassKey(key string) error {
	class, err := c.clusterServiceClassLister.Get(key)
	if errors.IsNotFound(err) {
		klog.Infof("ClusterServiceClass %q: Not doing work because it has been deleted", key)
		return nil
	}
	if err != nil {
		klog.Infof("ClusterServiceClass %q: Unable to retrieve object from store: %v", key, err)
		return err
	}

	return c.reconcileClusterServiceClass(class)
}

func (c *controller) reconcileClusterServiceClass(serviceClass *v1beta1.ClusterServiceClass) error {
	klog.Infof("ClusterServiceClass %q (ExternalName: %q): processing", serviceClass.Name, serviceClass.Spec.ExternalName)

	if !serviceClass.Status.RemovedFromBrokerCatalog {
		return nil
	}

	klog.Infof("ClusterServiceClass %q (ExternalName: %q): has been removed from broker catalog; determining whether there are instances remaining", serviceClass.Name, serviceClass.Spec.ExternalName)

	serviceInstances, err := c.findServiceInstancesOnClusterServiceClass(serviceClass)
	if err != nil {
		return err
	}
	klog.Infof("Found %d ServiceInstances", len(serviceInstances.Items))

	if len(serviceInstances.Items) != 0 {
		return nil
	}

	klog.Infof("ClusterServiceClass %q (ExternalName: %q): has been removed from broker catalog and has zero instances remaining; deleting", serviceClass.Name, serviceClass.Spec.ExternalName)
	return c.serviceCatalogClient.ClusterServiceClasses().Delete(serviceClass.Name, &metav1.DeleteOptions{})
}

func (c *controller) findServiceInstancesOnClusterServiceClass(serviceClass *v1beta1.ClusterServiceClass) (*v1beta1.ServiceInstanceList, error) {
	labelSelector := labels.SelectorFromSet(labels.Set{
		v1beta1.GroupName + "/" + v1beta1.FilterSpecClusterServiceClassRefName: serviceClass.Name,
	}).String()

	listOpts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	return c.serviceCatalogClient.ServiceInstances(metav1.NamespaceAll).List(listOpts)
}
