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
	"strconv"

	"github.com/kubernetes-incubator/service-catalog/controller/watch"

	"github.com/gorilla/mux"
)

// Server is an instance of the service controller server.
type Server struct {
	controller *controller
	port       int
	k8sHandler *k8sHandler
}

// CreateServer creates an instance of the service controller server.
func CreateServer(serviceStorage ServiceStorage, port int, w *watch.Watcher) (*Server, error) {
	k8sStorage := NewThirdPartyServiceStorage(w)
	c := createController(k8sStorage)
	k8sHandler, err := createK8sHandler(c, w)
	if err != nil {
		log.Printf("Couldn't create the k8s native handler, watcher will not be installed: %v\n", err)
		return nil, err
	}
	return &Server{
		controller: c,
		port:       port,
		k8sHandler: k8sHandler,
	}, nil
}

// Start starts the server and begins listening on a TCP port.
func (s *Server) Start() {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", healthZHandler).Methods("GET")

	port := strconv.Itoa(s.port)
	log.Println("Server started on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	log.Println(err.Error())
}

func healthZHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
