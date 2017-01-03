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
	"fmt"

	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
)

// The set of kubernetes objects which are injected into a cluster for a
// binding
type injectionSet struct {
	secret *v1.Secret
}

type k8sBindingInjector struct {
	client *kubernetes.Clientset
}

// CreateK8sBindingInjector creates an instance of a BindingInjector which
// manages the injection of binding information within the Kubernetes
// environment.
func CreateK8sBindingInjector() (BindingInjector, error) {
	// This assumes that we are running withing a kubernetes cluster. If this
	// needs to be able to run outside the cluster, it will need to be modified
	// to take a non-default config.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &k8sBindingInjector{
		client: client,
	}, nil
}

func (b *k8sBindingInjector) Inject(binding *servicecatalog.Binding, cred *sbmodel.Credential) error {
	is := makeInjectionSet(binding, cred)

	if err := b.injectSecret(is.secret); err != nil {
		return err
	}

	return nil
}

func (b *k8sBindingInjector) Uninject(binding *servicecatalog.Binding) error {
	return fmt.Errorf("Not implemented")
}

func (b *k8sBindingInjector) injectSecret(s *v1.Secret) error {
	cmc := b.client.Core().Secrets("default")
	_, err := cmc.Create(s)
	return err
}

func makeInjectionSet(binding *servicecatalog.Binding, cred *sbmodel.Credential) *injectionSet {
	secret := makeSecret(binding, cred)

	return &injectionSet{
		secret: secret,
	}
}

func makeSecret(binding *servicecatalog.Binding, cred *sbmodel.Credential) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: binding.Name,
		},
		Data: map[string][]byte{
			"hostname": []byte(cred.Hostname),
			"port":     []byte(cred.Port),
			"username": []byte(cred.Username),
			"password": []byte(cred.Password),
		},
	}
}
