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
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/util"
)

const (
	catalogFormatString         = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"

	httpTimeoutSeconds = 15
)

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

func (c *openServiceBrokerClient) CreateServiceInstance(ID string, req *brokerapi.ServiceInstanceRequest) (*brokerapi.ServiceInstance, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.Spec.URL, ID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	glog.Infof("Doing a request to: %s", url)
	resp, err := c.client.Do(createHTTPReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Align this with the actual response model.
	si := brokerapi.ServiceInstance{}
	if err := util.ResponseBodyToObject(resp, &si); err != nil {
		return nil, err
	}
	return &si, nil
}

func (c *openServiceBrokerClient) UpdateServiceInstance(ID string, req *brokerapi.ServiceInstanceRequest) (*brokerapi.ServiceInstance, error) {
	// TODO: https://github.com/kubernetes-incubator/service-catalog/issues/114
	return nil, fmt.Errorf("Not implemented")
}

func (c *openServiceBrokerClient) DeleteServiceInstance(ID string) error {
	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.Spec.URL, ID)

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
