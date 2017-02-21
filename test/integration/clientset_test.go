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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/api/v1"

	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"

	// TODO: fix this upstream
	// we shouldn't have to install things to use our own generated client.

	// avoid error `servicecatalog/v1alpha1 is not enabled`
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	// avoid error `no kind is registered for the type v1.ListOptions`
	_ "k8s.io/kubernetes/pkg/api/install"
	// our versioned types
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	// our versioned client
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
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
	if sc, e := scClient.ServiceClasses().Create(&v1alpha1.ServiceClass{}); nil == e {
		t.Fatal("needs a name", sc.Name)
	}
	if i, e := scClient.Instances(ns).Create(&v1alpha1.Instance{}); nil == e {
		t.Fatal("needs a name", i.Name)
	}
	if bi, e := scClient.Bindings(ns).Create(&v1alpha1.Binding{}); nil == e {
		t.Fatal("needs a name", bi.Name)
	}
}

func TestBrokerClient(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	brokerClient := client.Servicecatalog().Brokers()

	broker := &v1alpha1.Broker{
		ObjectMeta: v1.ObjectMeta{Name: "test-broker"},
		Spec: v1alpha1.BrokerSpec{
			URL:     "https://example.com",
			OSBGUID: "OSBGUID field",
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
		t.Fatal("didn't get the same broker twice", brokerServer, brokerListed)
	}

	authSecret := &v1.ObjectReference{
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	brokerServer.Spec.AuthSecret = authSecret

	brokerUpdated, err := brokerClient.Update(brokerServer)
	if nil != err ||
		"test-namespace" != brokerUpdated.Spec.AuthSecret.Namespace ||
		"test-name" != brokerUpdated.Spec.AuthSecret.Name {
		t.Fatal("broker wasn't updated", brokerServer, brokerUpdated)
	}

	brokerServer, err = brokerClient.Get("test-broker")
	if nil != err ||
		"test-namespace" != brokerServer.Spec.AuthSecret.Namespace ||
		"test-name" != brokerServer.Spec.AuthSecret.Name {
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

func TestServiceClassClient(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	serviceClassClient := client.Servicecatalog().ServiceClasses()

	serviceClass := &v1alpha1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{Name: "test-serviceclass"},
		BrokerName: "test-broker",
		Bindable:   true,
	}

	// start from scratch
	serviceClasses, err := serviceClassClient.List(v1.ListOptions{})
	if len(serviceClasses.Items) > 0 {
		t.Fatalf("serviceClasses should not exist on start, had %v serviceClasses", len(serviceClasses.Items))
	}

	serviceClassAtServer, err := serviceClassClient.Create(serviceClass)
	if nil != err {
		t.Fatal("error creating the ServiceClass", serviceClass)
	}
	if serviceClass.Name != serviceClassAtServer.Name {
		t.Fatalf("didn't get the same ServiceClass back from the server \n%+v\n%+v", serviceClass, serviceClassAtServer)
	}

	serviceClasses, err = serviceClassClient.List(v1.ListOptions{})
	if 1 != len(serviceClasses.Items) {
		t.Fatalf("should have exactly one ServiceClass, had %v ServiceClasses", len(serviceClasses.Items))
	}

	serviceClassAtServer, err = serviceClassClient.Get(serviceClass.Name)
	if serviceClassAtServer.Name != serviceClass.Name &&
		serviceClass.ResourceVersion == serviceClassAtServer.ResourceVersion {
		t.Fatalf("didn't get the same ServiceClass back from the server \n%+v\n%+v", serviceClass, serviceClassAtServer)
	}

	err = serviceClassClient.Delete("test-serviceclass", &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("serviceclass should be deleted", err)
	}

	serviceClassDeleted, err := serviceClassClient.Get("test-serviceclass")
	if nil == err {
		t.Fatal("serviceclass should be deleted", serviceClassDeleted)
	}

}

func TestInstanceClient(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	instanceClient := client.Servicecatalog().Instances("test-namespace")

	// Instance represents a provisioned instance of a ServiceClass.
	instance := &v1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{Name: "test-instance"},
		Spec: v1alpha1.InstanceSpec{
			ServiceClassName: "service-class-name",
			PlanName:         "plan-name",
		},
		Status: v1alpha1.InstanceStatus{
			Conditions: []v1alpha1.InstanceCondition{
				{
					Type:    v1alpha1.InstanceConditionReady,
					Status:  v1alpha1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
			},
		},
	}

	instances, err := instanceClient.List(v1.ListOptions{})
	if len(instances.Items) > 0 {
		t.Fatalf("instances should not exist on start, had %v instances", len(instances.Items))
	}

	instanceServer, err := instanceClient.Create(instance)
	if nil != err {
		t.Fatal("error creating the instance", instance)
	}
	if instance.Name != instanceServer.Name {
		t.Fatalf("didn't get the same instance back from the server \n%+v\n%+v", instance, instanceServer)
	}

	instances, err = instanceClient.List(v1.ListOptions{})
	if 1 != len(instances.Items) {
		t.Fatalf("should have exactly one instance, had %v instances", len(instances.Items))
	}

	instanceServer, err = instanceClient.Get(instance.Name)
	if instanceServer.Name != instance.Name &&
		instanceServer.ResourceVersion == instance.ResourceVersion {
		t.Fatalf("didn't get the same instance back from the server \n%+v\n%+v", instance, instanceServer)
	}

	err = instanceClient.Delete("test-instance", &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("instance should be deleted", err)
	}

	instanceDeleted, err := instanceClient.Get("test-instance")
	if nil == err {
		t.Fatal("instance should be deleted", instanceDeleted)
	}
}

func TestBindingClient(t *testing.T) {
	client, shutdownServer := getFreshApiserverAndClient(t)
	defer shutdownServer()
	bindingClient := client.Servicecatalog().Bindings("test-namespace")

	// Binding represents a "used by" relationship between an application
	// and an Instance.
	binding := &v1alpha1.Binding{
		ObjectMeta: v1.ObjectMeta{Name: "test-binding"},
		Spec: v1alpha1.BindingSpec{
			InstanceRef: v1.ObjectReference{
				Name:      "bar",
				Namespace: "test-namespace",
			},
			AppLabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			SecretName:    "secret-name",
			ServiceName:   "service-name",
			ConfigMapName: "configmap-name",
			OSBGUID:       "UUID-string",
		},
		Status: v1alpha1.BindingStatus{
			Conditions: []v1alpha1.BindingCondition{
				{
					Type:    v1alpha1.BindingConditionReady,
					Status:  v1alpha1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
			},
		},
	}

	bindings, err := bindingClient.List(v1.ListOptions{})
	if len(bindings.Items) > 0 {
		t.Fatalf("bindings should not exist on start, had %v bindings", len(bindings.Items))
	}

	bindingServer, err := bindingClient.Create(binding)
	if nil != err {
		t.Fatal("error creating the binding", binding)
	}
	if binding.Name != bindingServer.Name {
		t.Fatalf("didn't get the same binding back from the server \n%+v\n%+v", binding, bindingServer)
	}

	bindings, err = bindingClient.List(v1.ListOptions{})
	if 1 != len(bindings.Items) {
		t.Fatalf("should have exactly one binding, had %v bindings", len(bindings.Items))
	}

	bindingServer, err = bindingClient.Get(binding.Name)
	if bindingServer.Name != binding.Name &&
		bindingServer.ResourceVersion == binding.ResourceVersion {
		t.Fatalf("didn't get the same binding back from the server \n%+v\n%+v", binding, bindingServer)
	}

	err = bindingClient.Delete("test-binding", &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("broker should be deleted", err)
	}

	bindingDeleted, err := bindingClient.Get("test-binding")
	if nil == err {
		t.Fatal("broker should be deleted", bindingDeleted)
	}
}
