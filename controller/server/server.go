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
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kubernetes-incubator/service-catalog/controller/watch"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	controller  *Controller
	httpHandler *HttpHandler
	port        int
	k8sHandler  *K8sHandler
}

func CreateServer(serviceStorage ServiceStorage, port int, w *watch.Watcher) (*Server, error) {
	k8sStorage := NewThirdPartyServiceStorage(w)
	c := CreateController(k8sStorage)
	k8sHandler, err := NewK8sHandler(c, w)
	if err != nil {
		log.Printf("Couldn't create the k8s native handler, watcher will not be installed: %v\n", err)
		return nil, err
	}
	return &Server{
		controller:  c,
		port:        port,
		httpHandler: NewHttpHandler(c, k8sStorage),
		k8sHandler:  k8sHandler,
	}, nil
}

func (s *Server) Start() {
	router := mux.NewRouter()

	// TODO: the actual inventory API should be /v2/services[/...] and
	// /v2/service_plans[/...].
	router.HandleFunc("/v2/service_plans", s.controller.Inventory).Methods("GET")

	// Broker related stuff
	router.HandleFunc("/v2/service_brokers", s.controller.ListServiceBrokers).Methods("GET")
	router.HandleFunc("/v2/service_brokers", s.httpHandler.CreateServiceBroker).Methods("POST")
	router.HandleFunc("/v2/service_brokers/{broker}", s.controller.GetServiceBroker).Methods("GET")
	router.HandleFunc("/v2/service_brokers/{broker}", s.controller.DeleteServiceBroker).Methods("DELETE")
	router.HandleFunc("/v2/service_brokers/{broker}:refresh", s.controller.RefreshServiceBroker).Methods("POST")
	// TODO: implement updating a service broker.
	// router.HandleFunc("/v2/service_brokers/{broker_id}", s.Controller.UpdateServiceBroker).Methods.("PUT")

	router.HandleFunc("/v2/service_instances", s.controller.ListServiceInstances).Methods("GET")
	router.HandleFunc("/v2/service_instances", s.httpHandler.CreateServiceInstance).Methods("POST")
	router.HandleFunc("/v2/service_instances/{service}", s.controller.GetServiceInstance).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service}", s.controller.DeleteServiceInstance).Methods("DELETE")
	// TODO: implement list service bindings for this service instance.
	// router.HandleFunc("/v2/service_instances/{service_id}/service_bindings", s.controller.ListServiceInstanceBindings).Methods("GET")

	router.HandleFunc("/v2/service_bindings", s.controller.ListServiceBindings).Methods("GET")
	router.HandleFunc("/v2/service_bindings", s.httpHandler.CreateServiceBinding).Methods("POST")
	router.HandleFunc("/v2/service_bindings/{binding}", s.controller.GetServiceBinding).Methods("GET")
	router.HandleFunc("/v2/service_bindings/{binding}", s.controller.DeleteServiceBinding).Methods("DELETE")

	http.Handle("/", handlers.LoggingHandler(os.Stderr, router))

	port := strconv.Itoa(s.port)
	log.Println("Server started on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	log.Println(err.Error())
}
