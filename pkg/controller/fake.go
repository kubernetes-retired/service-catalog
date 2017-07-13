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

package controller

import (
	"net/http/httptest"
	"time"

	fakebrokerserver "github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/fake/server"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
	v1alpha1informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1alpha1"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

// TestControllerWithBrokerServer contains information about a controller that uses mock clients
// except one that points to a broker server that runs in-memory
type TestControllerWithBrokerServer struct {
	FakeKubeClient      *clientgofake.Clientset
	FakeCatalogClient   *servicecatalogclientset.Clientset
	Controller          Controller
	Informers           v1alpha1informers.Interface
	BrokerServerHandler *fakebrokerserver.Handler
	BrokerServer        *httptest.Server
}

// Close releases all resources associated with t. It generally should be called in a defer after
// calling NewTestControllerWithBrokerServer
func (t *TestControllerWithBrokerServer) Close() {
	t.BrokerServer.Close()
}

// NewTestControllerWithBrokerServer creates a new TestControllerWithBrokerServer, or returns a
// non-nil error if there was a problem creating it. When a non-nil TestControllerWithBrokerServer
// is returned, it should be Close()-ed when the caller is done using it
func NewTestControllerWithBrokerServer(
	brokerUsername,
	brokerPassword string,
) (*TestControllerWithBrokerServer, error) {
	// create a fake kube client
	fakeKubeClient := &clientgofake.Clientset{}
	// create a fake sc client
	fakeCatalogClient := &servicecatalogclientset.Clientset{}

	brokerHandler := fakebrokerserver.NewHandler()
	brokerServer := fakebrokerserver.Run(brokerHandler, brokerUsername, brokerPassword)

	// create informers
	informerFactory := servicecataloginformers.NewSharedInformerFactory(fakeCatalogClient, 0)
	serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1alpha1()

	fakeRecorder := record.NewFakeRecorder(5)

	// create a test controller
	testController, err := NewController(
		fakeKubeClient,
		fakeCatalogClient.ServicecatalogV1alpha1(),
		serviceCatalogSharedInformers.Brokers(),
		serviceCatalogSharedInformers.ServiceClasses(),
		serviceCatalogSharedInformers.Instances(),
		serviceCatalogSharedInformers.Bindings(),
		osb.NewClient,
		24*time.Hour,
		true, /* enable OSB context profile */
		fakeRecorder,
	)
	if err != nil {
		return nil, err
	}

	return &TestControllerWithBrokerServer{
		FakeKubeClient:      fakeKubeClient,
		FakeCatalogClient:   fakeCatalogClient,
		Controller:          testController,
		Informers:           serviceCatalogSharedInformers,
		BrokerServerHandler: brokerHandler,
		BrokerServer:        brokerServer,
	}, nil
}
