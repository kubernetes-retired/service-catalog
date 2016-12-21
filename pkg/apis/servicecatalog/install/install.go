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

// Package install registers the service-catalog API group
package install

import (
	"k8s.io/kubernetes/pkg/apimachinery/announced"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

func init() {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:              servicecatalog.GroupName,
			VersionPreferenceOrder: []string{v1alpha1.SchemeGroupVersion.Version},
			ImportPrefix:           "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog",
			// TODO: what does this do?
			RootScopedKinds: nil, // nil is allowed
			// TODO: Do we have 'internal objects'? What is an 'internal object'?
			// mhb: ? broker/catalog/service/instance are our 'internal objects' ?
			AddInternalObjectsToScheme: servicecatalog.AddToScheme, // nil if there are no 'internal objects'
		},
		// TODO what does this do? Is it necessary?
		announced.VersionToSchemeFunc{
			v1alpha1.SchemeGroupVersion.Version: v1alpha1.AddToScheme,
		},
	).Announce().RegisterAndEnable(); err != nil {
		panic(err)
	}
}
