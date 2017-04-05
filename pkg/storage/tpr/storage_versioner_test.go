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

package tpr

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/api"
	_ "k8s.io/client-go/rest"
	fakerestclient "k8s.io/client-go/rest/fake"
)

func TestUpdateNonExistentObject(t *testing.T) {
	v := getTestVersioner(
		t,
		func(*http.Request) (*http.Response, error) {
			return getMockUpdateResponse(http.StatusNotFound), nil
		},
	)
	// For convenience, we're passing 0 as the resource version, because we know
	// that, internally, the function under test ONLY passes it to the core
	// apiserver-- which is mocked out for the purposes of this test and will do
	// nothing with it.
	if err := v.UpdateObject(getTestBroker(), 0); err == nil {
		t.Fatal("did not receive expected error")
	}
}

func TestSuccessfulUpdateObject(t *testing.T) {
	brokerObj := getTestBroker()
	brokerCodec := getBrokerCodec(t)
	versionedBrokerBytes, err := runtime.Encode(brokerCodec, brokerObj)
	if err != nil {
		t.Fatalf(
			"error encoding broker object to versioned broker "+
				"bytes: %s", err,
		)
	}
	versionedBrokerObj, err := runtime.Decode(brokerCodec, versionedBrokerBytes)
	if err != nil {
		t.Fatalf(
			"error decoding versioned broker bytes to versioned broker "+
				"object: %s", err,
		)
	}
	v := getTestVersioner(
		t,
		func(request *http.Request) (*http.Response, error) {
			// Assert that what was sent to the mock apiserver was exactly the
			// bytes for the versioned equivalent of the input to UpdateObject()
			bodyBytes, err := ioutil.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("error getting request body bytes: %s", err)
			}
			decodedBrokerObj, err := runtime.Decode(
				getBrokerCodec(t),
				bodyBytes,
			)
			if err != nil {
				t.Fatalf("error decoding request body: %s", err)
			}
			if !reflect.DeepEqual(versionedBrokerObj, decodedBrokerObj) {
				t.Fatal(
					"request did not include expected serialized, versioned " +
						"broker",
				)
			}
			return getMockUpdateResponse(http.StatusOK), nil
		},
	)
	// For convenience, we're passing 0 as the resource version, because we know
	// that, internally, the function under test ONLY passes it to the core
	// apiserver-- which is mocked out for the purposes of this test and will do
	// nothing with it.
	if err := v.UpdateObject(brokerObj, 0); err != nil {
		t.Fatalf("error updating list: %s", err)
	}
}

func TestUpdateListWithNonExistentObject(t *testing.T) {
	v := getTestVersioner(
		t,
		func(*http.Request) (*http.Response, error) {
			return getMockUpdateResponse(http.StatusNotFound), nil
		},
	)
	// For convenience, we're passing 0 as the resource version, because we know
	// that, internally, the function under test completely ignores it anyway.
	if err := v.UpdateList(getTestBrokerList(), 0); err == nil {
		t.Fatal("did not receive expected error")
	}
}

func TestSuccessfulUpdateList(t *testing.T) {
	brokerList := getTestBrokerList()
	var callCount int
	v := getTestVersioner(
		t,
		func(*http.Request) (*http.Response, error) {
			callCount++
			return getMockUpdateResponse(http.StatusOK), nil
		},
	)
	// For convenience, we're passing 0 as the resource version, because we know
	// that, internally, the function under test completely ignores it anyway.
	if err := v.UpdateList(brokerList, 0); err != nil {
		t.Fatalf("error updating list: %s", err)
	}
	// Assert the correct number of calls were made to out mocked out apiserver
	if callCount != len(brokerList.Items) {
		t.Fatalf(
			"incorrect number of object updates sent to the apiserver; "+
				"expected %d, got %d", len(brokerList.Items), callCount,
		)
	}
}

func getTestBroker() *servicecatalog.Broker {
	return &servicecatalog.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: "servicecatalog.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:          map[string]string{"name": "broker_foo"},
			ResourceVersion: "0",
		},
	}
}

func getTestBrokerList() *servicecatalog.BrokerList {
	return &servicecatalog.BrokerList{
		Items: []servicecatalog.Broker{
			*getTestBroker(),
			*getTestBroker(),
		},
	}
}

func getMockUpdateResponse(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewBufferString("")),
	}
}

func getTestVersioner(
	t *testing.T,
	roundTripper func(*http.Request) (*http.Response, error),
) *versioner {
	return &versioner{
		codec:        getBrokerCodec(t),
		singularKind: ServiceBrokerKind,
		listKind:     ServiceBrokerListKind,
		restClient: &fakerestclient.RESTClient{
			APIRegistry: api.Registry,
			Client:      fakerestclient.CreateHTTPClient(roundTripper),
			NegotiatedSerializer: serializer.DirectCodecFactory{
				CodecFactory: api.Codecs,
			},
		},
		defaultNS: "test-ns",
	}
}

func getBrokerCodec(t *testing.T) runtime.Codec {
	codec, err := testapi.GetCodecForObject(getTestBroker())
	if err != nil {
		t.Fatalf("error getting codec: %s", err)
	}
	return codec
}
