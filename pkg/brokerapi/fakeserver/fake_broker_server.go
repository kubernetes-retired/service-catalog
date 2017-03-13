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

package fakeserver

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/util"
)

// FakeBrokerServer is an http server that implements the Open Service Broker
// REST API.
type FakeBrokerServer struct {
	server *httptest.Server

	Catalog       *brokerapi.Catalog
	CatalogStatus *int

	sync.Mutex

	// actions and reactions

	// Actions is the list of actions that have been run against the
	// FakeBrokerServer.
	Actions []Action
	// ProvisionReactions define how provision requests should be treated.
	ProvisionReactions map[string]ProvisionReaction
	// DeprovisionReactions define how deprovision requests should be treated.
	DeprovisionReactions map[string]DeprovisionReaction
	// BindReactions define how bind requests should be treated.
	BindReactions map[string]BindReaction
	// UnbindReactions define how unbind requests should be treated.
	UnbindReactions map[string]UnbindReaction

	// state

	// ActiveProvisions is a map of the active provision reactions for async
	// provision requests that have been accepted.  The key is the operation
	// ID.
	ActiveProvisions map[string]ProvisionReaction
	// ActiveDeprovisions in a map of the active deprovision reactions for
	// async deprovision requests that have been accepted.  The key is the
	// operation ID.
	ActiveDeprovisions map[string]DeprovisionReaction
	// ActiveInstances is a map of instances that the broker has told the user
	// were correctly provisioned.
	ActiveInstances map[string]brokerapi.CreateServiceInstanceRequest
	// OriginatingProvisionRequests is a map of instance ID to the async
	// request for it to be provisioned.  Used to implement the correct
	// semantics when the same request is issued.
	OriginatingProvisionRequests map[string]brokerapi.CreateServiceInstanceRequest

	// old fields - remove
	responseStatus     int
	pollsRemaining     int
	shouldSucceedAsync bool
	operation          string

	// For inspecting on what was sent on the wire.
	RequestObject interface{}
	Request       *http.Request
}

func NewFakeBrokerServer() *FakeBrokerServer {
	return &FakeBrokerServer{
		Actions:                      []Action{},
		ProvisionReactions:           map[string]ProvisionReaction{},
		DeprovisionReactions:         map[string]DeprovisionReaction{},
		BindReactions:                map[string]BindReaction{},
		UnbindReactions:              map[string]UnbindReaction{},
		ActiveProvisions:             map[string]ProvisionReaction{},
		ActiveDeprovisions:           map[string]DeprovisionReaction{},
		ActiveInstances:              map[string]brokerapi.CreateServiceInstanceRequest{},
		OriginatingProvisionRequests: map[string]brokerapi.CreateServiceInstanceRequest{},
	}

}

type ProvisionReaction struct {
	Status   int
	Response *brokerapi.CreateServiceInstanceResponse

	Async       bool
	Operation   string
	Polls       int
	AsyncResult string
}

type DeprovisionReaction struct {
	Status   int
	Response brokerapi.DeleteServiceInstanceResponse

	Async       bool
	Operation   string
	Polls       int
	AsyncResult string
}

type BindReaction struct {
	Status   int
	Response brokerapi.CreateServiceBindingResponse
}

type UnbindReaction struct {
	Status int
}

type Action struct {
	Path    string
	Verb    string
	Request *http.Request
	Object  interface{}
}

const (
	// TODO: make all methods use instanceIDKey
	instanceIDKey = "id"

	bindingIDKey = "binding_id"
)

// Start starts the fake broker server listening on a random port, passing
// back the server's URL.
func (f *FakeBrokerServer) Start() string {
	router := mux.NewRouter()
	router.HandleFunc("/v2/catalog", f.catalogHandler).Methods("GET")
	router.HandleFunc("/v2/service_instances/{id}/last_operation", f.lastOperationHandler).Methods("GET")
	router.HandleFunc("/v2/service_instances/{id}", f.provisionHandler).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{id}", f.updateHandler).Methods("PATCH")
	router.HandleFunc("/v2/service_instances/{id}", f.deprovisionHandler).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", f.bindHandler).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", f.unbindHandler).Methods("DELETE")
	f.server = httptest.NewServer(router)
	return f.server.URL
}

// Stop shuts down the server.
func (f *FakeBrokerServer) Stop() {
	f.server.Close()
}

