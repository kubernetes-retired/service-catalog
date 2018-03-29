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
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/kube"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
)

// App is the underlying application behind the svcat cli.
type App struct {
	*servicecatalog.SDK

	// CurrentNamespace is the namespace set in the current context.
	CurrentNamespace string
}

// NewApp creates an svcat application.
func NewApp(kubeConfig, kubeContext string) (*App, error) {
	// Initialize a service catalog client
	cl, ns, err := getServiceCatalogClient(kubeConfig, kubeContext)
	if err != nil {
		return nil, err
	}

	app := &App{
		SDK: &servicecatalog.SDK{
			ServiceCatalogClient: cl,
		},
		CurrentNamespace: ns,
	}

	return app, nil
}

// getServiceCatalogClient creates a Service Catalog config and client for a given kubeconfig context.
func getServiceCatalogClient(kubeConfig, kubeContext string) (client *clientset.Clientset, namespaces string, err error) {
	config := kube.GetConfig(kubeContext, kubeConfig)

	currentNamespace, _, err := config.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("could not determine the namespace for the current context %q: %s", kubeContext, err)
	}

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("could not get Kubernetes config for context %q: %s", kubeContext, err)
	}

	client, err = clientset.NewForConfig(restConfig)
	return client, currentNamespace, err
}
