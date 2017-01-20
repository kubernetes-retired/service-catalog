/*
Copyright 2016 The Kubernetes Authors.

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

package injector

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"k8s.io/client-go/1.5/kubernetes/fake"
	kapi "k8s.io/kubernetes/pkg/api"
	"testing"
)

func TestInjectOne(t *testing.T) {
	binding, _ := getBindings()
	cred, _ := getCreds()
	injector := fakeK8sBindingInjector()
	inject(t, injector, binding, cred)

	secretsCl := injector.client.Core().Secrets(binding.Namespace)
	secret, err := secretsCl.Get(binding.Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, cred)
}

func TestInjectTwo(t *testing.T) {
	binding0, binding1 := getBindings()
	cred0, cred1 := getCreds()

	injector := fakeK8sBindingInjector()
	inject(t, injector, binding0, cred0)
	inject(t, injector, binding1, cred1)

	secretsCl := injector.client.Core().Secrets(binding0.Namespace)
	secret, err := secretsCl.Get(binding0.Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, cred0)

	secretsCl = injector.client.Core().Secrets(binding1.Namespace)
	secret, err = secretsCl.Get(binding1.Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, cred1)
}

func TestUninjectOne(t *testing.T) {
	binding, _ := getBindings()
	cred, _ := getCreds()

	injector := fakeK8sBindingInjector()
	inject(t, injector, binding, cred)
	injector.Uninject(binding)

	secretsCl := injector.client.Core().Secrets(binding.Namespace)
	secret, err := secretsCl.Get(binding.Name)
	if err == nil {
		testCredentialsUninjected(t, secret.Data)
	}
}

func TestUninjectTwo(t *testing.T) {
	binding0, binding1 := getBindings()
	cred0, cred1 := getCreds()

	injector := fakeK8sBindingInjector()
	inject(t, injector, binding0, cred0)
	inject(t, injector, binding1, cred1)

	injector.Uninject(binding0)

	// test that binding0 is gone
	secretsCl := injector.client.Core().Secrets(binding0.Namespace)
	secret, err := secretsCl.Get(binding0.Name)
	if err == nil {
		testCredentialsUninjected(t, secret.Data)
	}

	//test that binding1 is still there
	secretsCl = injector.client.Core().Secrets(binding1.Namespace)
	secret, err = secretsCl.Get(binding1.Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, cred1)

	// test that binding1 is gone after uninject
	injector.Uninject(binding1)

	secretsCl = injector.client.Core().Secrets(binding1.Namespace)
	secret, err = secretsCl.Get(binding1.Name)
	if err == nil {
		testCredentialsUninjected(t, secret.Data)
	}
}

func fakeK8sBindingInjector() *k8sBindingInjector {
	return &k8sBindingInjector{
		client: fake.NewSimpleClientset(),
	}
}

func getBindings() (*servicecatalog.Binding, *servicecatalog.Binding) {
	binding0 := &servicecatalog.Binding{
		ObjectMeta: kapi.ObjectMeta{
			Name:      "name0",
			Namespace: "namespace0",
		},
	}
	binding1 := &servicecatalog.Binding{
		ObjectMeta: kapi.ObjectMeta{
			Name:      "name1",
			Namespace: "namespace1",
		},
	}
	return binding0, binding1
}

func getCreds() (*brokerapi.Credential, *brokerapi.Credential) {
	cred0 := &brokerapi.Credential{
		Hostname: "host0",
		Port:     "123",
		Username: "user0",
		Password: "password!@#!@#!0)",
	}
	cred1 := &brokerapi.Credential{
		Hostname: "host1",
		Port:     "456",
		Username: "user1",
		Password: "password*(&*1)",
	}
	return cred0, cred1
}

func inject(t *testing.T, injector BindingInjector,
	binding *servicecatalog.Binding, cred *brokerapi.Credential) {

	err := injector.Inject(binding, cred)
	if err != nil {
		t.Fatalf("Error when injecting credentials: %s", err)
	}
}

// tests all fields of credentials are there and also the same value
func testCredentialsInjected(t *testing.T, data map[string][]byte, cred *brokerapi.Credential) {
	testField := func(key string, expectedValue string) {
		val, ok := data[key]
		if !ok {
			t.Errorf("%s not in secret after injecting", key)
		} else if string(val) != expectedValue {
			t.Errorf("%s does not match. Expected: %s; Actual: %s", key, expectedValue, val)
		}
	}

	testField("hostname", cred.Hostname)
	testField("port", cred.Port)
	testField("username", cred.Username)
	testField("password", cred.Password)
}

// test that fields from credential is no longer there
func testCredentialsUninjected(t *testing.T, data map[string][]byte) {
	testField := func(key string) {
		_, ok := data[key]
		if ok {
			t.Errorf("%s found in map when it's expected to not be there", key)
		}
	}

	testField("hostname")
	testField("port")
	testField("username")
	testField("password")
}
