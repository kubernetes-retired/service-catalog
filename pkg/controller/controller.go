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

package controller

import (
	"errors"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient/tpr"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/watch"
	"k8s.io/client-go/1.5/kubernetes"
	k8swatch "k8s.io/client-go/1.5/pkg/watch"
)

// Controller is an instance of the service catalog controller.
type Controller struct {
	controller *controller
}

// New creates an instance of the service catalog Controller.
func New(k8sClient *kubernetes.Clientset, w *watch.Watcher, inj injector.BindingInjector, brokerClCreator brokerapi.CreateFunc) (*Controller, error) {
	h := createHandler(k8sClient, tpr.NewAPIClient(w), inj, brokerClCreator)
	return &Controller{
		controller: createController(h, w),
	}, nil
}

// Run starts the controller
func (c *Controller) Run() {
	c.controller.Run()
}

type controller struct {
	handler Handler
	watcher *watch.Watcher
}

func createController(h Handler, w *watch.Watcher) *controller {
	return &controller{
		handler: h,
		watcher: w,
	}
}

func (c *controller) Run() error {
	if c.watcher == nil {
		glog.Infoln("No watcher (was nil), so not interfacing with kubernetes directly")
		return errors.New("No watcher (was nil)")
	}

	glog.Infoln("Starting to watch for new Service Brokers")
	c.watcher.Watch(watch.ServiceBroker, "default", c.serviceBrokerCallback)

	glog.Infoln("Starting to watch for new Service Instances")
	c.watcher.Watch(watch.ServiceInstance, "default", c.serviceInstanceCallback)

	glog.Infoln("Starting to watch for new Service Bindings")
	c.watcher.Watch(watch.ServiceBinding, "default", c.serviceBindingCallback)

	return nil
}

func (c *controller) serviceInstanceCallback(e k8swatch.Event) error {
	var si servicecatalog.Instance
	err := util.TPRObjectToSCObject(e.Object, &si)
	if err != nil {
		glog.Errorf("Failed to decode the received object %#v", err)
	}

	if e.Type == k8swatch.Added {
		created, err := c.handler.CreateServiceInstance(&si)
		if err != nil {
			glog.Errorf("Failed to create service instance: %v\n", err)
			return err
		}
		glog.Infof("Created Service Instance: %s\n", created.Name)
	} else {
		glog.Warningf("Received unsupported service instance event type %s", e.Type)
	}
	return nil
}

func (c *controller) serviceBindingCallback(e k8swatch.Event) error {
	var sb servicecatalog.Binding
	err := util.TPRObjectToSCObject(e.Object, &sb)
	if err != nil {
		glog.Errorf("Failed to decode the received object %#v", err)
	}

	switch {
	case e.Type == k8swatch.Added:
		created, err := c.handler.CreateServiceBinding(&sb)
		if err != nil {
			glog.Errorf("Failed to create service binding: %v\n", err)
			return err
		}
		glog.Infof("Created Service Binding: %s\n%v\n", sb.Name, created)
	case e.Type == k8swatch.Deleted:
		// if we've already processed the deletion event, ignore
		if sb.DeletionTimestamp != nil {
			return nil
		}

		if err := c.handler.DeleteServiceBinding(&sb); err != nil {
			return err
		}
	case e.Type == k8swatch.Modified && sb.DeletionTimestamp != nil:
		// do nothing, the binding is in the process of deletion
		glog.Info("binding was modified during deletion, ignoring")
	default:
		glog.Warningf("Received unsupported service binding event type %s", e.Type)
	}
	return nil
}

func (c *controller) serviceBrokerCallback(e k8swatch.Event) error {
	var sb servicecatalog.Broker
	err := util.TPRObjectToSCObject(e.Object, &sb)
	if err != nil {
		glog.Errorf("Failed to decode the received object %#v", err)
		return err
	}

	if e.Type == k8swatch.Added {
		created, err := c.handler.CreateServiceBroker(&sb)
		if err != nil {
			glog.Errorf("Failed to create service broker: %v\n", err)
			return err
		}
		glog.Infof("Created Service Broker: %s\n", created.Name)
	} else {
		glog.Warningf("Received unsupported service broker event type %s", e.Type)
	}
	return nil
}
