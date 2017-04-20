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

package servicecatalog

import (
	"fmt"

	testapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/testapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
)

// APIGroup returns an API group suitable for registering with the testapi.Groups map
func APIGroup() testapi.TestGroup {
	// OOPS: didn't register the right group version
	groupVersion, err := schema.ParseGroupVersion("servicecatalog.k8s.io/v1alpha1")
	if err != nil {
		panic(fmt.Sprintf("Error parsing groupversion: %v", err))
	}

	externalGroupVersion := schema.GroupVersion{
		Group:   GroupName,
		Version: api.Registry.GroupOrDie(GroupName).GroupVersion.Version,
	}

	return testapi.NewTestGroup(
		groupVersion,
		SchemeGroupVersion,
		api.Scheme.KnownTypes(SchemeGroupVersion),
		api.Scheme.KnownTypes(externalGroupVersion),
	)
}