// SetResponseStatus sets the default response status of the broker to the
// given HTTP status code.
func (f *FakeBrokerServer) SetResponseStatus(status int) {
	f.responseStatus = status
}

// SetAsynchronous sets the number of polls before finished, final state, and
// operation for asynchronous operations.
func (f *FakeBrokerServer) SetAsynchronous(numPolls int, shouldSucceed bool, operation string) {
	f.pollsRemaining = numPolls
	f.shouldSucceedAsync = shouldSucceed
	f.operation = operation
}

// HANDLERS

func (f *FakeBrokerServer) catalogHandler(w http.ResponseWriter, r *http.Request) {
	if f.CatalogStatus != nil {
		util.WriteResponse(w, *f.CatalogStatus, nil)
	}

	util.WriteResponse(w, http.StatusOK, f.Catalog)
}

func (f *FakeBrokerServer) lastOperationHandler(w http.ResponseWriter, r *http.Request) {
	req := &brokerapi.LastOperationRequest{}
	if err := util.BodyToObject(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.RequestObject = req

	f.Lock()
	defer f.Unlock()

	id := mux.Vars(r)[instanceIDKey]

	// check active provisions
	activeProvision, ok := f.ActiveProvisions[req.Operation]
	if ok {
		glog.Infof("last_operation request for provision %v; remaining polls: %v", req.Operation, activeProvision.Polls)

		activeProvision.Polls -= 1
		f.ActiveProvisions[req.Operation] = activeProvision

		if activeProvision.Polls > 0 {
			util.WriteResponse(w, http.StatusOK, brokerapi.LastOperationResponse{
				State: brokerapi.StateInProgress,
			})
		} else {
			// TODO: add instance to list of active instances
			// if activeProvision.AsyncResult == brokerapi.StateSucceeded
			// 	f.ActiveInstances =
			delete(f.ActiveProvisions, req.Operation)

			glog.Infof("(provision) Returning response status %v to last_operation request %v", activeProvision.AsyncResult, req.Operation)

			util.WriteResponse(w, http.StatusOK, brokerapi.LastOperationResponse{
				State: activeProvision.AsyncResult,
			})
		}
		return
	}

	// check active deprovisions
	activeDeprovision, ok := f.ActiveDeprovisions[req.Operation]
	if ok {
		glog.Infof("last_operation request for deprovision %v; remaining polls: %v", req.Operation, activeDeprovision.Polls)

		activeDeprovision.Polls -= 1
		f.ActiveDeprovisions[req.Operation] = activeDeprovision

		if activeDeprovision.Polls > 0 {
			util.WriteResponse(w, http.StatusOK, brokerapi.LastOperationResponse{
				State: brokerapi.StateInProgress,
			})
		} else {
			delete(f.ActiveInstances, id)
			delete(f.ActiveDeprovisions, req.Operation)

			glog.Infof("Returning response status %v to last_operation request %v", activeDeprovision.AsyncResult, req.Operation)

			util.WriteResponse(w, http.StatusOK, brokerapi.LastOperationResponse{
				State: activeDeprovision.AsyncResult,
			})
		}
		return
	} else {
		glog.Info("Couldn't find active deprovision for %v", req.Operation)
	}

	util.WriteResponse(w, http.StatusInternalServerError, "shrug")
}

func (f *FakeBrokerServer) provisionHandler(w http.ResponseWriter, r *http.Request) {
	// create a new action for this call
	action := Action{
		Verb:    r.Method,
		Path:    r.RequestURI,
		Request: r,
	}

	// deserialize the request
	req := &brokerapi.CreateServiceInstanceRequest{}
	if err := util.BodyToObject(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.RequestObject = req
	action.Object = req

	f.Lock()
	defer f.Unlock()

	// store the action
	f.Actions = append(f.Actions, action)

	// find the reaction for this request
	id := mux.Vars(r)[instanceIDKey]
	reaction, ok := f.ProvisionReactions[id]
	if !ok {
		// TODO: what's the default response if there's no reaction defined?
	}

	activeRequest, instanceIsActive := f.ActiveInstances[id]
	if !reaction.Async && !req.AcceptsIncomplete {
		// In order to return an async response, the request must set the
		// `accepts_incomplete=true` param.
		// TODO: does our client actually implement sending this correctly?

		if instanceIsActive {
			if reflect.DeepEqual(req, activeRequest) {
				// we got the same request again; return a 200 and the
				// reaction response
				util.WriteResponse(w, http.StatusOK, reaction.Response)
			} else {
				// TODO: send correct conflict response body
				util.WriteResponse(w, http.StatusConflict, "{}")
			}

			return
		}

		// if the reaction has status OK or completed, record the request used
		// to create the instance
		if reaction.Status == http.StatusOK || reaction.Status == http.StatusCreated {
			f.ActiveInstances[id] = *req
		}

		util.WriteResponse(w, reaction.Status, reaction.Response)
	} else if reaction.Async && req.AcceptsIncomplete {
		// Asynchronous

		// we got the same request again; return a 200 and the reaction response
		if instanceIsActive && !reflect.DeepEqual(req, activeRequest) {
			util.WriteResponse(w, http.StatusOK, reaction.Response)
			return
		}

		// record the state of the async reaction
		f.ActiveProvisions[reaction.Operation] = reaction

		util.WriteResponse(w, reaction.Status, &brokerapi.CreateServiceInstanceResponse{
			Operation: reaction.Operation,
		})
	} else {
		// The reaction was supposed to be async, but we got a synchronous request.

		// TODO: send the expected 422 response body
		util.WriteResponse(w, http.StatusUnprocessableEntity, reaction.Response)
	}
}

func (f *FakeBrokerServer) deprovisionHandler(w http.ResponseWriter, r *http.Request) {
	action := Action{
		Verb:    r.Method,
		Path:    r.RequestURI,
		Request: r,
	}

	req := &brokerapi.DeleteServiceInstanceRequest{}
	if err := util.BodyToObject(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.RequestObject = req
	action.Object = req

	f.Lock()
	defer f.Unlock()

	// store the action
	f.Actions = append(f.Actions, action)

	id := mux.Vars(r)[instanceIDKey]
	reaction, ok := f.DeprovisionReactions[id]
	if !ok {
		// TODO: what's the default response if there's no reaction defined?
	}

	glog.Infof("Handling deprovision request for instance %v with reaction %#v", id, reaction)

	_, instanceIsActive := f.ActiveInstances[id]
	if !reaction.Async && !req.AcceptsIncomplete {
		if reaction.Status == http.StatusOK {
			if instanceIsActive {
				// if the reaction status is 'ok', and the instance is
				// currently active, delete it and send a 'success' response
				delete(f.ActiveInstances, id)
				util.WriteResponse(w, reaction.Status, &brokerapi.DeleteServiceInstanceResponse{})
				return
			} else {
				// if the reaction status is 'ok', and the instance isn't
				// currently active, send a 'gone' response
				util.WriteResponse(w, http.StatusGone, &brokerapi.DeleteServiceInstanceResponse{})
				return
			}
		}

		// Fall-through: if the reaction has a status that is not 'ok', return
		// it here.
		util.WriteResponse(w, reaction.Status, &brokerapi.DeleteServiceInstanceResponse{})
	} else if reaction.Async && req.AcceptsIncomplete {
		// Asynchronous

		if instanceIsActive {
			glog.Infof("instance %v is active; storing operation %v in active deprovisions", id, reaction.Operation)
			f.ActiveDeprovisions[reaction.Operation] = reaction

			util.WriteResponse(w, reaction.Status, &brokerapi.DeleteServiceInstanceResponse{
				reaction.Operation,
			})
		} else {
			glog.Infof("instance %v is not active; returning 'gone' response", id)
			// if the reaction status is 'ok', and the instance isn't
			// currently active, send a 'gone' response
			util.WriteResponse(w, http.StatusGone, &brokerapi.DeleteServiceInstanceResponse{})
		}
	} else {
		// The reaction was supposed to be async, but we got a synchronous
		// request.

		// TODO: send the expected 422 response body
		util.WriteResponse(w, http.StatusUnprocessableEntity, reaction.Response)
	}
}

func (f *FakeBrokerServer) updateHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	util.WriteResponse(w, http.StatusForbidden, nil)
}

func (f *FakeBrokerServer) bindHandler(w http.ResponseWriter, r *http.Request) {
	f.Request = r
	req := &brokerapi.BindingRequest{}
	if err := util.BodyToObject(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.RequestObject = req
	util.WriteResponse(w, f.responseStatus, &brokerapi.DeleteServiceInstanceResponse{})
}

func (f *FakeBrokerServer) unbindHandler(w http.ResponseWriter, r *http.Request) {
	f.Request = r
	util.WriteResponse(w, f.responseStatus, &brokerapi.DeleteServiceInstanceResponse{})
}
