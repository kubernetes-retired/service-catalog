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
	"reflect"
	"strconv"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"k8s.io/client-go/1.5/kubernetes/fake"
	v1 "k8s.io/client-go/1.5/pkg/api/v1"
	kapi "k8s.io/kubernetes/pkg/api"
)

func TestCreateSerializedSecret(t *testing.T) {
	cases := []struct {
		name   string
		cred   *brokerapi.Credential
		secret *v1.Secret
	}{
		{
			name: "string type",
			cred: &brokerapi.Credential{
				"Hostname": "host",
				"Port":     "123",
				"Username": "user",
				"Password": "password!@#!@#!0)",
			},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"Hostname": []byte("host"),
					"Port":     []byte("123"),
					"Username": []byte("user"),
					"Password": []byte("password!@#!@#!0)"),
				},
			},
		},
		{
			name: "float type",
			cred: &brokerapi.Credential{
				"Hostname": "host",
				"Port":     "123",
				"Username": "user",
				"Password": 1.23,
			},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"Hostname": []byte("host"),
					"Port":     []byte("123"),
					"Username": []byte("user"),
					"Password": []byte("1.23"),
				},
			},
		},
		{
			name: "int type",
			cred: &brokerapi.Credential{
				"Hostname": "host",
				"Port":     123,
				"Username": "user",
				"Password": 1,
			},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"Hostname": []byte("host"),
					"Port":     []byte("123"),
					"Username": []byte("user"),
					"Password": []byte("1"),
				},
			},
		},
		{
			name: "slice type",
			cred: &brokerapi.Credential{
				"Hostname": "host",
				"Port":     "123",
				"Username": "user",
				"Password": []string{"one", "two"},
			},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"Hostname": []byte("host"),
					"Port":     []byte("123"),
					"Username": []byte("user"),
					"Password": []byte(`["one","two"]`),
				},
			},
		},
		{
			name: "map type",
			cred: &brokerapi.Credential{
				"Hostname": "host",
				"Port":     "123",
				"Username": "user",
				"Password": map[string]int{"one": 1, "two": 2},
			},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"Hostname": []byte("host"),
					"Port":     []byte("123"),
					"Username": []byte("user"),
					"Password": []byte(`{"one":1,"two":2}`),
				},
			},
		},
	}

	namespace := "test-namespace"
	binding := &servicecatalog.Binding{
		ObjectMeta: kapi.ObjectMeta{
			Name:      "binding-name",
			Namespace: namespace,
		},
		Spec: servicecatalog.BindingSpec{
			InstanceRef: kapi.ObjectReference{
				Namespace: namespace,
			},
		},
	}
	for _, tc := range cases {
		actualSecret, err := createSerializedSecret(binding, tc.cred)
		if err != nil {
			t.Errorf("%s: unexpected error making secret: %v", tc.name, err)
		}
		if e, a := tc.secret, actualSecret; !reflect.DeepEqual(e.Data, a.Data) {
			t.Errorf("%s: expected and actual secret data do not match", tc.name)
		}

	}
}

func TestInjectOne(t *testing.T) {
	binding := createFakeBindings(1)[0]
	cred := createCreds(1)[0]
	injector := fakeK8sBindingInjector()
	if err := injector.Inject(binding, cred); err != nil {
		t.Fatal(err)
	}

	if err := testCredentialsInjected(injector, binding, cred); err != nil {
		t.Error(err)
	}
}

func TestInjectTwo(t *testing.T) {
	bindings := createFakeBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	if err := injector.Inject(bindings[0], creds[0]); err != nil {
		t.Fatal(err)
	}
	if err := injector.Inject(bindings[1], creds[1]); err != nil {
		t.Fatal(err)
	}

	if err := testCredentialsInjected(injector, bindings[0], creds[0]); err != nil {
		t.Error(err)
	}

	if err := testCredentialsInjected(injector, bindings[1], creds[1]); err != nil {
		t.Error(err)
	}
}

