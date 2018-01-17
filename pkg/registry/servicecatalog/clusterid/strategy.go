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

package clusterid

import (
	"github.com/golang/glog"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy,
// NamespaceScopedStrategy
type clusterIDRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
}

var (
	clusterIDStrategies = clusterIDRESTStrategy{
		ObjectTyper: api.Scheme,
	}

	_ rest.RESTCreateStrategy = clusterIDStrategies
	_ rest.RESTUpdateStrategy = clusterIDStrategies
	_ rest.RESTDeleteStrategy = clusterIDStrategies
)

func (clusterIDRESTStrategy) GenerateName(base string) string {
	return "cluster-id"
}

// NewScopeStrategy returns a new NamespaceScopedStrategy for ClusterIDs
func NewScopeStrategy() rest.NamespaceScopedStrategy {
	return clusterIDStrategies
}

// Canonicalize does not transform a ClusterID.
func (clusterIDRESTStrategy) Canonicalize(obj runtime.Object) {
	_, ok := obj.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to create")
	}
}

// NamespaceScoped returns false as ClusterIDs are not scoped to a namespace.
func (clusterIDRESTStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate receives a the incoming ClusterServiceClusterID and clears it's
// Status. Status is not a user settable field.
func (clusterIDRESTStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	_, ok := obj.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to create")
	}
}

func (clusterIDRESTStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return scv.ValidateClusterID(obj.(*sc.ClusterID))
}

func (clusterIDRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (clusterIDRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (clusterIDRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newClusterID, ok := new.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to update to")
	}
	oldClusterID, ok := old.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to update from")
	}

	if !apiequality.Semantic.DeepEqual(oldClusterID, newClusterID) {
	}
}

func (clusterIDRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newClusterID, ok := new.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to validate to")
	}
	oldClusterID, ok := old.(*sc.ClusterID)
	if !ok {
		glog.Fatal("received a non-ClusterID object to validate from")
	}

	return scv.ValidateClusterIDUpdate(newClusterID, oldClusterID)
}
