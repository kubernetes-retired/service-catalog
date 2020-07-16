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

type unbindSuccessResponseBody struct {
	Operation *string `json:"operation"`
}

func (c *client) Unbind(r *UnbindRequest) (*UnbindResponse, error) {
	if r.AcceptsIncomplete {
		if err := c.validateClientVersionIsAtLeast(Version2_14()); err != nil {
			return nil, AsyncBindingOperationsNotAllowedError{
				reason: err.Error(),
			}
		}
	}

	if err := validateUnbindRequest(r); err != nil {
		return nil, err
	}

	fullURL := fmt.Sprintf(bindingURLFmt, c.URL, r.InstanceID, r.BindingID)
	params := map[string]string{}
	params[VarKeyServiceID] = r.ServiceID
	params[VarKeyPlanID] = r.PlanID
	if r.AcceptsIncomplete {
		params[AcceptsIncomplete] = "true"
	}

	response, err := c.prepareAndDo(http.MethodDelete, fullURL, params, nil, r.OriginatingIdentity)
	if err != nil {
		return nil, err
	}

	defer func() {
		drainReader(response.Body)
		response.Body.Close()
	}()

	switch response.StatusCode {
	case http.StatusOK, http.StatusGone:
		userResponse := &UnbindResponse{}
		if err := c.unmarshalResponse(response, userResponse); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		return userResponse, nil
	case http.StatusAccepted:
		if !r.AcceptsIncomplete {
			return nil, c.handleFailureResponse(response)
		}

		responseBodyObj := &unbindSuccessResponseBody{}
		if err := c.unmarshalResponse(response, responseBodyObj); err != nil {
			return nil, HTTPStatusCodeError{StatusCode: response.StatusCode, ResponseError: err}
		}

		var opPtr *OperationKey
		if responseBodyObj.Operation != nil {
			opStr := *responseBodyObj.Operation
			op := OperationKey(opStr)
			opPtr = &op
		}

		userResponse := &UnbindResponse{
			OperationKey: opPtr,
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

func validateUnbindRequest(request *UnbindRequest) error {
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
