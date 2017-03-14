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

package server

import (
	"errors"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/storage/tpr"

	kubeclientfake "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5/fake"
	"k8s.io/kubernetes/pkg/client/testing/core"
	"k8s.io/kubernetes/pkg/runtime"
)

// make sure RunServer returns with an error when TPR fails to install
func TestRunServerInstallTPRFails(t *testing.T) {
	options := &ServiceCatalogServerOptions{}

	fakeClientset := &kubeclientfake.Clientset{}
	fakeClientset.AddReactor("get", "thirdpartyresources", func(core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("TPR not found")
	})
	fakeClientset.AddReactor("create", "thirdpartyresources", func(core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("Failed to create TPR")
	})

	options.StorageTypeString = "tpr"
	options.TPROptions = &TPROptions{
		"default-name-space",
		fakeClientset.Core().RESTClient(),
		installTPRsToCore(fakeClientset),
		"name-space",
	}

	err := RunServer(options)
	if _, ok := err.(tpr.ErrTPRInstall); !ok {
		t.Errorf("API Server did not report failure after failing to install Third Party Resources")
	}

	// make sure no more action after tpr failed to install
	actions := fakeClientset.Actions()
	for _, action := range actions {
		if action.GetResource().Resource != "thirdpartyresources" {
			t.Errorf("Unexpected action performed after failing to install third party resource")
		}
	}
}
