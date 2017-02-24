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
	"encoding/json"
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

	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/diff"
)

// Used for testing instance parameters
type ipStruct struct {
	Bar    string            `json:"bar"`
	Values map[string]string `json:"values"`
}

var instanceParameter = `{
    "bar": "barvalue",
    "values": {
      "first" : "firstvalue",
      "second" : "secondvalue"
    }
  }
`

// Used for testing binding parameters
type bpStruct struct {
	Foo string   `json:"foo"`
	Baz []string `json:"baz"`
}

var bindingParameter = `{
    "foo": "bar",
    "baz": [
      "first",
      "second"
    ]
  }
`

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

// TestBrokerClient exercises the Broker client.
func TestBrokerClient(t *testing.T) {
	const name = "test-broker"

	var (
		client, shutdownServer = getFreshApiserverAndClient(t)
		brokerClient           = client.Servicecatalog().Brokers()

		broker = &v1alpha1.Broker{
			ObjectMeta: v1.ObjectMeta{Name: name},
			Spec: v1alpha1.BrokerSpec{
				URL:     "https://example.com",
				OSBGUID: "OSBGUID field",
			},
		}
	)
	defer shutdownServer()

	// start from scratch
	brokers, err := brokerClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing brokers (%s)", err)
	}
	if len(brokers.Items) > 0 {
		t.Fatalf("brokers should not exist on start, had %v brokers", len(brokers.Items))
	}

	brokerServer, err := brokerClient.Create(broker)
	if nil != err {
		t.Fatalf("error creating the broker '%s' (%s)", broker, err)
	}
	if name != brokerServer.Name {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}

	brokers, err = brokerClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing brokers (%s)", err)
	}
	if 1 != len(brokers.Items) {
		t.Fatalf("should have exactly one broker, had %v brokers", len(brokers.Items))
	}

	brokerServer, err = brokerClient.Get(name)
	if err != nil {
		t.Fatalf("error getting broker %s (%s)", name, err)
	}
	if name != brokerServer.Name &&
		broker.ResourceVersion == brokerServer.ResourceVersion {
		t.Fatalf("didn't get the same broker back from the server \n%+v\n%+v", broker, brokerServer)
	}

	// check that the broker is the same from get and list
	brokerListed := &brokers.Items[0]
	if !reflect.DeepEqual(brokerServer, brokerListed) {
		t.Fatalf("Didn't get the same instance from list and get: diff: %v", diff.ObjectReflectDiff(brokerServer, brokerListed))
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

	readyConditionTrue := v1alpha1.BrokerCondition{
		Type:    v1alpha1.BrokerConditionReady,
		Status:  v1alpha1.ConditionTrue,
		Reason:  "ConditionReason",
		Message: "ConditionMessage",
	}
	brokerUpdated.Status = v1alpha1.BrokerStatus{
		Conditions: []v1alpha1.BrokerCondition{
			readyConditionTrue,
		},
	}
	brokerUpdated.Spec.URL = "http://shouldnotupdate.com"

	brokerUpdated2, err := brokerClient.UpdateStatus(brokerUpdated)
	if nil != err || len(brokerUpdated2.Status.Conditions) != 1 {
		t.Fatal("broker status wasn't updated")
	}
	if e, a := readyConditionTrue, brokerUpdated2.Status.Conditions[0]; !reflect.DeepEqual(e, a) {
		t.Fatal("Didn't get matching ready conditions:\nexpected: %v\n\ngot: %v", e, a)
	}
	if e, a := "https://example.com", brokerUpdated2.Spec.URL; e != a {
		t.Fatalf("Should not be able to update spec from status subresource")
	}

	readyConditionFalse := v1alpha1.BrokerCondition{
		Type:    v1alpha1.BrokerConditionReady,
		Status:  v1alpha1.ConditionFalse,
		Reason:  "ConditionReason",
		Message: "ConditionMessage",
	}
	brokerUpdated2.Status.Conditions[0] = readyConditionFalse

	brokerUpdated3, err := brokerClient.UpdateStatus(brokerUpdated2)
	if nil != err || len(brokerUpdated3.Status.Conditions) != 1 {
		t.Fatal("broker status wasn't updated")
	}

	brokerServer, err = brokerClient.Get(name)
	if nil != err ||
		"test-namespace" != brokerServer.Spec.AuthSecret.Namespace ||
		"test-name" != brokerServer.Spec.AuthSecret.Name {
		t.Fatal("broker wasn't updated", brokerServer)
	}
	if e, a := readyConditionFalse, brokerServer.Status.Conditions[0]; !reflect.DeepEqual(e, a) {
		t.Fatal("Didn't get matching ready conditions:\nexpected: %v\n\ngot: %v", e, a)
	}

	err = brokerClient.Delete(name, &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("broker should be deleted", err)
	}

	brokerDeleted, err := brokerClient.Get(name)
	if nil == err {
		t.Fatal("broker should be deleted", brokerDeleted)
	}
}

