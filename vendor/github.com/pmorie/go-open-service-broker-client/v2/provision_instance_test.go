package v2

import (
	"fmt"
	"net/http"
	"testing"
)

const (
	testInstanceID       = "test-instance-id"
	testServiceID        = "test-service-id"
	testPlanID           = "test-plan-id"
	testOrganizationGUID = "test-organization-guid"
	testSpaceGUID        = "test-space-guid"
)

func defaultProvisionRequest() *ProvisionRequest {
	return &ProvisionRequest{
		InstanceID:       testInstanceID,
		ServiceID:        testServiceID,
		PlanID:           testPlanID,
		OrganizationGUID: testOrganizationGUID,
		SpaceGUID:        testSpaceGUID,
	}
}

func defaultAsyncProvisionRequest() *ProvisionRequest {
	r := defaultProvisionRequest()
	r.AcceptsIncomplete = true
	return r
}

const successProvisionRequestBody = `{"service_id":"test-service-id","plan_id":"test-plan-id","organization_guid":"test-organization-guid","space_guid":"test-space-guid"}`

const successProvisionResponseBody = `{
  "dashboard_url": "https://example.com/dashboard"
}`

var testDashboardURL = "https://example.com/dashboard"
var testOperation OperationKey = "test-operation-key"

func successProvisionResponse() *ProvisionResponse {
	return &ProvisionResponse{
		DashboardURL: &testDashboardURL,
	}
}

const successAsyncProvisionResponseBody = `{
  "dashboard_url": "https://example.com/dashboard",
  "operation": "test-operation-key"
}`

func successProvisionResponseAsync() *ProvisionResponse {
	r := successProvisionResponse()
	r.Async = true
	r.OperationKey = &testOperation
	return r
}

const alphaContextProvisionRequestBody = `{"service_id":"test-service-id","plan_id":"test-plan-id","organization_guid":"test-organization-guid","space_guid":"test-space-guid","context":{"foo":"bar"}}`

func TestProvisionInstance(t *testing.T) {
	cases := []struct {
		name               string
		enableAlpha        bool
		request            *ProvisionRequest
		httpChecks         httpChecks
		httpReaction       httpReaction
		expectedResponse   *ProvisionResponse
		expectedErrMessage string
		expectedErr        error
	}{
		{
			name: "invalid request",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.InstanceID = ""
				return r
			}(),
			expectedErrMessage: "instanceID is required",
		},
		{
			name: "success - created",
			httpReaction: httpReaction{
				status: http.StatusCreated,
				body:   successProvisionResponseBody,
			},
			expectedResponse: successProvisionResponse(),
		},
		{
			name: "success - ok",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successProvisionResponseBody,
			},
			expectedResponse: successProvisionResponse(),
		},
		{
			name:    "success - asynchronous",
			request: defaultAsyncProvisionRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					asyncQueryParamKey: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncProvisionResponseBody,
			},
			expectedResponse: successProvisionResponseAsync(),
		},
		{
			name: "http error",
			httpReaction: httpReaction{
				err: fmt.Errorf("http error"),
			},
			expectedErrMessage: "http error",
		},
		{
			name: "200 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   malformedResponse,
			},
			expectedErrMessage: "unexpected end of JSON input",
		},
		{
			name: "500 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   malformedResponse,
			},
			expectedErrMessage: "unexpected end of JSON input",
		},
		{
			name: "500 with conventional failure response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   conventionalFailureResponseBody,
			},
			expectedErr: testHttpStatusCodeError(),
		},
		{
			name:        "alpha - context",
			enableAlpha: true,
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.AlphaContext = map[string]interface{}{
					"foo": "bar",
				}
				return r
			}(),
			httpChecks: httpChecks{
				body: alphaContextProvisionRequestBody,
			},
			httpReaction: httpReaction{
				status: http.StatusCreated,
				body:   successProvisionResponseBody,
			},
			expectedResponse: successProvisionResponse(),
		},
	}

	for _, tc := range cases {
		if tc.request == nil {
			tc.request = defaultProvisionRequest()
		}

		if tc.httpChecks.URL == "" {
			tc.httpChecks.URL = "/v2/service_instances/test-instance-id"
		}

		if tc.httpChecks.body == "" {
			tc.httpChecks.body = successProvisionRequestBody
		}

		klient := newTestClient(t, tc.name, tc.enableAlpha, tc.httpChecks, tc.httpReaction)

		response, err := klient.ProvisionInstance(tc.request)

		doResponseChecks(t, tc.name, response, err, tc.expectedResponse, tc.expectedErrMessage, tc.expectedErr)
	}
}

func TestValidateProvisionRequest(t *testing.T) {
	cases := []struct {
		name    string
		request *ProvisionRequest
		valid   bool
	}{
		{
			name:    "valid",
			request: defaultProvisionRequest(),
			valid:   true,
		},
		{
			name: "missing instance ID",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.InstanceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing service ID",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.ServiceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing plan ID",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.PlanID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing organization GUID",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.OrganizationGUID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing space GUID",
			request: func() *ProvisionRequest {
				r := defaultProvisionRequest()
				r.SpaceGUID = ""
				return r
			}(),
			valid: false,
		},
	}

	for _, tc := range cases {
		err := validateProvisionRequest(tc.request)
		if err != nil {
			if tc.valid {
				t.Errorf("%v: expected valid, got error: %v", tc.name, err)
			}
		} else if !tc.valid {
			t.Errorf("%v: expected invalid, got valid", tc.name)
		}
	}
}
