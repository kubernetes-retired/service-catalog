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

package integration

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/golang/glog"

	// metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"

	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"

	// TODO: fix this upstream
	// we shouldn't have to install things to use our own generated client.

	// avoid error `servicecatalog/v1alpha1 is not enabled`
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	// avoid error `no kind is registered for the type v1.ListOptions`
	_ "k8s.io/kubernetes/pkg/api/install"

	// to start our server locally
	"github.com/kubernetes-incubator/service-catalog/cmd/apiserver/app/server"
	// our versioned types
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	// our versioned client
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	servicecatalogclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

// TestGroupVersion is trivial.
func TestGroupVersion(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()

	gv := client.Servicecatalog().RESTClient().APIVersion()
	if gv.Group != servicecatalog.GroupName {
		t.Fatal("we should be testing the servicecatalog group, not ", gv.Group)
	}
}

// TestNoName checks that all creates fail for objects that have no
// name given.
func TestNoName(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	scClient := client.Servicecatalog()

	ns := "namespace"

	if br, e := scClient.Brokers().Create(&v1alpha1.Broker{}); nil == e {
		t.Fatal("needs a name", br.Name)
	}
	if sc, e := scClient.ServiceClasses(ns).Create(&v1alpha1.ServiceClass{}); nil == e {
		t.Fatal("needs a name", sc.Name)
	}
	if i, e := scClient.Instances(ns).Create(&v1alpha1.Instance{}); nil == e {
		t.Fatal("needs a name", i.Name)
	}
	if bi, e := scClient.Bindings(ns).Create(&v1alpha1.Binding{}); nil == e {
		t.Fatal("needs a name", bi.Name)
	}
}

func TestBroker(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	brokerClient := client.Servicecatalog().Brokers()

	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: "test-broker"},
		Spec: v1alpha1.BrokerSpec{
			URL:          "https://example.com",
			AuthUsername: "auth username field value",
			AuthPassword: "auth password field value",
			OSBGUID:      "OSBGUID field",
		},
	}

	// start from scratch
	brokers, err := brokerClient.List(v1.ListOptions{})
	if len(brokers.Items) > 0 {
		t.Fatalf("brokers should not exist on start, had %v brokers", len(brokers.Items))
	}

	brokerServer, err := brokerClient.Create(broker)
	if nil != err {
		t.Fatal("error creating the broker", broker)
	}
	if broker.Name != brokerServer.Name {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}

	brokers, err = brokerClient.List(v1.ListOptions{})
	if 1 != len(brokers.Items) {
		t.Fatalf("should have exactly one broker, had %v brokers", len(brokers.Items))
	}

	brokerServer, err = brokerClient.Get(broker.Name)
	if broker.Name != brokerServer.Name &&
		broker.ResourceVersion == brokerServer.ResourceVersion {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}

	// check that the broker is the same both ways
	brokerListed := &brokers.Items[0]
	if !reflect.DeepEqual(brokerServer, brokerListed) {
		t.Fatalf("didn't get the same broker twice", brokerServer, brokerListed)
	}

	brokerServer.Spec.AuthUsername = "dug"
	brokerServer.Spec.AuthPassword = "paul"
	brokerUpdated, err := brokerClient.Update(brokerServer)
	if nil != err ||
		"dug" != brokerUpdated.Spec.AuthUsername ||
		"paul" != brokerUpdated.Spec.AuthPassword {
		t.Fatal("broker wasn't updated", brokerServer, brokerUpdated)
	}

	brokerServer, err = brokerClient.Get("test-broker")
	if nil != err ||
		"dug" != brokerServer.Spec.AuthUsername ||
		"paul" != brokerServer.Spec.AuthPassword {
		t.Fatal("broker wasn't updated", brokerServer)
	}

	err = brokerClient.Delete("test-broker", &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("broker should be deleted", err)
	}

	brokerDeleted, err := brokerClient.Get("test-broker")
	if nil == err {
		t.Fatal("broker should be deleted", brokerDeleted)
	}
}

func getFreshApiserverAndClient(t *testing.T) (servicecatalogclient.Interface, func()) {
	securePort := 65535
	serverIP := fmt.Sprintf("https://localhost:%d", securePort)
	stopCh := make(chan struct{})
	serverFailed := make(chan struct{})
	//defer close(stopCh)
	shutdown := func() {
		close(stopCh)
	}

	secureServingOptions := genericserveroptions.NewSecureServingOptions()
	go func() {
		options := &server.ServiceCatalogServerOptions{
			GenericServerRunOptions: genericserveroptions.NewServerRunOptions(),
			SecureServingOptions:    secureServingOptions,
			EtcdOptions:             genericserveroptions.NewEtcdOptions(),
			AuthenticationOptions:   genericserveroptions.NewDelegatingAuthenticationOptions(),
			AuthorizationOptions:    genericserveroptions.NewDelegatingAuthorizationOptions(),
		}
		options.SecureServingOptions.ServingOptions.BindPort = securePort
		options.EtcdOptions.StorageConfig.ServerList = []string{"http://localhost:2379"}
		if err := options.RunServer(stopCh); err != nil {
			close(serverFailed)
			t.Fatalf("Error in bringing up the server: %v", err)
		}
	}()

	if err := waitForApiserverUp(serverIP, serverFailed); err != nil {
		t.Fatalf("%v", err)
	}

	config := &restclient.Config{}
	config.Host = serverIP
	config.Insecure = true
	clientset, err := servicecatalogclient.NewForConfig(config)
	if nil != err {
		t.Fatal("can't make the client from the config", err)
	}
	return clientset, shutdown
}

func waitForApiserverUp(serverIP string, stopCh <-chan struct{}) error {
	minuteTimeout := time.After(time.Minute)
	for {
		select {
		case <-stopCh:
			return fmt.Errorf("apiserver failed")
		case <-minuteTimeout:
			return fmt.Errorf("waiting for apiserver timed out")
		default:
			glog.Infof("Waiting for : %#v", serverIP)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			_, err := client.Get(serverIP)
			if err == nil {
				return nil
			}
		}
		// no success or overall timeout or stop due to failure
		// wait and go around again
		<-time.After(100 * time.Millisecond)
	}
}
