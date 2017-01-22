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
	"errors"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"k8s.io/client-go/1.5/kubernetes/fake"
	v1 "k8s.io/client-go/1.5/pkg/api/v1"
	kapi "k8s.io/kubernetes/pkg/api"
	"testing"
)

func TestInjectOne(t *testing.T) {
	binding := createBindings(1)[0]
	cred := createCreds(1)[0]
	injector := fakeK8sBindingInjector()
	if err := inject(injector, binding, cred); err != nil {
		t.Fatal(err)
	}

	secret, err := getSecret(injector, binding)
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	if err := testCredentialsInjected(secret.Data, cred); err != nil {
		t.Error(err)
	}
}

func TestInjectTwo(t *testing.T) {
	bindings := createBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	if err := inject(injector, bindings[0], creds[0]); err != nil {
		t.Fatal(err)
	}
	if err := inject(injector, bindings[1], creds[1]); err != nil {
		t.Fatal(err)
	}

	secret, err := getSecret(injector, bindings[0])
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	if err := testCredentialsInjected(secret.Data, creds[0]); err != nil {
		t.Error(err)
	}

	secret, err = getSecret(injector, bindings[1])
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	if err := testCredentialsInjected(secret.Data, creds[1]); err != nil {
		t.Error(err)
	}
}

func TestUninjectOne(t *testing.T) {
	binding := createBindings(1)[0]
	cred := createCreds(1)[0]

	injector := fakeK8sBindingInjector()
	if err := inject(injector, binding, cred); err != nil {
		t.Fatal(err)
	}
	injector.Uninject(binding)

	if err := testCredentialsUninjected(injector, binding); err != nil {
		t.Fatal(err)
	}
}

func TestUninjectTwo(t *testing.T) {
	bindings := createBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	if err := inject(injector, bindings[0], creds[0]); err != nil {
		t.Fatal(err)
	}
	if err := inject(injector, bindings[1], creds[1]); err != nil {
		t.Fatal(err)
	}

	injector.Uninject(bindings[0])

	// test that bindings[0] is gone
	if err := testCredentialsUninjected(injector, bindings[0]); err != nil {
		t.Fatal(err)
	}

	//test that bindings[1] is still there
	secret, err := getSecret(injector, bindings[1])
	if err != nil {
		t.Fatalf("Error when getting secret: %s", err)
	}
	if err := testCredentialsInjected(secret.Data, creds[1]); err != nil {
		t.Error(err)
	}

	// test that bindings[1] is gone after uninject
	injector.Uninject(bindings[1])

	if err := testCredentialsUninjected(injector, bindings[1]); err != nil {
		t.Fatal(err)
	}
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

func getSecret(injector *k8sBindingInjector, binding *servicecatalog.Binding) (*v1.Secret, error) {
	secretsCl := injector.client.Core().Secrets(binding.Namespace)
	return secretsCl.Get(binding.Name)
}

func inject(injector BindingInjector,
	binding *servicecatalog.Binding, cred *brokerapi.Credential) error {

	err := injector.Inject(binding, cred)
	if err != nil {
		return fmt.Errorf("Error when injecting credentials: %s", err)
	}
	return nil
}

// tests all fields of credentials are there and also the same value
func testCredentialsInjected(data map[string][]byte, cred *brokerapi.Credential) error {
	testField := func(key string, expectedValue string) error {
		val, ok := data[key]
		if !ok {
			return fmt.Errorf("%s not in secret after injecting", key)
		} else if string(val) != expectedValue {
			return fmt.Errorf("%s does not match. Expected: %s; Actual: %s",
				key, expectedValue, val)
		}
		return nil
	}

	// TODO change so that it's not hard coded to Credential struct fields
	if err := testField("hostname", cred.Hostname); err != nil {
		return err
	}
	if err := testField("port", cred.Port); err != nil {
		return err
	}
	if err := testField("username", cred.Username); err != nil {
		return err
	}
	if err := testField("password", cred.Password); err != nil {
		return err
	}
	return nil
}

// test that credential is no longer there
func testCredentialsUninjected(injector *k8sBindingInjector, binding *servicecatalog.Binding) error {
	_, err := getSecret(injector, binding)
	if err == nil {
		return errors.New("Credentials still present after Uninject")
	}
	return nil
}
