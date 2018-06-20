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

	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func (c *controller) serviceClassAdd(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.serviceClassQueue.Add(key)
}

func (c *controller) serviceClassUpdate(oldObj, newObj interface{}) {
	c.serviceClassAdd(newObj)
}

func (c *controller) serviceClassDelete(obj interface{}) {
	serviceClass, ok := obj.(*v1beta1.ServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for ServiceClass %v; no further processing will occur", serviceClass.Name)
}

// reconcileServiceClassKey reconciles a ServiceClass due to controller resync
// or an event on the ServiceClass.  Note that this is NOT the main
// reconciliation loop for ServiceClass. ServiceClasses are primarily
// reconciled in a separate flow when a ServiceBroker is reconciled.
func (c *controller) reconcileServiceClassKey(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	pcb := pretty.NewContextBuilder(pretty.ServiceClass, namespace, name, "")
	class, err := c.serviceClassLister.ServiceClasses(namespace).Get(name)
	if errors.IsNotFound(err) {
		glog.Info(pcb.Message("Not doing work because the ServiceClass has been deleted"))
		return nil
	}
	if err != nil {
		glog.Infof(pcb.Message("Unable to retrieve"))
		return err
	}

	return c.reconcileServiceClass(class)
}

func (c *controller) reconcileServiceClass(serviceClass *v1beta1.ServiceClass) error {
	pcb := pretty.NewContextBuilder(pretty.ServiceClass, serviceClass.Namespace, serviceClass.Name, "")
	glog.Info(pcb.Message("Processing"))

	if !serviceClass.Status.RemovedFromBrokerCatalog {
		return nil
	}

	glog.Info(pcb.Message("Removed from broker catalog; determining whether there are instances remaining"))

	serviceInstances, err := c.findServiceInstancesOnServiceClass(serviceClass)
	if err != nil {
		return err
	}

	if len(serviceInstances.Items) != 0 {
		return nil
	}

	glog.Info(pcb.Message("Removed from broker catalog and has zero instances remaining; deleting"))
	return c.serviceCatalogClient.ServiceClasses(serviceClass.Namespace).Delete(serviceClass.Name, &metav1.DeleteOptions{})
}

func (c *controller) findServiceInstancesOnServiceClass(serviceClass *v1beta1.ServiceClass) (*v1beta1.ServiceInstanceList, error) {
	fieldSet := fields.Set{
		"spec.serviceClassRef.name": serviceClass.Name,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := metav1.ListOptions{FieldSelector: fieldSelector}

	return c.serviceCatalogClient.ServiceInstances(serviceClass.Namespace).List(listOpts)
}
