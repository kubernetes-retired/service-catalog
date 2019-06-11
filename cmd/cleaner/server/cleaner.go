/*
Copyright 2019 The Kubernetes Authors.

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

package server

import (
	"fmt"
	"github.com/kubernetes-sigs/service-catalog/pkg/cleaner"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// RunCommand executes one of the command from CleanerOptions
func RunCommand(opt *CleanerOptions) error {
	if err := opt.Validate(); nil != err {
		return err
	}

	k8sKubeconfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client config: %v", err)
	}

	client, err := kubernetes.NewForConfig(k8sKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes client: %s", err)
	}

	scClient, err := sc.NewForConfig(k8sKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get ServiceCatalog client: %v", err)
	}

	apiextClient, err := apiextensionsclientset.NewForConfig(k8sKubeconfig)
	if err != nil {
		return fmt.Errorf("failed to get Apiextensions client: %v", err)
	}

	clr := cleaner.New(client, scClient, apiextClient)

	return clr.RemoveCRDs(opt.ReleaseNamespace, opt.ControllerManagerName, opt.WebhookConfigurationsName())
}
