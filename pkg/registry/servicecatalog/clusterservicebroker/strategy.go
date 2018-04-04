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

package clusterservicebroker

// this was copied from where else and edited to fit our objects

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/golang/glog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

// NewScopeStrategy returns a new NamespaceScopedStrategy for brokers
func NewScopeStrategy() rest.NamespaceScopedStrategy {
	return clusterServiceBrokerRESTStrategies
}

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy,
// NamespaceScopedStrategy
type clusterServiceBrokerRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	names.NameGenerator // GenerateName method for CreateStrategy
}

// implements interface RESTUpdateStrategy
type clusterServiceBrokerStatusRESTStrategy struct {
	clusterServiceBrokerRESTStrategy
}

var (
	clusterServiceBrokerRESTStrategies = clusterServiceBrokerRESTStrategy{
		// embeds to pull in existing code behavior from upstream

		// this has an interesting NOTE on it. Not sure if it applies to us.
		ObjectTyper: api.Scheme,
		// use the generator from upstream k8s, or implement method
		// `GenerateName(base string) string`
		NameGenerator: names.SimpleNameGenerator,
	}
	_ rest.RESTCreateStrategy = clusterServiceBrokerRESTStrategies
	_ rest.RESTUpdateStrategy = clusterServiceBrokerRESTStrategies
	_ rest.RESTDeleteStrategy = clusterServiceBrokerRESTStrategies

	clusterServiceBrokerStatusUpdateStrategy = clusterServiceBrokerStatusRESTStrategy{
		clusterServiceBrokerRESTStrategies,
	}
	_ rest.RESTUpdateStrategy = clusterServiceBrokerStatusUpdateStrategy
)

// Canonicalize does not transform a broker.
func (clusterServiceBrokerRESTStrategy) Canonicalize(obj runtime.Object) {
	_, ok := obj.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to create")
	}
}

// NamespaceScoped returns false as clusterservicebrokers are not scoped to a
// namespace.
func (clusterServiceBrokerRESTStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate receives a the incoming ClusterServiceBroker and clears it's
// Status. Status is not a user settable field.
func (clusterServiceBrokerRESTStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	broker, ok := obj.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to create")
	}
	// Is there anything to pull out of the context `ctx`?

	// Creating a brand new object, thus it must have no
	// status. We can't fail here if they passed a status in, so
	// we just wipe it clean.
	broker.Status = sc.ClusterServiceBrokerStatus{}
	// Fill in the first entry set to "creating"?
	broker.Status.Conditions = []sc.ServiceBrokerCondition{}
	broker.Finalizers = []string{sc.FinalizerServiceCatalog}
	broker.Generation = 1
}

func (clusterServiceBrokerRESTStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return scv.ValidateClusterServiceBroker(obj.(*sc.ClusterServiceBroker))
}

func (clusterServiceBrokerRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (clusterServiceBrokerRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (clusterServiceBrokerRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newClusterServiceBroker, ok := new.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to update to")
	}
	oldClusterServiceBroker, ok := old.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to update from")
	}

	newClusterServiceBroker.Status = oldClusterServiceBroker.Status

	// Ignore the RelistRequests field when it is the default value
	if newClusterServiceBroker.Spec.RelistRequests == 0 {
		newClusterServiceBroker.Spec.RelistRequests = oldClusterServiceBroker.Spec.RelistRequests
	}

	// Spec updates bump the generation so that we can distinguish between
	// spec changes and other changes to the object.
	if !apiequality.Semantic.DeepEqual(oldClusterServiceBroker.Spec, newClusterServiceBroker.Spec) {
		newClusterServiceBroker.Generation = oldClusterServiceBroker.Generation + 1
	}
}

func (clusterServiceBrokerRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newClusterServiceBroker, ok := new.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to validate to")
	}
	oldClusterServiceBroker, ok := old.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to validate from")
	}

	return scv.ValidateClusterServiceBrokerUpdate(newClusterServiceBroker, oldClusterServiceBroker)
}

func (clusterServiceBrokerStatusRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newClusterServiceBroker, ok := new.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to update to")
	}
	oldClusterServiceBroker, ok := old.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to update from")
	}
	// status changes are not allowed to update spec
	newClusterServiceBroker.Spec = oldClusterServiceBroker.Spec
}

func (clusterServiceBrokerStatusRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newClusterServiceBroker, ok := new.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to validate to")
	}
	oldClusterServiceBroker, ok := old.(*sc.ClusterServiceBroker)
	if !ok {
		glog.Fatal("received a non-clusterservicebroker object to validate from")
	}

	return scv.ValidateClusterServiceBrokerStatusUpdate(newClusterServiceBroker, oldClusterServiceBroker)
}