// TestServiceClassClient exercises the ServiceClass client.
func TestServiceClassClient(t *testing.T) {
	const name = "test-serviceclass"

	var (
		client, shutdownServer = getFreshApiserverAndClient(t)
		serviceClassClient     = client.Servicecatalog().ServiceClasses()

		serviceClass = &v1alpha1.ServiceClass{
			ObjectMeta: v1.ObjectMeta{Name: name},
			BrokerName: "test-broker",
			Bindable:   true,
		}
	)
	defer shutdownServer()

	// start from scratch
	serviceClasses, err := serviceClassClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing service classes (%s)", err)
	}
	if len(serviceClasses.Items) > 0 {
		t.Fatalf("serviceClasses should not exist on start, had %v serviceClasses", len(serviceClasses.Items))
	}

	serviceClassAtServer, err := serviceClassClient.Create(serviceClass)
	if nil != err {
		t.Fatal("error creating the ServiceClass", serviceClass)
	}
	if name != serviceClassAtServer.Name {
		t.Fatalf("didn't get the same ServiceClass back from the server \n%+v\n%+v", serviceClass, serviceClassAtServer)
	}

	serviceClasses, err = serviceClassClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing service classes (%s)", err)
	}
	if 1 != len(serviceClasses.Items) {
		t.Fatalf("should have exactly one ServiceClass, had %v ServiceClasses", len(serviceClasses.Items))
	}

	serviceClassAtServer, err = serviceClassClient.Get(name)
	if err != nil {
		t.Fatalf("error listing service classes (%s)", err)
	}
	if serviceClassAtServer.Name != name &&
		serviceClass.ResourceVersion == serviceClassAtServer.ResourceVersion {
		t.Fatalf("didn't get the same ServiceClass back from the server \n%+v\n%+v", serviceClass, serviceClassAtServer)
	}

	// check that the broker is the same from get and list
	serviceClassListed := &serviceClasses.Items[0]
	if !reflect.DeepEqual(serviceClassAtServer, serviceClassListed) {
		t.Fatalf("Didn't get the same instance from list and get: diff: %v", diff.ObjectReflectDiff(serviceClassAtServer, serviceClassListed))
	}

	serviceClassAtServer.Bindable = false
	_, err = serviceClassClient.Update(serviceClassAtServer)
	if err != nil {
		t.Fatalf("Error updating serviceClass: %v", err)
	}
	updated, err := serviceClassClient.Get(name)
	if err != nil {
		t.Fatalf("Error getting serviceClass: %v", err)
	}
	if updated.Bindable {
		t.Fatal("Failed to update service class")
	}

	err = serviceClassClient.Delete(name, &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("serviceclass should be deleted", err)
	}

	serviceClassDeleted, err := serviceClassClient.Get(name)
	if nil == err {
		t.Fatal("serviceclass should be deleted", serviceClassDeleted)
	}

	serviceClassDeleted, err = serviceClassClient.Get(name)
	if nil == err {
		t.Fatal("serviceclass should be deleted", serviceClassDeleted)
	}
}

