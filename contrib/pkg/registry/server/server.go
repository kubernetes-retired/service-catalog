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

	"github.com/kubernetes-incubator/service-catalog/pkg/util"

	"github.com/gorilla/mux"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

type server struct {
	controller Controller
}

// CreateHandler creates registry HTTP handler based on an implementation
// of a Controller interface.
func CreateHandler(c Controller) http.Handler {
	s := server{
		controller: c,
	}

	router := mux.NewRouter()

	// Broker related stuff
	router.HandleFunc("/services", s.ListServices).Methods("GET")
	router.HandleFunc("/services", s.CreateService).Methods("POST")
	router.HandleFunc("/services/{service_id}", s.GetService).Methods("GET")
	router.HandleFunc("/services/{service_id}", s.DeleteService).Methods("DELETE")

	var handler http.Handler = router
	return handler
}

// Start creates the HTTP handler based on an implementation of a
// Controller interface, and begins to listen on the specified port.
func Start(serverPort int, c Controller) {
	log.Printf("Starting server on port %d\n", serverPort)
	http.Handle("/", CreateHandler(c))
	if err := http.ListenAndServe(":"+strconv.Itoa(serverPort), nil); err != nil {
		panic(err)
	}
}

func (c *server) ListServices(w http.ResponseWriter, r *http.Request) {
	l, err := c.controller.ListServices()
	if err != nil {
		util.WriteResponse(w, http.StatusBadRequest, err)
		return
	}
	util.WriteResponse(w, http.StatusOK, l)
}

func (c *server) GetService(w http.ResponseWriter, r *http.Request) {
	id := util.ExtractVarFromRequest(r, "service_id")
	log.Printf("GetService: %s", id)

	s, err := c.controller.GetService(id)
	if err != nil {
		log.Printf("Got Error: %#v", err)
		util.WriteResponse(w, http.StatusBadRequest, err)
	} else {
		util.WriteResponse(w, http.StatusOK, s)
	}
}

func (c *server) CreateService(w http.ResponseWriter, r *http.Request) {
	var s brokerapi.Service
	if err := util.BodyToObject(r, &s); err != nil {
		log.Printf("Error unmarshaling: %#v", err)
		util.WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := c.controller.CreateService(&s); err != nil {
		util.WriteResponse(w, http.StatusBadRequest, err)
	} else {
		util.WriteResponse(w, http.StatusOK, s)
	}
}

func (c *server) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := util.ExtractVarFromRequest(r, "service_id")

	if err := c.controller.DeleteService(id); err != nil {
		util.WriteResponse(w, http.StatusBadRequest, err)
	} else {
		util.WriteResponse(w, http.StatusOK, "")
	}
}
