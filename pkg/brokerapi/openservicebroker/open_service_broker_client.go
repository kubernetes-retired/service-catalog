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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/util"

	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/openservicebroker/constants"
)

const (
	catalogFormatString               = "%s/v2/catalog"
	serviceInstanceFormatString       = "%s/v2/service_instances/%s"
	serviceInstanceDeleteFormatString = "%s/v2/service_instances/%s?service_id=%s&plan_id=%s"
	pollingFormatString               = "%s/v2/service_instances/%s/last_operation"
	bindingFormatString               = "%s/v2/service_instances/%s/service_bindings/%s"
	bindingDeleteFormatString         = "%s/v2/service_instances/%s/service_bindings/%s?service_id=%s&plan_id=%s"

	httpTimeoutSeconds     = 15
	pollingIntervalSeconds = 1
	pollingAmountLimit     = 30
)

var (
	errConflict        = errors.New("Service instance with same id but different attributes exists")
	errBindingConflict = errors.New("Service binding with same service instance id and binding id already exists")
	errBindingGone     = errors.New("There is no binding with the specified service instance id and binding id")
	errAsynchronous    = errors.New("Broker only supports this action asynchronously")
	errFailedState     = errors.New("Failed state received from broker")
	errUnknownState    = errors.New("Unknown state received from broker")
	errPollingTimeout  = errors.New("Timed out while polling broker")
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
	name     string
	url      string
	username string
	password string
	*http.Client
}

// NewClient creates an instance of BrokerClient for communicating with brokers
// which implement the Open Service Broker API.
func NewClient(name, url, username, password string) brokerapi.BrokerClient {
	// TODO(vaikas): Make this into a flag/config option. Necessary to talk to brokers that
	// have non-root signed certs.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &openServiceBrokerClient{
		name:     name,
		url:      strings.TrimRight(url, "/"), // remove trailing slashes from broker server URLs
		username: username,
		password: password,
		Client: &http.Client{
			Timeout:   httpTimeoutSeconds * time.Second,
			Transport: tr,
		},
	}
}

func (c *openServiceBrokerClient) GetCatalog() (*brokerapi.Catalog, error) {
	catalogURL := fmt.Sprintf(catalogFormatString, c.url)

	req, err := c.newOSBRequest(http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	resp, err := c.Do(req)
	if err != nil {
		glog.Errorf("Failed to fetch catalog %q from %s: response: %v error: %#v", c.name, catalogURL, resp, err)
		return nil, err
	}

	var catalog brokerapi.Catalog
	if err = util.ResponseBodyToObject(resp, &catalog); err != nil {
		glog.Errorf("Failed to unmarshal catalog from broker %q: %#v", c.name, err)
		return nil, err
	}

	return &catalog, nil
}

func (c *openServiceBrokerClient) CreateServiceInstance(ID string, req *brokerapi.CreateServiceInstanceRequest) (*brokerapi.CreateServiceInstanceResponse, error) {
	serviceInstanceURL := fmt.Sprintf(serviceInstanceFormatString, c.url, ID)
	// TODO: Handle the auth
	resp, err := sendOSBRequest(c, http.MethodPut, serviceInstanceURL, req)
	if err != nil {
		glog.Errorf("Error sending create service instance request to broker %q at %v: response: %v error: %#v", c.name, serviceInstanceURL, resp, err)
		return nil, errRequest{message: err.Error()}
	}
	defer resp.Body.Close()

	createServiceInstanceResponse := brokerapi.CreateServiceInstanceResponse{}
	if err := util.ResponseBodyToObject(resp, &createServiceInstanceResponse); err != nil {
		glog.Errorf("Error unmarshalling create service instance response from broker %q: %#v", c.name, err)
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
	serviceInstanceURL := fmt.Sprintf(serviceInstanceDeleteFormatString, c.url, ID, req.ServiceID, req.PlanID)
	// TODO: Handle the auth
	resp, err := sendOSBRequest(c, http.MethodDelete, serviceInstanceURL, req)
	if err != nil {
		glog.Errorf("Error sending delete service instance request to broker %q at %v: response: %v error: %#v", c.name, serviceInstanceURL, resp, err)
		return errRequest{message: err.Error()}
	}
	defer resp.Body.Close()

	deleteServiceInstanceResponse := brokerapi.DeleteServiceInstanceResponse{}
	if err := util.ResponseBodyToObject(resp, &deleteServiceInstanceResponse); err != nil {
		glog.Errorf("Error unmarshalling delete service instance response from broker %q: %#v", c.name, err)
		return errResponse{message: err.Error()}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusAccepted:
		glog.V(3).Infof("Asynchronous response received. Polling broker %q", c.name)
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

	serviceBindingURL := fmt.Sprintf(bindingFormatString, c.url, sID, bID)

	// TODO: Handle the auth
	createHTTPReq, err := c.newOSBRequest("PUT", serviceBindingURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	glog.Infof("Doing a request to: %s", serviceBindingURL)
	resp, err := c.Do(createHTTPReq)
	if err != nil {
		glog.Errorf("Failed to PUT: %#v", err)
		return nil, err
	}
	defer resp.Body.Close()

	createServiceBindingResponse := brokerapi.CreateServiceBindingResponse{}
	if err := util.ResponseBodyToObject(resp, &createServiceBindingResponse); err != nil {
		glog.Errorf("Error unmarshalling create binding response from broker: %#v", err)
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return &createServiceBindingResponse, nil
	case http.StatusOK:
		return &createServiceBindingResponse, nil
	case http.StatusConflict:
		return nil, errBindingConflict
	default:
		return nil, errStatusCode{statusCode: resp.StatusCode}
	}
}

func (c *openServiceBrokerClient) DeleteServiceBinding(sID, bID, serviceID, planID string) error {
	serviceBindingURL := fmt.Sprintf(bindingDeleteFormatString, c.url, sID, bID, serviceID, planID)

	// TODO: Handle the auth
	deleteHTTPReq, err := c.newOSBRequest("DELETE", serviceBindingURL, nil)
	if err != nil {
		glog.Errorf("Failed to create new HTTP request: %v", err)
		return err
	}

	glog.Infof("Doing a request to: %s", serviceBindingURL)
	resp, err := c.Do(deleteHTTPReq)
	if err != nil {
		glog.Errorf("Failed to DELETE: %#v", err)
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusGone:
		return errBindingGone
	default:
		return errStatusCode{statusCode: resp.StatusCode}
	}

}

func (c *openServiceBrokerClient) pollBroker(ID string, operation string) error {
	pollReq := brokerapi.LastOperationRequest{}
	if operation != "" {
		pollReq.Operation = operation
	}

	pollingURL := fmt.Sprintf(pollingFormatString, c.url, ID)
	for i := 0; i < pollingAmountLimit; i++ {
		glog.V(3).Infof("Polling broker %v at %s attempt %v", c.name, pollingURL, i+1)
		pollResp, err := sendOSBRequest(c, http.MethodGet, pollingURL, pollReq)
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

// SendRequest will serialize 'object' and send it using the given method to
// the given URL, through the provided client
func sendOSBRequest(c *openServiceBrokerClient, method string, url string, object interface{}) (*http.Response, error) {
	data, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %s", err.Error())
	}

	req, err := c.newOSBRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request object: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %s", err.Error())
	}

	return resp, nil
}

func (c *openServiceBrokerClient) newOSBRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add(constants.APIVersionHeader, constants.APIVersion)
	req.SetBasicAuth(c.username, c.password)
	return req, nil
}
