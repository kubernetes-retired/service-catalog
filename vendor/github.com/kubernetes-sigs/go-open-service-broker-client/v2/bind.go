/*
Copyright 2019 The Kubernetes Authors.

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

package v2

import (
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

// internal message body types

type bindRequestBody struct {
	ServiceID    string                 `json:"service_id"`
	PlanID       string                 `json:"plan_id"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	BindResource map[string]interface{} `json:"bind_resource,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type bindSuccessResponseBody struct {
	Credentials     map[string]interface{} `json:"credentials"`
	SyslogDrainURL  *string                `json:"syslog_drain_url"`
	RouteServiceURL *string                `json:"route_service_url"`
	VolumeMounts    []interface{}          `json:"volume_mounts"`
	Operation       *string                `json:"operation"`
}

const (
	bindResourceAppGUIDKey = "app_guid"
	bindResourceRouteKey   = "route"
)

func (c *client) Bind(r *BindRequest) (*BindResponse, error) {
	if r.AcceptsIncomplete {
		if err := c.validateClientVersionIsAtLeast(Version2_14()); err != nil {
			return nil, AsyncBindingOperationsNotAllowedError{
				reason: err.Error(),
			}
		}
	}

	if err := validateBindRequest(r); err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf(bindingURLFmt, c.URL, r.InstanceID, r.BindingID)

	params := map[string]string{}
	if r.AcceptsIncomplete {
		params[AcceptsIncomplete] = "true"
	}

	requestBody := &bindRequestBody{
		ServiceID:  r.ServiceID,
		PlanID:     r.PlanID,
		Parameters: r.Parameters,
	}

	if c.APIVersion.AtLeast(Version2_13()) {
		requestBody.Context = r.Context
	}

	if r.BindResource != nil {
		requestBody.BindResource = map[string]interface{}{}
		if r.BindResource.AppGUID != nil {
			requestBody.BindResource[bindResourceAppGUIDKey] = *r.BindResource.AppGUID
		}
		if r.BindResource.Route != nil {
			requestBody.BindResource[bindResourceRouteKey] = *r.BindResource.AppGUID
		}
	}

	response, err := c.prepareAndDo(http.MethodPut, fullURL, params, requestBody, r.OriginatingIdentity)
	if err != nil {
		return nil, err
	}

	defer func() {
		drainReader(response.Body)
		response.Body.Close()
	}()

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		userResponse := &BindResponse{}
		if err := c.unmarshalResponse(response, userResponse); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		return userResponse, nil
	case http.StatusAccepted:
		if !r.AcceptsIncomplete {
			return nil, c.handleFailureResponse(response)
		}

		responseBodyObj := &bindSuccessResponseBody{}
		if err := c.unmarshalResponse(response, responseBodyObj); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		var opPtr *OperationKey
		if responseBodyObj.Operation != nil {
			opStr := *responseBodyObj.Operation
			op := OperationKey(opStr)
			opPtr = &op
		}

		userResponse := &BindResponse{
			Credentials:     responseBodyObj.Credentials,
			SyslogDrainURL:  responseBodyObj.SyslogDrainURL,
			RouteServiceURL: responseBodyObj.RouteServiceURL,
			VolumeMounts:    responseBodyObj.VolumeMounts,
			OperationKey:    opPtr,
		}
		if response.StatusCode == http.StatusAccepted {
			if c.Verbose {
				klog.Infof("broker %q: received asynchronous response", c.Name)
			}
			userResponse.Async = true
		}

		return userResponse, nil
	default:
		return nil, c.handleFailureResponse(response)
	}
}

func validateBindRequest(request *BindRequest) error {
	if request.BindingID == "" {
		return required("bindingID")
	}

	if request.InstanceID == "" {
		return required("instanceID")
	}

	if request.ServiceID == "" {
		return required("serviceID")
	}

	if request.PlanID == "" {
		return required("planID")
	}

	return nil
}
