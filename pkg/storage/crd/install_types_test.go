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

package crd

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	core "k8s.io/client-go/testing"
)

func setup(getFn, createFn func(core.Action) (bool, runtime.Object, error)) *crdclientfake.Clientset {
	fakeClientset := &crdclientfake.Clientset{}

	fakeClientset.AddReactor("get", "customresourcedefinitions", getFn)

	fakeClientset.AddReactor("create", "customresourcedefinitions", createFn)

	return fakeClientset
}

//make sure all resources types are installed
func TestInstallTypesAllResources(t *testing.T) {
	getCallCount := 0
	createCallCount := 0

	fakeClientset := setup(
		func(core.Action) (bool, runtime.Object, error) {
			getCallCount++
			// if 'create' has been called on all crds, return 'nil' error to indicate crd is created
			if createCallCount == len(customResourceDefinitions) {
				return true, &crdv1beta1.CustomResourceDefinition{}, nil
			}

			// return error to indicate crd is not found
			return true, nil, errors.New("Resource not found")
		},
		func(core.Action) (bool, runtime.Object, error) {
			createCallCount++
			return true, nil, nil
		},
	)

	if err := InstallTypes(fakeClientset.ApiextensionsV1beta1().CustomResourceDefinitions()); err != nil {
		t.Fatalf("error installing types (%s)", err)
	}

	expectTotal := len(customResourceDefinitions)
	if createCallCount != expectTotal {
		t.Errorf("Expected %d Custom Resources created instead of %d", expectTotal, createCallCount)
	}
}

//make sure to skip resource that is already installed
func TestInstallTypesResourceExisted(t *testing.T) {
	getCallCount := 0
	createCallCount := 0
	createCallArgs := []string{}

	fakeClientset := setup(
		func(core.Action) (bool, runtime.Object, error) {
			getCallCount++
			if getCallCount == 1 {
				// return broker CRD on 1st call to indicate broker CRD exists
				return true, serviceBrokerCRD, nil
			} else if createCallCount == len(customResourceDefinitions)-1 {
				// once 'create' has been called on all crds, return 'nil' error to indicate crd is created
				return true, &crdv1beta1.CustomResourceDefinition{}, nil
			}

			return true, nil, errors.New("Resource not found")
		},
		func(action core.Action) (bool, runtime.Object, error) {
			createCallCount++
			createCallArgs = append(createCallArgs, action.(core.CreateAction).GetObject().(*crdv1beta1.CustomResourceDefinition).Name)
			return true, nil, nil
		},
	)

	if err := InstallTypes(fakeClientset.ApiextensionsV1beta1().CustomResourceDefinitions()); err != nil {
		t.Fatalf("error installing (%s)", err)
	}

	if createCallCount != len(customResourceDefinitions)-1 {
		t.Errorf("Failed to skip 1 installed Custom Resource")
	}

	for _, name := range createCallArgs {
		if name == serviceBrokerCRD.Name {
			t.Errorf("Failed to skip installing 'broker' as Custom Resource as it already existed")
		}
	}
}

//make sure all errors are received for all failed install
func TestInstallTypesErrors(t *testing.T) {
	getCallCount := 0
	createCallCount := 0

	fakeClientset := setup(
		func(core.Action) (bool, runtime.Object, error) {
			getCallCount++
			// if 'create' has been called on all crds, return 'nil' error to indicate crd is created
			if createCallCount == len(customResourceDefinitions) {
				return true, &crdv1beta1.CustomResourceDefinition{}, nil
			}

			// return error to indicate crd is not found
			return true, nil, errors.New("Resource not found")
		},
		func(core.Action) (bool, runtime.Object, error) {
			createCallCount++
			if createCallCount <= 2 {
				return true, nil, errors.New("Error " + strconv.Itoa(createCallCount))
			}
			return true, nil, nil
		},
	)

	err := InstallTypes(fakeClientset.ApiextensionsV1beta1().CustomResourceDefinitions())

	errStr := err.Error()
	if !strings.Contains(errStr, "Error 1") && !strings.Contains(errStr, "Error 2") {
		t.Errorf("Failed to receive correct errors during installation of Custom Resource concurrently, error received: %s", errStr)
	}
}

//make sure we don't poll on resource that was failed on install
func TestInstallTypesPolling(t *testing.T) {
	getCallCount := 0
	createCallCount := 0
	getCallArgs := []string{}

	fakeClientset := setup(
		func(action core.Action) (bool, runtime.Object, error) {
			getCallCount++
			if getCallCount > len(customResourceDefinitions) {
				getCallArgs = append(getCallArgs, action.(core.GetAction).GetName())
				return true, &crdv1beta1.CustomResourceDefinition{}, nil
			}

			return true, nil, errors.New("Resource not found")
		},
		func(action core.Action) (bool, runtime.Object, error) {
			createCallCount++
			name := action.(core.CreateAction).GetObject().(*crdv1beta1.CustomResourceDefinition).Name
			if name == serviceBrokerCRD.Name || name == serviceInstanceCRD.Name {
				return true, nil, errors.New("Error creatingCRD")
			}
			return true, nil, nil
		},
	)

	if err := InstallTypes(fakeClientset.ApiextensionsV1beta1().CustomResourceDefinitions()); err == nil {
		t.Fatal("InstallTypes was supposed to error but didn't")
	}

	for _, name := range getCallArgs {
		if name == serviceBrokerCRD.Name || name == serviceInstanceCRD.Name {
			t.Errorf("Failed to skip polling for resource that failed to install")
		}
	}
}
