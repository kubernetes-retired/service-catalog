package v2

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsHTTPError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-http error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "http error",
			err:      HTTPStatusCodeError{},
			expected: true,
		},
		{
			name:     "nil",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range cases {
		if e, a := tc.expected, IsHTTPError(tc.err); e != a {
			t.Errorf("%v: expected %v, got %v", tc.name, e, a)
		}
	}
}

func TestIsConflictError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-http error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "http non-conflict error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name: "http conflict error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusConflict,
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		if e, a := tc.expected, IsConflictError(tc.err); e != a {
			t.Errorf("%v: expected %v, got %v", tc.name, e, a)
		}
	}
}

func TestIsGoneError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-http error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "http non-gone error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name: "http gone error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusGone,
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		if e, a := tc.expected, IsGoneError(tc.err); e != a {
			t.Errorf("%v: expected %v, got %v", tc.name, e, a)
		}
	}
}

func strPtr(s string) *string {
	return &s
}

func TestIsAsyncRequiredError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-http error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "other http error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name: "async required error",
			err: HTTPStatusCodeError{
				StatusCode:   http.StatusUnprocessableEntity,
				ErrorMessage: strPtr(asyncErrorMessage),
				Description:  strPtr(asyncErrorDescription),
			},
			expected: true,
		},
		{
			name: "app guid required error",
			err: HTTPStatusCodeError{
				StatusCode:   http.StatusUnprocessableEntity,
				ErrorMessage: strPtr(appGUIDRequiredErrorMessage),
				Description:  strPtr(appGUIDRequiredErrorDescription),
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		if e, a := tc.expected, IsAsyncRequiredError(tc.err); e != a {
			t.Errorf("%v: expected %v, got %v", tc.name, e, a)
		}
	}
}

func TestIsAppGUIDRequiredError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "non-http error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "other http error",
			err: HTTPStatusCodeError{
				StatusCode: http.StatusForbidden,
			},
			expected: false,
		},
		{
			name: "async required error",
			err: HTTPStatusCodeError{
				StatusCode:   http.StatusUnprocessableEntity,
				ErrorMessage: strPtr(asyncErrorMessage),
				Description:  strPtr(asyncErrorDescription),
			},
			expected: false,
		},
		{
			name: "app guid required error",
			err: HTTPStatusCodeError{
				StatusCode:   http.StatusUnprocessableEntity,
				ErrorMessage: strPtr(appGUIDRequiredErrorMessage),
				Description:  strPtr(appGUIDRequiredErrorDescription),
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		if e, a := tc.expected, IsAppGUIDRequiredError(tc.err); e != a {
			t.Errorf("%v: expected %v, got %v", tc.name, e, a)
		}
	}
}
