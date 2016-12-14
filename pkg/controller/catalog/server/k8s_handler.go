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

package server

import (
	"errors"
	"log"

	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/watch"

	k8swatch "k8s.io/client-go/1.5/pkg/watch"
)

type k8sHandler struct {
	controller ServiceController
	watcher    *watch.Watcher
}

func createK8sHandler(c ServiceController, w *watch.Watcher) (*k8sHandler, error) {
	ret := &k8sHandler{
		controller: c,
		watcher:    w,
	}
	if w != nil {
		log.Println("Starting to watch for new Service Brokers")
		err := w.Watch(watch.ServiceBroker, "default", ret.serviceBrokerCallback)
		if err != nil {
			log.Printf("Failed to start a watcher for Service Brokers: %v\n", err)
		}

		log.Println("Starting to watch for new Service Instances")
		err = w.Watch(watch.ServiceInstance, "default", ret.serviceInstanceCallback)
		if err != nil {
			log.Printf("Failed to start a watcher for Service Instances: %v\n", err)
		}

		log.Println("Starting to watch for new Service Bindings")
		err = w.Watch(watch.ServiceBinding, "default", ret.serviceBindingCallback)
		if err != nil {
			log.Printf("Failed to start a watcher for Service Bindings: %v\n", err)
			return nil, err
		}
	} else {
		log.Println("No watcher (was nil), so not interfacing with kubernetes directly")
		return nil, errors.New("No watcher (was nil)")
	}

	return ret, nil
}

func (s *k8sHandler) serviceInstanceCallback(e k8swatch.Event) error {
	var si scmodel.ServiceInstance
	err := util.TPRObjectToSCObject(e.Object, &si)
	if err != nil {
		log.Printf("Failed to decode the received object %#v", err)
	}

	if e.Type == "ADDED" || e.Type == "MODIFIED" {
		if si.LastOperation != nil {
			log.Printf("%s already created, skipping: %v\n", si.Name, si.LastOperation)
			return nil
		}
		created, err := s.controller.CreateServiceInstance(&si)
		if err != nil {
			log.Printf("Failed to create service instance: %v\n", err)
			return err
		}
		log.Printf("Created Service Instance: %s\n", created.Name)
	}
	return nil
}

func (s *k8sHandler) serviceBindingCallback(e k8swatch.Event) error {
	var sb scmodel.ServiceBinding
	err := util.TPRObjectToSCObject(e.Object, &sb)
	if err != nil {
		log.Printf("Failed to decode the received object %#v", err)
	}

	if e.Type == "ADDED" || e.Type == "MODIFIED" {
		if sb.Credentials.Hostname != "" {
			log.Printf("Bindings have already been created, skipping: %s\n", sb.Name)
			return nil
		}
		created, err := s.controller.CreateServiceBinding(&sb)
		if err != nil {
			log.Printf("Failed to create service binding: %v\n", err)
			return err
		}
		log.Printf("Created Service Binding: %s\n%v\n", sb.Name, created)
	}
	return nil
}

func (s *k8sHandler) serviceBrokerCallback(e k8swatch.Event) error {
	var sb scmodel.ServiceBroker
	err := util.TPRObjectToSCObject(e.Object, &sb)
	if err != nil {
		log.Printf("Failed to decode the received object %#v", err)
		return err
	}

	if e.Type == "ADDED" || e.Type == "MODIFIED" {
		created, err := s.controller.CreateServiceBroker(&sb)
		if err != nil {
			log.Printf("Failed to create service broker: %v\n", err)
			return err
		}
		log.Printf("Created Service Broker: %s\n", created.Name)
	} else {
		return errors.New("Delete Service Broker not implemented yet")
	}
	return nil
}
