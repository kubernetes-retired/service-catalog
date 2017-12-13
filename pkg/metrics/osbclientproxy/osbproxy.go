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

// Package osbclientproxy proxies the OSB Client Library enabling
// metrics instrumentation
package osbclientproxy

import (
	"strconv"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/metrics"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

// proxyclient provides a functional implementation of the OSB V2 Client
// interface
type proxyclient struct {
	brokerName    string
	realOSBClient osb.Client
}

// NewClient is a CreateFunc for creating a new functional Client and
// implements the CreateFunc interface.
func NewClient(config *osb.ClientConfiguration) (osb.Client, error) {
	osbClient, err := osb.NewClient(config)
	if err != nil {
		return nil, err
	}
	proxy := proxyclient{realOSBClient: osbClient}
	proxy.brokerName = config.Name
	return proxy, nil
}

var _ osb.CreateFunc = NewClient

// GetCatalog implements go-open-service-broker-client/v2/Client.GetCatalog by
// proxying the method to the underlying implementation and capturing request
// metrics.
func (pc proxyclient) GetCatalog() (*osb.CatalogResponse, error) {
	glog.V(9).Info("OSBClientProxy getCatalog()")
	response, err := pc.realOSBClient.GetCatalog()
	pc.updateMetrics(err)
	return response, err
}

// ProvisionInstance implements
// go-open-service-broker-client/v2/Client.ProvisionInstance by proxying the
// method to the underlying implementation and capturing request metrics.
func (pc proxyclient) ProvisionInstance(r *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
	glog.V(9).Info("OSBClientProxy ProvisionInstance()")
	response, err := pc.realOSBClient.ProvisionInstance(r)
	pc.updateMetrics(err)
	return response, err

}

// UpdateInstance implements
// go-open-service-broker-client/v2/Client.UpdateInstance by proxying the method
// to the underlying implementation and capturing request metrics.
func (pc proxyclient) UpdateInstance(r *osb.UpdateInstanceRequest) (*osb.UpdateInstanceResponse, error) {
	glog.V(9).Info("OSBClientProxy UpdateInstance()")
	response, err := pc.realOSBClient.UpdateInstance(r)
	pc.updateMetrics(err)
	return response, err
}

// DeprovisionInstance implements
// go-open-service-broker-client/v2/Client.DeprovisionInstance by proxying the
// method to the underlying implementation and capturing request metrics.
func (pc proxyclient) DeprovisionInstance(r *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
	glog.V(9).Info("OSBClientProxy DeprovisionInstance()")
	response, err := pc.realOSBClient.DeprovisionInstance(r)
	pc.updateMetrics(err)
	return response, err
}

// PollLastOperation implements
// go-open-service-broker-client/v2/Client.PollLastOperation by proxying the
// method to the underlying implementation and capturing request metrics.
func (pc proxyclient) PollLastOperation(r *osb.LastOperationRequest) (*osb.LastOperationResponse, error) {
	glog.V(9).Info("OSBClientProxy PollLastOperation()")
	response, err := pc.realOSBClient.PollLastOperation(r)
	pc.updateMetrics(err)
	return response, err
}

// PollBindingLastOperation implements
// go-open-service-broker-client/v2/Client.PollBindingLastOperation by proxying
// the method to the underlying implementation and capturing request metrics.
func (pc proxyclient) PollBindingLastOperation(r *osb.BindingLastOperationRequest) (*osb.LastOperationResponse, error) {
	glog.V(9).Info("OSBClientProxy PollBindingLastOperation()")
	response, err := pc.realOSBClient.PollBindingLastOperation(r)
	pc.updateMetrics(err)
	return response, err
}

// Bind implements go-open-service-broker-client/v2/Client.Bind by proxying the
// method to the underlying implementation and capturing request metrics.
func (pc proxyclient) Bind(r *osb.BindRequest) (*osb.BindResponse, error) {
	glog.V(9).Info("OSBClientProxy Bind().")
	response, err := pc.realOSBClient.Bind(r)
	pc.updateMetrics(err)
	return response, err
}

// Unbind implements go-open-service-broker-client/v2/Client.Unbind by proxying
// the method to the underlying implementation and capturing request metrics.
func (pc proxyclient) Unbind(r *osb.UnbindRequest) (*osb.UnbindResponse, error) {
	glog.V(9).Info("OSBClientProxy Unbind()")
	response, err := pc.realOSBClient.Unbind(r)
	pc.updateMetrics(err)
	return response, err
}

// GetBinding implements go-open-service-broker-client/v2/Client.GetBinding by
// proxying the method to the underlying implementation and capturing request
// metrics.
func (pc proxyclient) GetBinding(r *osb.GetBindingRequest) (*osb.GetBindingResponse, error) {
	glog.V(9).Info("OSBClientProxy GetBinding()")
	response, err := pc.realOSBClient.GetBinding(r)
	pc.updateMetrics(err)
	return response, err
}

// updateMetrics bumps the request count metric for the specific broker and
// status
func (pc proxyclient) updateMetrics(err error) {
	if err == nil {
		metrics.OSBRequestCount.WithLabelValues(pc.brokerName, "200").Inc()
	} else {
		status, ok := osb.IsHTTPError(err)
		if ok {
			metrics.OSBRequestCount.WithLabelValues(pc.brokerName, strconv.Itoa(status.StatusCode/100*100)).Inc()
		} else {
			metrics.OSBRequestCount.WithLabelValues(pc.brokerName, "client-error").Inc()
		}
	}
}
