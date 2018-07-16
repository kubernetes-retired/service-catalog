/*
Copyright 2018 The Kubernetes Authors.

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

package svcat

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	k8sclient "k8s.io/client-go/kubernetes"
)

// App is the underlying application behind the svcat cli.
type App struct {
	// SvcatClient is the interface for performing actions against the service-catalog apiserver
	servicecatalog.SvcatClient
	// CurrentNamespace is the namespace set in the current context.
	CurrentNamespace string
}

// NewApp creates an svcat application.
func NewApp(k8sClient k8sclient.Interface, serviceCatalogClient clientset.Interface, ns string) (*App, error) {
	app := &App{
		SvcatClient: &servicecatalog.SDK{
			K8sClient:            k8sClient,
			ServiceCatalogClient: serviceCatalogClient,
		},
		CurrentNamespace: ns,
	}

	return app, nil
}
