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

package util

import (
	"fmt"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/probe"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
)

// WaitForServiceCatalogCRDs waits for Service Catalog CRDs to be activated in main api-server
// because it may take some time before Catalog CRDs registration shows up.
// It is useful to ensure that CRDs are ready before creating Service Catalog clients/informers.
func WaitForServiceCatalogCRDs(restConfig *restclient.Config) error {
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create apiextension clientset: %v", err)
	}

	readinessProbe := probe.NewCRDProbe(apiextensionsClient, 0)

	// Attempt to get resources every 10 seconds and quit after 3 minutes if unsuccessful.
	err = wait.PollImmediate(10*time.Second, 3*time.Minute, readinessProbe.IsReady)
	if err != nil {
		if err == wait.ErrWaitTimeout {
			return fmt.Errorf("CRDs are not available")
		}
		return err
	}
	return nil
}