// TestInstanceClient exercises the Instance client.
func TestInstanceClient(t *testing.T) {
	const name = "test-instance"

	var (
		client, shutdownServer = getFreshApiserverAndClient(t)
		instanceClient         = client.Servicecatalog().Instances("test-namespace")

		instance = &v1alpha1.Instance{
			ObjectMeta: v1.ObjectMeta{Name: name},
			Spec: v1alpha1.InstanceSpec{
				ServiceClassName: "service-class-name",
				PlanName:         "plan-name",
				Parameters:       runtime.RawExtension{Raw: []byte(instanceParameter)},
			},
		}
	)
	defer shutdownServer()

	instances, err := instanceClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing instances (%s)", err)
	}
	if len(instances.Items) > 0 {
		t.Fatalf("instances should not exist on start, had %v instances", len(instances.Items))
	}

	instanceServer, err := instanceClient.Create(instance)
	if nil != err {
		t.Fatal("error creating the instance", instance)
	}
	if name != instanceServer.Name {
		t.Fatalf("didn't get the same instance back from the server \n%+v\n%+v", instance, instanceServer)
	}

	instances, err = instanceClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing instances (%s)", err)
	}
	if 1 != len(instances.Items) {
		t.Fatalf("should have exactly one instance, had %v instances", len(instances.Items))
	}

	instanceServer, err = instanceClient.Get(name)
	if err != nil {
		t.Fatalf("error getting instance (%s)", err)
	}
	if instanceServer.Name != name &&
		instanceServer.ResourceVersion == instance.ResourceVersion {
		t.Fatalf("didn't get the same instance back from the server \n%+v\n%+v", instance, instanceServer)
	}

	instanceListed := &instances.Items[0]
	if !reflect.DeepEqual(instanceListed, instanceServer) {
		t.Fatalf("Didn't get the same instance from list and get: diff: %v", diff.ObjectReflectDiff(instanceListed, instanceServer))
	}

	parameters := ipStruct{}
	err = json.Unmarshal(instanceServer.Spec.Parameters.Raw, &parameters)
	if err != nil {
		t.Fatal("Couldn't unmarshal returned instance parameters: %v", err)
	}
	if parameters.Bar != "barvalue" {
		t.Fatal("Didn't get back 'barvalue' value for key 'bar' was %+v", parameters)
	}
	if len(parameters.Values) != 2 {
		t.Fatal("Didn't get back 'barvalue' value for key 'bar' was %+v", parameters)
	}
	if parameters.Values["first"] != "firstvalue" {
		t.Fatal("Didn't get back 'firstvalue' value for key 'first' in Values map was %+v", parameters)
	}
	if parameters.Values["second"] != "secondvalue" {
		t.Fatal("Didn't get back 'secondvalue' value for key 'second' in Values map was %+v", parameters)
	}

	readyConditionTrue := v1alpha1.InstanceCondition{
		Type:    v1alpha1.InstanceConditionReady,
		Status:  v1alpha1.ConditionTrue,
		Reason:  "ConditionReason",
		Message: "ConditionMessage",
	}
	instanceServer.Status = v1alpha1.InstanceStatus{
		Conditions: []v1alpha1.InstanceCondition{readyConditionTrue},
	}
	_, err = instanceClient.UpdateStatus(instanceServer)
	if err != nil {
		t.Fatalf("Error updating instance: %v", err)
	}
	instanceServer, err = instanceClient.Get(name)
	if err != nil {
		t.Fatalf("error getting instance (%s)", err)
	}
	if e, a := readyConditionTrue, instanceServer.Status.Conditions[0]; !reflect.DeepEqual(e, a) {
		t.Fatal("Didn't get matching ready conditions:\nexpected: %v\n\ngot: %v", e, a)
	}

	err = instanceClient.Delete(name, &v1.DeleteOptions{})
	if nil != err {
		t.Fatal("instance should be deleted", err)
	}

	instanceDeleted, err := instanceClient.Get(name)
	if nil == err {
		t.Fatal("instance should be deleted", instanceDeleted)
	}
}

