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
	binding := createBindings(1)[0]
	cred := createCreds(1)[0]
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
	bindings := createBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	inject(t, injector, bindings[0], creds[0])
	inject(t, injector, bindings[1], creds[1])

	secretsCl := injector.client.Core().Secrets(bindings[0].Namespace)
	secret, err := secretsCl.Get(bindings[0].Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, creds[0])

	secretsCl = injector.client.Core().Secrets(bindings[1].Namespace)
	secret, err = secretsCl.Get(bindings[1].Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, creds[1])
}

func TestUninjectOne(t *testing.T) {
	binding := createBindings(1)[0]
	cred := createCreds(1)[0]

	injector := fakeK8sBindingInjector()
	inject(t, injector, binding, cred)
	injector.Uninject(binding)

	testCredentialsUninjected(t, injector, binding)
}

func TestUninjectTwo(t *testing.T) {
	bindings := createBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	inject(t, injector, bindings[0], creds[0])
	inject(t, injector, bindings[1], creds[1])

	injector.Uninject(bindings[0])

	// test that bindings[0] is gone
	testCredentialsUninjected(t, injector, bindings[0])

	//test that bindings[1] is still there
	secretsCl := injector.client.Core().Secrets(bindings[1].Namespace)
	secret, err := secretsCl.Get(bindings[1].Name)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	testCredentialsInjected(t, secret.Data, creds[1])

	// test that bindings[1] is gone after uninject
	injector.Uninject(bindings[1])

	testCredentialsUninjected(t, injector, bindings[1])
}

func createBindings(length int) []*servicecatalog.Binding {
	ret := make([]*servicecatalog.Binding, length, length)
	for i := range ret {
		ret[i] = &servicecatalog.Binding{
			ObjectMeta: kapi.ObjectMeta{
				Name:      "name" + string(i),
				Namespace: "namespace" + string(i),
			},
		}
	}
	return ret
}

func createCreds(length int) []*brokerapi.Credential {
	ret := make([]*brokerapi.Credential, length, length)
	for i := range ret {
		ret[i] = &brokerapi.Credential{
			Hostname: "host" + string(i),
			Port:     "123" + string(i),
			Username: "user" + string(i),
			Password: "password!@#!@#!0)" + string(i),
		}
	}
	return ret
}

func fakeK8sBindingInjector() *k8sBindingInjector {
	return &k8sBindingInjector{
		client: fake.NewSimpleClientset(),
	}
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

	// TODO change so that it's not hard coded to Credential struct fields
	testField("hostname", cred.Hostname)
	testField("port", cred.Port)
	testField("username", cred.Username)
	testField("password", cred.Password)
}

// test that credential is no longer there
func testCredentialsUninjected(t *testing.T, injector *k8sBindingInjector, binding *servicecatalog.Binding) {
	secretsCl := injector.client.Core().Secrets(binding.Namespace)
	_, err := secretsCl.Get(binding.Name)
	if err == nil {
		t.Fatal("Credentials still present after Uninject")
	}
}
