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

package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	model "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

const (
	catalogFormatString         = "%s/v2/catalog"
	serviceInstanceFormatString = "%s/v2/service_instances/%s"
	bindingFormatString         = "%s/v2/service_instances/%s/service_bindings/%s"

	httpTimeoutSeconds = 15
)

type cfV2BrokerClient struct {
	broker *scmodel.ServiceBroker
	client *http.Client
}

// CreateCFV2BrokerClient creates an instance of BrokerClient for communicating
// with brokers which implement the Cloud Foundry Service Broker V2 API.
func CreateCFV2BrokerClient(b *scmodel.ServiceBroker) BrokerClient {
	return &cfV2BrokerClient{
		broker: b,
		client: &http.Client{
			Timeout: httpTimeoutSeconds * time.Second,
		},
	}
}

func (c *cfV2BrokerClient) GetCatalog() (*model.Catalog, error) {
	url := fmt.Sprintf(catalogFormatString, c.broker.BrokerURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.broker.AuthUsername, c.broker.AuthPassword)
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch catalog from %s\n%v", url, resp)
		log.Printf("err: %#v", err)
		return nil, err
	}

	var catalog model.Catalog
	if err = ResponseBodyToObject(resp, &catalog); err != nil {
		log.Printf("Failed to unmarshal catalog: %#v", err)
		return nil, err
	}

	return &catalog, nil
}

func (c *cfV2BrokerClient) CreateServiceInstance(ID string, req *model.ServiceInstanceRequest) (*model.ServiceInstance, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.BrokerURL, ID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	log.Printf("Doing a request to: %s", url)
	resp, err := c.client.Do(createHTTPReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Align this with the actual response model.
	si := model.ServiceInstance{}
	if err := ResponseBodyToObject(resp, &si); err != nil {
		return nil, err
	}
	return &si, nil
}

func (c *cfV2BrokerClient) UpdateServiceInstance(ID string, req *model.ServiceInstanceRequest) (*model.ServiceInstance, error) {
	// TODO: https://github.com/kubernetes-incubator/service-catalog/issues/114
	return nil, fmt.Errorf("Not implemented")
}

func (c *cfV2BrokerClient) DeleteServiceInstance(ID string) error {
	url := fmt.Sprintf(serviceInstanceFormatString, c.broker.BrokerURL, ID)

	// TODO: Handle the auth
	deleteHTTPReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("Failed to create new HTTP request: %v", err)
		return err
	}

	log.Printf("Doing a request to: %s", url)
	resp, err := c.client.Do(deleteHTTPReq)
	if err != nil {
		log.Printf("Failed to DELETE: %#v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *cfV2BrokerClient) CreateServiceBinding(sID, bID string, req *model.BindingRequest) (*model.CreateServiceBindingResponse, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf("Failed to marshal: %#v", err)
		return nil, err
	}

	url := fmt.Sprintf(bindingFormatString, c.broker.BrokerURL, sID, bID)

	// TODO: Handle the auth
	createHTTPReq, err := http.NewRequest("PUT", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	log.Printf("Doing a request to: %s", url)
	resp, err := c.client.Do(createHTTPReq)
	if err != nil {
		log.Printf("Failed to PUT: %#v", err)
		return nil, err
	}
	defer resp.Body.Close()

	sbr := model.CreateServiceBindingResponse{}
	err = ResponseBodyToObject(resp, &sbr)
	if err != nil {
		log.Printf("Failed to unmarshal: %#v", err)
		return nil, err
	}

	return &sbr, nil
}

func (c *cfV2BrokerClient) DeleteServiceBinding(sID, bID string) error {
	url := fmt.Sprintf(bindingFormatString, c.broker.BrokerURL, sID, bID)

	// TODO: Handle the auth
	deleteHTTPReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("Failed to create new HTTP request: %v", err)
		return err
	}

	log.Printf("Doing a request to: %s", url)
	resp, err := c.client.Do(deleteHTTPReq)
	if err != nil {
		log.Printf("Failed to DELETE: %#v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
