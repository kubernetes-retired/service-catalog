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

package openservicebroker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/util"
)

const (
	catalogFormatString         = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	pollingFormatString         = "%s/v2/service_instances/%s/last_operation"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"

	httpTimeoutSeconds     = 15
	pollingIntervalSeconds = 1
	pollingAmountLimit     = 30
)

var (
	errConflict       = errors.New("Service instance with same id but different attributes exists")
	errAsynchronous   = errors.New("Broker only supports this action asynchronously")
	errFailedState    = errors.New("Failed state received from broker")
	errUnknownState   = errors.New("Unknown state received from broker")
	errPollingTimeout = errors.New("Timed out while polling broker")
)

type (
	errRequest struct {
		message string
	}

	errResponse struct {
		message string
	}

	errStatusCode struct {
		statusCode int
	}
)

func (e errRequest) Error() string {
	return fmt.Sprintf("Failed to send request: %s", e.message)
}

func (e errResponse) Error() string {
	return fmt.Sprintf("Failed to parse broker response: %s", e.message)
}

func (e errStatusCode) Error() string {
	return fmt.Sprintf("Unexpected status code from broker response: %v", e.statusCode)
}

type openServiceBrokerClient struct {
	broker *servicecatalog.Broker
	client *http.Client
}

// NewClient creates an instance of BrokerClient for communicating with brokers
// which implement the Open Service Broker API.
func NewClient(b *servicecatalog.Broker) brokerapi.BrokerClient {
	return &openServiceBrokerClient{
		broker: b,
		client: &http.Client{
			Timeout: httpTimeoutSeconds * time.Second,
		},
	}
}

func (c *openServiceBrokerClient) GetCatalog() (*brokerapi.Catalog, error) {
	url := fmt.Sprintf(catalogFormatString, c.broker.Spec.URL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.broker.Spec.AuthUsername, c.broker.Spec.AuthPassword)
	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to fetch catalog from %s\n%v", url, resp)
		glog.Errorf("err: %#v", err)
		return nil, err
	}

	var catalog brokerapi.Catalog
	if err = util.ResponseBodyToObject(resp, &catalog); err != nil {
		glog.Errorf("Failed to unmarshal catalog: %#v", err)
		return nil, err
	}

	return &catalog, nil
}

func (c *openServiceBrokerClient) CreateServiceInstance(ID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.Spec.URL, ID)
	// TODO: Handle the auth
	resp, err := util.SendRequest(c.client, http.MethodPut, url, req)
	if err != nil {
		return nil, errRequest{message: err.Error()}
	}
	defer resp.Body.Close()

	createServiceInstanceResponse := brokerapi.CreateServiceInstanceResponse{}
	if err := util.ResponseBodyToObject(resp, &createServiceInstanceResponse); err != nil {
		return nil, errResponse{message: err.Error()}
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return &createServiceInstanceResponse, nil
	case http.StatusOK:
		return &createServiceInstanceResponse, nil
	case http.StatusAccepted:
		glog.V(3).Infof("Asynchronous response received. Polling broker.")
		if err := c.pollBroker(ID, createServiceInstanceResponse.Operation); err != nil {
			return nil, err
		}

		return &createServiceInstanceResponse, nil
	case http.StatusConflict:
		return nil, errConflict
	case http.StatusUnprocessableEntity:
		return nil, errAsynchronous
	default:
		return nil, errStatusCode{statusCode: resp.StatusCode}
	}
}

func (c *openServiceBrokerClient) UpdateServiceInstance(ID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.ServiceInstance, error) {
	// TODO: https://github.com/kubernetes-incubator/service-catalog/issues/114
	return nil, fmt.Errorf("Not implemented")
}

func (c *openServiceBrokerClient) DeleteServiceInstance(ID string, req *brokerapi.DeleteServiceInstanceRequest) error {
	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.Spec.URL, ID)
	// TODO: Handle the auth
	resp, err := util.SendRequest(c.client, http.MethodDelete, url, req)
	if err != nil {
		return errRequest{message: err.Error()}
	}
	defer resp.Body.Close()

	deleteServiceInstanceResponse := brokerapi.DeleteServiceInstanceResponse{}
	if err := util.ResponseBodyToObject(resp, &deleteServiceInstanceResponse); err != nil {
		return errResponse{message: err.Error()}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusAccepted:
		glog.V(3).Infof("Asynchronous response received. Polling broker.")
		if err := c.pollBroker(ID, deleteServiceInstanceResponse.Operation); err != nil {
			return err
		}

		return nil
	case http.StatusGone:
		return nil
	case http.StatusUnprocessableEntity:
		return errAsynchronous
	default:
		return errStatusCode{statusCode: resp.StatusCode}
	}
}

func (c *openServiceBrokerClient) CreateServiceBinding(sID, bID string, req *brokerapi.BindingRequest) (*brokerapi.CreateServiceBindingResponse, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		glog.Errorf("Failed to marshal: %#v", err)
		return nil, err
	}

	url := fmt.Sprintf(bindingFormatString, c.broker.Spec.URL, sID, bID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	glog.Infof("Doing a request to: %s", url)
	resp, err := c.client.Do(createHTTPReq)
	if err != nil {
		glog.Errorf("Failed to PUT: %#v", err)
		return nil, err
	}
	defer resp.Body.Close()

	sbr := brokerapi.CreateServiceBindingResponse{}
	err = util.ResponseBodyToObject(resp, &sbr)
	if err != nil {
		glog.Errorf("Failed to unmarshal: %#v", err)
		return nil, err
	}

	return &sbr, nil
}

func (c *openServiceBrokerClient) DeleteServiceBinding(sID, bID string) error {
	url := fmt.Sprintf(bindingFormatString, c.broker.Spec.URL, sID, bID)

	// TODO: Handle the auth
	deleteHTTPReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		glog.Errorf("Failed to create new HTTP request: %v", err)
		return err
	}

	glog.Infof("Doing a request to: %s", url)
	resp, err := c.client.Do(deleteHTTPReq)
	if err != nil {
		glog.Errorf("Failed to DELETE: %#v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *openServiceBrokerClient) pollBroker(ID string, operation string) error {
	pollReq := brokerapi.LastOperationRequest{}
	if operation != "" {
		pollReq.Operation = operation
	}

	url := fmt.Sprintf(pollingFormatString, c.broker.Spec.URL, ID)
	for i := 0; i < pollingAmountLimit; i++ {
		glog.V(3).Infof("Polling attempt #%v: %s", i+1, url)
		pollResp, err := util.SendRequest(c.client, http.MethodGet, url, pollReq)
		if err != nil {
			return err
		}
		defer pollResp.Body.Close()

		lo := brokerapi.LastOperationResponse{}
		if err := util.ResponseBodyToObject(pollResp, &lo); err != nil {
			return err
		}

		switch lo.State {
		case brokerapi.StateInProgress:
		case brokerapi.StateSucceeded:
			return nil
		case brokerapi.StateFailed:
			return errFailedState
		default:
			return errUnknownState
		}

		time.Sleep(pollingIntervalSeconds * time.Second)
	}

	return errPollingTimeout
}
