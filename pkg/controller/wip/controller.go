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

package wip

import (
	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/client/cache"
	k8sclientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/util/runtime"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

// NewController returns a new Open Service Broker catalog
// controller.
func NewController(
	kubeClient k8sclientset.Interface,
	serviceCatalogClient servicecatalogclientset.Interface,
	brokerInformer cache.SharedInformer,
	serviceClassInformer cache.SharedInformer,
	instanceInformer cache.SharedInformer,
	bindingInformer cache.SharedInformer,
) (Controller, error) {
	controller := &controller{
		kubeClient:           kubeClient,
		serviceCatalogClient: serviceCatalogClient,
		brokerInformer:       brokerInformer,
		serviceClassInformer: serviceClassInformer,
		instanceInformer:     instanceInformer,
		bindingInformer:      bindingInformer,
	}

	brokerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.brokerAdd,
		UpdateFunc: controller.brokerUpdate,
		DeleteFunc: controller.brokerDelete,
	})

	serviceClassInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.serviceClassAdd,
		UpdateFunc: controller.serviceClassUpdate,
		DeleteFunc: controller.serviceClassDelete,
	})

	instanceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.instanceAdd,
		UpdateFunc: controller.instanceUpdate,
		DeleteFunc: controller.instanceDelete,
	})

	bindingInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.bindingAdd,
		UpdateFunc: controller.bindingUpdate,
		DeleteFunc: controller.bindingDelete,
	})

	return controller, nil
}

// Controller describes a controller that backs the service catalog API for
// Open Service Broker compliant Brokers.
type Controller interface {
	// Run runs the controller until the given stop channel can be read from.
	Run(stopCh <-chan struct{})
}

type controller struct {
	kubeClient           k8sclientset.Interface
	serviceCatalogClient servicecatalogclientset.Interface
	brokerInformer       cache.SharedInformer
	serviceClassInformer cache.SharedInformer
	instanceInformer     cache.SharedInformer
	bindingInformer      cache.SharedInformer
}

func (c *controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	glog.Info("Starting service-catalog controller")

	<-stopCh
	glog.Info("Shutting down service-catalog controller")
}

func (c *controller) brokerAdd(obj interface{}) {
	broker, ok := obj.(*v1alpha1.Broker)
	if broker == nil || !ok {
		return
	}

	c.reconcileBroker(broker)
}

func (c *controller) brokerUpdate(oldObj, newObj interface{}) {
	c.brokerAdd(newObj)
}

func (c *controller) brokerDelete(obj interface{}) {
	broker, ok := obj.(*v1alpha1.Broker)
	if broker == nil || !ok {
		return
	}

	glog.V(4).Info("Received delete event for Broker %v", broker.Name)
}

func (c *controller) reconcileBroker(broker *v1alpha1.Broker) {
	glog.V(4).Infof("Processing Broker %v", broker.Name)
}

func (c *controller) serviceClassAdd(obj interface{}) {
	serviceClass, ok := obj.(*v1alpha1.ServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	c.reconcileServiceClass(serviceClass)
}

func (c *controller) serviceClassUpdate(oldObj, newObj interface{}) {
	c.serviceClassAdd(newObj)
}

func (c *controller) reconcileServiceClass(serviceClass *v1alpha1.ServiceClass) {
	glog.V(4).Infof("Processing ServiceClass %v", serviceClass.Name)
}

func (c *controller) serviceClassDelete(obj interface{}) {
	serviceClass, ok := obj.(*v1alpha1.ServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for ServiceClass %v", serviceClass.Name)
}

func (c *controller) instanceAdd(obj interface{}) {
	instance, ok := obj.(*v1alpha1.Instance)
	if instance == nil || !ok {
		return
	}

	c.reconcileInstance(instance)
}

func (c *controller) instanceUpdate(oldObj, newObj interface{}) {
	c.instanceAdd(newObj)
}

func (c *controller) reconcileInstance(instance *v1alpha1.Instance) {
	glog.V(4).Infof("Processing Instance %v", instance.Name)
}

func (c *controller) instanceDelete(obj interface{}) {
	instance, ok := obj.(*v1alpha1.Instance)
	if instance == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for Instance %v", instance.Name)
}

func (c *controller) bindingAdd(obj interface{}) {
	binding, ok := obj.(*v1alpha1.Binding)
	if binding == nil || !ok {
		return
	}

	c.reconcileBinding(binding)
}

func (c *controller) bindingUpdate(oldObj, newObj interface{}) {
	c.bindingAdd(newObj)
}

func (c *controller) reconcileBinding(binding *v1alpha1.Binding) {
	glog.V(4).Infof("Processing Binding %v", binding.Name)
}

func (c *controller) bindingDelete(obj interface{}) {
	binding, ok := obj.(*v1alpha1.Binding)
	if binding == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for Binding %v", binding.Name)
}
