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

package servicebroker

// this was copied from where else and edited to fit our objects

import (
	"context"

	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/golang/glog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

// NewScopeStrategy returns a new NamespaceScopedStrategy for brokers
func NewScopeStrategy() rest.NamespaceScopedStrategy {
	return serviceBrokerRESTStrategies
}

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy,
// NamespaceScopedStrategy
type serviceBrokerRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	names.NameGenerator // GenerateName method for CreateStrategy
}

// implements interface RESTUpdateStrategy
type serviceBrokerStatusRESTStrategy struct {
	serviceBrokerRESTStrategy
}

var (
	serviceBrokerRESTStrategies = serviceBrokerRESTStrategy{
		// embeds to pull in existing code behavior from upstream

		// this has an interesting NOTE on it. Not sure if it applies to us.
		ObjectTyper: api.Scheme,
		// use the generator from upstream k8s, or implement method
		// `GenerateName(base string) string`
		NameGenerator: names.SimpleNameGenerator,
	}
	_ rest.RESTCreateStrategy = serviceBrokerRESTStrategies
	_ rest.RESTUpdateStrategy = serviceBrokerRESTStrategies
	_ rest.RESTDeleteStrategy = serviceBrokerRESTStrategies

	serviceBrokerStatusUpdateStrategy = serviceBrokerStatusRESTStrategy{
		serviceBrokerRESTStrategies,
	}
	_ rest.RESTUpdateStrategy = serviceBrokerStatusUpdateStrategy
)

// Canonicalize does not transform a broker.
func (serviceBrokerRESTStrategy) Canonicalize(obj runtime.Object) {
	_, ok := obj.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to create")
	}
}

// NamespaceScoped returns true as servicebrokers are scoped to a namespace.
func (serviceBrokerRESTStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate receives a the incoming ServiceBroker and clears it's
// Status. Status is not a user settable field.
func (serviceBrokerRESTStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	broker, ok := obj.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to create")
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

func (serviceBrokerRESTStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return scv.ValidateServiceBroker(obj.(*sc.ServiceBroker))
}

func (serviceBrokerRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (serviceBrokerRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (serviceBrokerRESTStrategy) PrepareForUpdate(ctx context.Context, new, old runtime.Object) {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to update to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to update from")
	}

	newServiceBroker.Status = oldServiceBroker.Status

	// Ignore the RelistRequests field when it is the default value
	if newServiceBroker.Spec.RelistRequests == 0 {
		newServiceBroker.Spec.RelistRequests = oldServiceBroker.Spec.RelistRequests
	}

	// Spec updates bump the generation so that we can distinguish between
	// spec changes and other changes to the object.
	if !apiequality.Semantic.DeepEqual(oldServiceBroker.Spec, newServiceBroker.Spec) {
		newServiceBroker.Generation = oldServiceBroker.Generation + 1
	}
}

func (serviceBrokerRESTStrategy) ValidateUpdate(ctx context.Context, new, old runtime.Object) field.ErrorList {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to validate to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to validate from")
	}

	return scv.ValidateServiceBrokerUpdate(newServiceBroker, oldServiceBroker)
}

func (serviceBrokerStatusRESTStrategy) PrepareForUpdate(ctx context.Context, new, old runtime.Object) {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to update to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to update from")
	}
	// status changes are not allowed to update spec
	newServiceBroker.Spec = oldServiceBroker.Spec
}

func (serviceBrokerStatusRESTStrategy) ValidateUpdate(ctx context.Context, new, old runtime.Object) field.ErrorList {
	newServiceBroker, ok := new.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to validate to")
	}
	oldServiceBroker, ok := old.(*sc.ServiceBroker)
	if !ok {
		glog.Fatal("received a non-servicebroker object to validate from")
	}

	return scv.ValidateServiceBrokerStatusUpdate(newServiceBroker, oldServiceBroker)
}