func TestInjectOverride(t *testing.T) {
	binding := createFakeBindings(1)[0]
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	if err := injector.Inject(binding, creds[0]); err != nil {
		t.Fatal(err)
	}

	// note that we expect a failure here
	if err := injector.Inject(binding, creds[0]); err == nil {
		t.Fatal("Injecting over the same binding succeeded even though it shouldn't")
	}
}

func TestUninjectEmpty(t *testing.T) {
	binding := createFakeBindings(1)[0]
	injector := fakeK8sBindingInjector()
	if err := injector.Uninject(binding); err == nil {
		t.Fatal("Uninject empty expected error but none returned!")
	}
}

func TestUninjectOne(t *testing.T) {
	binding := createFakeBindings(1)[0]
	cred := createCreds(1)[0]

	injector := fakeK8sBindingInjector()
	if err := injector.Inject(binding, cred); err != nil {
		t.Fatal(err)
	}
	if err := injector.Uninject(binding); err != nil {
		t.Fatal("Unexpected error when uninjecting")
	}

	if err := testCredentialsUninjected(injector, binding); err != nil {
		t.Fatal(err)
	}
}

func TestUninjectSame(t *testing.T) {
	binding := createFakeBindings(1)[0]
	cred := createCreds(1)[0]

	injector := fakeK8sBindingInjector()
	if err := injector.Inject(binding, cred); err != nil {
		t.Fatal(err)
	}
	if err := injector.Uninject(binding); err != nil {
		t.Fatal("Unexpected err when uninjecting:", err)
	}
	if err := injector.Uninject(binding); err == nil {
		t.Fatal("Expected err when uninjecting twice but none found!")
	}
}

func TestUninjectTwo(t *testing.T) {
	bindings := createFakeBindings(2)
	creds := createCreds(2)

	injector := fakeK8sBindingInjector()
	if err := injector.Inject(bindings[0], creds[0]); err != nil {
		t.Fatal(err)
	}
	if err := injector.Inject(bindings[1], creds[1]); err != nil {
		t.Fatal(err)
	}

	if err := injector.Uninject(bindings[0]); err != nil {
		t.Fatal("Unexpected err when uninjecting")
	}

	// test that bindings[0] is gone
	if err := testCredentialsUninjected(injector, bindings[0]); err != nil {
		t.Fatal(err)
	}

	//test that bindings[1] is still there
	if err := testCredentialsInjected(injector, bindings[1], creds[1]); err != nil {
		t.Error(err)
	}

	// test that bindings[1] is gone after uninject
	if err := injector.Uninject(bindings[1]); err != nil {
		t.Fatal("Unexpected err when uninjecting")
	}

	if err := testCredentialsUninjected(injector, bindings[1]); err != nil {
		t.Fatal(err)
	}
}

func createFakeBindings(length int) []*servicecatalog.Binding {
	ret := make([]*servicecatalog.Binding, length, length)
	for i := range ret {
		namespace := "namespace" + strconv.Itoa(i)
		ret[i] = &servicecatalog.Binding{
			ObjectMeta: kapi.ObjectMeta{
				Name:      "name" + strconv.Itoa(i),
				Namespace: namespace,
			},
			Spec: servicecatalog.BindingSpec{
				InstanceRef: kapi.ObjectReference{
					Namespace: namespace,
				},
			},
		}
	}
	return ret
}

func createCreds(length int) []*brokerapi.Credential {
	ret := make([]*brokerapi.Credential, length, length)
	for i := range ret {
		ret[i] = &brokerapi.Credential{
			"Hostname": "host" + strconv.Itoa(i),
			"Port":     "123" + strconv.Itoa(i),
			"Username": "user" + strconv.Itoa(i),
			"Password": "password!@#!@#!0)" + strconv.Itoa(i),
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

// tests all fields of credentials are there and also the same value
func testCredentialsInjected(injector *k8sBindingInjector, binding *servicecatalog.Binding, cred *brokerapi.Credential) error {
	secret, err := getSecret(injector, binding)
	if err != nil {
		return err
	}
	for k := range *cred {
		_, ok := secret.Data[k]
		if !ok {
			return fmt.Errorf("%s not in secret after injecting", k)
		}
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