// TestBindingClient exercises the Binding client.
func TestBindingClient(t *testing.T) {
	const name = "test-binding"

	var (
		client, shutdownServer = getFreshApiserverAndClient(t)
		bindingClient          = client.Servicecatalog().Bindings("test-namespace")

		binding = &v1alpha1.Binding{
			ObjectMeta: v1.ObjectMeta{Name: "test-binding"},
			Spec: v1alpha1.BindingSpec{
				InstanceRef: v1.ObjectReference{
					Name:      "bar",
					Namespace: "test-namespace",
				},
				AppLabelSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar"},
				},
				Parameters:    runtime.RawExtension{Raw: []byte(bindingParameter)},
				SecretName:    "secret-name",
				ServiceName:   "service-name",
				ConfigMapName: "configmap-name",
				OSBGUID:       "UUID-string",
			},
		}
	)
	defer shutdownServer()

	bindings, err := bindingClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing bindings (%s)", err)
	}
	if len(bindings.Items) > 0 {
		t.Fatalf("bindings should not exist on start, had %v bindings", len(bindings.Items))
	}

	bindingServer, err := bindingClient.Create(binding)
	if nil != err {
		t.Fatal("error creating the binding", binding)
	}
	if name != bindingServer.Name {
		t.Fatalf("didn't get the same binding back from the server \n%+v\n%+v", binding, bindingServer)
	}

	bindings, err = bindingClient.List(v1.ListOptions{})
	if err != nil {
		t.Fatalf("error listing bindings (%s)", err)
	}
	if 1 != len(bindings.Items) {
		t.Fatalf("should have exactly one binding, had %v bindings", len(bindings.Items))
	}

	bindingServer, err = bindingClient.Get(name)
	if err != nil {
		t.Fatalf("error getting binding (%s)", err)
	}
	if bindingServer.Name != name &&
		bindingServer.ResourceVersion == binding.ResourceVersion {
		t.Fatalf("didn't get the same binding back from the server \n%+v\n%+v", binding, bindingServer)
	}

	bindingListed := &bindings.Items[0]
	if !reflect.DeepEqual(bindingListed, bindingServer) {
		t.Fatalf("Didn't get the same binding from list and get: diff: %v", diff.ObjectReflectDiff(bindingListed, bindingServer))
	}

	parameters := bpStruct{}
	err = json.Unmarshal(bindingServer.Spec.Parameters.Raw, &parameters)
	if err != nil {
		t.Fatal("Couldn't unmarshal returned parameters: %v", err)
	}
	if parameters.Foo != "bar" {
		t.Fatal("Didn't get back 'bar' value for key 'foo' was %+v", parameters)
	}
	if len(parameters.Baz) != 2 {
		t.Fatal("Didn't get back two values for 'baz' array in parameters was %+v", parameters)
	}
	foundFirst := false
	foundSecond := false
	for _, val := range parameters.Baz {
		if val == "first" {
			foundFirst = true
		}
		if val == "second" {
			foundSecond = true
		}
	}
	if !foundFirst {
		t.Fatal("Didn't find first value in parameters.baz was %+v", parameters)
	}
	if !foundSecond {
		t.Fatal("Didn't find second value in parameters.baz was %+v", parameters)
	}

	readyConditionTrue := v1alpha1.BindingCondition{
		Type:    v1alpha1.BindingConditionReady,
		Status:  v1alpha1.ConditionTrue,
		Reason:  "ConditionReason",
		Message: "ConditionMessage",
	}
	bindingServer.Status = v1alpha1.BindingStatus{
		Conditions: []v1alpha1.BindingCondition{readyConditionTrue},
	}
	if _, err = bindingClient.UpdateStatus(bindingServer); err != nil {
		t.Fatalf("Error updating binding: %v", err)
	}
	bindingServer, err = bindingClient.Get(name)
	if err != nil {
		t.Fatalf("error getting binding: %v", err)
	}
	if e, a := readyConditionTrue, bindingServer.Status.Conditions[0]; !reflect.DeepEqual(e, a) {
		t.Fatal("Didn't get matching ready conditions:\nexpected: %v\n\ngot: %v", e, a)
	}

	if err = bindingClient.Delete(name, &v1.DeleteOptions{}); nil != err {
		t.Fatal("broker should be deleted", err)
	}

	bindingDeleted, err := bindingClient.Get(name)
	if nil == err {
		t.Fatal("binding should be deleted", bindingDeleted)
	}
}
