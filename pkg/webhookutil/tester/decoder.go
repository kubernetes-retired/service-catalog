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

package tester

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"testing"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//TestDecoderHandler represents a handler with a decoder
type TestDecoderHandler interface {
	InjectDecoder(d *admission.Decoder) error
	Handle(ctx context.Context, req admission.Request) admission.Response
}

// AssertHandlerReturnErrorIfReqObjIsMalformed checks error handling of malformed requests
func AssertHandlerReturnErrorIfReqObjIsMalformed(t *testing.T, handler TestDecoderHandler, kind string) {
	// given
	sc.AddToScheme(scheme.Scheme)
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	fixReq := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			Operation: admissionv1beta1.Create,
			Name:      "test-name",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    kind,
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte("{malformed: JSON,,")},
		},
	}

	expReqResult := &metav1.Status{
		Code:    http.StatusBadRequest,
		Message: "couldn't get version/kind; json parse error: invalid character 'm' looking for beginning of object key string",
	}

	handler.InjectDecoder(decoder)

	// when
	resp := handler.Handle(context.Background(), fixReq)

	// then
	assert.False(t, resp.Allowed)
	assert.Equal(t, expReqResult, resp.Result)
}

// AssertHandlerReturnErrorIfGVKMismatch checks error handling when wrong type of object is passed
func AssertHandlerReturnErrorIfGVKMismatch(t *testing.T, handler TestDecoderHandler, kind string) {
	// given
	sc.AddToScheme(scheme.Scheme)
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	fixReq := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			Operation: admissionv1beta1.Create,
			Name:      "test-name",
			Namespace: "system",
			Kind: metav1.GroupVersionKind{
				Kind:    "Incorrect" + kind,
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
		},
	}

	expReqResult := &metav1.Status{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("type mismatch: want: servicecatalog.k8s.io/v1beta1, Kind=%s got: servicecatalog.k8s.io/v1beta1, Kind=Incorrect%s", kind, kind),
	}

	handler.InjectDecoder(decoder)

	// when
	resp := handler.Handle(context.Background(), fixReq)

	// then
	assert.False(t, resp.Allowed)
	assert.Equal(t, expReqResult, resp.Result)
}

// DiscardLoggedMsg turns off log messages
func DiscardLoggedMsg() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Set("stderrthreshold", "999")
	flag.Parse()
	klog.SetOutput(ioutil.Discard)
}
