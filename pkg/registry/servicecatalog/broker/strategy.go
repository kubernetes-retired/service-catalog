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

package broker

// this was copied from where else and edited to fit our objects

import (
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/pkg/api"

	"github.com/golang/glog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

// NewScopeStrategy returns a new NamespaceScopedStrategy for brokers
func NewScopeStrategy() rest.NamespaceScopedStrategy {
	return brokerRESTStrategies
}

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy,
// NamespaceScopedStrategy
type brokerRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	names.NameGenerator // GenerateName method for CreateStrategy
}

// implements interface RESTUpdateStrategy
type brokerStatusRESTStrategy struct {
	brokerRESTStrategy
}

// implements interface RESTUpdateStrategy
type brokerRelistRESTStrategy struct {
	brokerRESTStrategy
}

var (
	brokerRESTStrategies = brokerRESTStrategy{
		// embeds to pull in existing code behavior from upstream

		// this has an interesting NOTE on it. Not sure if it applies to us.
		ObjectTyper: api.Scheme,
		// use the generator from upstream k8s, or implement method
		// `GenerateName(base string) string`
		NameGenerator: names.SimpleNameGenerator,
	}
	_ rest.RESTCreateStrategy = brokerRESTStrategies
	_ rest.RESTUpdateStrategy = brokerRESTStrategies
	_ rest.RESTDeleteStrategy = brokerRESTStrategies

	brokerStatusUpdateStrategy = brokerStatusRESTStrategy{
		brokerRESTStrategies,
	}

	brokerRelistUpdateStrategy = brokerRelistRESTStrategy{
		brokerRESTStrategies,
	}
	_ rest.RESTUpdateStrategy = brokerRelistUpdateStrategy
)

// Canonicalize does not transform a broker.
func (brokerRESTStrategy) Canonicalize(obj runtime.Object) {
	_, ok := obj.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to create")
	}
}

// NamespaceScoped returns false as brokers are not scoped to a namespace.
func (brokerRESTStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate receives a the incoming ServiceBroker and clears it's
// Status. Status is not a user settable field.
func (brokerRESTStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	broker, ok := obj.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to create")
	}
	// Is there anything to pull out of the context `ctx`?

	// Creating a brand new object, thus it must have no
	// status. We can't fail here if they passed a status in, so
	// we just wipe it clean.
	broker.Status = sc.ServiceBrokerStatus{}
	// Fill in the first entry set to "creating"?
	broker.Status.Conditions = []sc.ServiceBrokerCondition{}
	broker.Finalizers = []string{sc.FinalizerServiceCatalog}
	broker.Generation = 1
}

func (brokerRESTStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return scv.ValidateServiceBroker(obj.(*sc.ServiceBroker))
}

func (brokerRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (brokerRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (brokerRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update from")
	}

	newServiceBroker.Status = oldServiceBroker.Status

	// Spec updates bump the generation so that we can distinguish between
	// spec changes and other changes to the object.
	if !apiequality.Semantic.DeepEqual(oldServiceBroker.Spec, newServiceBroker.Spec) {
		newServiceBroker.Generation = oldServiceBroker.Generation + 1
	}
}

func (brokerRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate from")
	}

	return scv.ValidateServiceBrokerUpdate(newServiceBroker, oldServiceBroker)
}

func (brokerStatusRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update from")
	}
	// status changes are not allowed to update spec
	newServiceBroker.Spec = oldServiceBroker.Spec
}

func (brokerStatusRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate from")
	}

	return scv.ValidateServiceBrokerStatusUpdate(newServiceBroker, oldServiceBroker)
}

func (r *RelistREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}

func (brokerRelistRESTStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to update from")
	}

	newServiceBroker.Spec = oldServiceBroker.Spec
	newServiceBroker.Status = oldServiceBroker.Status
	newServiceBroker.Spec.RelistRequests = oldServiceBroker.Spec.RelistRequests + 1
}

func (brokerRelistRESTStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-broker object to validate from")
	}

	return scv.ValidateServiceBrokerRelistUpdate(newServiceBroker, oldServiceBroker)
}
