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
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/golang/glog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy
type brokerRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	kapi.NameGenerator  // GenerateName method for CreateStrategy
}

var (
	brokerRESTStrategies = brokerRESTStrategy{
		// embeds to pull in existing code behavior from upstream

		// this has an interesting NOTE on it. Not sure if it applies to us.
		ObjectTyper: kapi.Scheme,
		// use the generator from upstream k8s, or implement method
		// `GenerateName(base string) string`
		NameGenerator: kapi.SimpleNameGenerator,
	}
	_ rest.RESTCreateStrategy = brokerRESTStrategies
	_ rest.RESTUpdateStrategy = brokerRESTStrategies
	_ rest.RESTDeleteStrategy = brokerRESTStrategies
)

// Canonicalize does not transform a broker.
func (brokerRESTStrategy) Canonicalize(obj runtime.Object) {
	broker, ok := obj.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to create")
	}
}

// NamespaceScoped returns false as brokers are not scoped to a namespace.
func (brokerRESTStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate receives a the incoming Broker and clears it's
// Status. Status is not a user settable field.
func (brokerRESTStrategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
	broker, ok := obj.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to create")
	}
	// Is there anything to pull out of the context `ctx`?

	// Creating a brand new object, thus it must have no
	// status. We can't fail here if they passed a status in, so
	// we just wipe it clean.
	broker.Status = sc.BrokerStatus{}
	// Fill in the first entry set to "creating"?
	broker.Status.Conditions = []sc.BrokerCondition{}
}

func (brokerRESTStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	return validateBroker(obj.(*sc.Broker))
}

func (brokerRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (brokerRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (brokerRESTStrategy) PrepareForUpdate(ctx kapi.Context, new, old runtime.Object) {
	newBroker, ok := new.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to update to")
	}
	oldBroker := old.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to update from")
	}

	newBroker.Status = oldBroker.Status
}

func (brokerRESTStrategy) ValidateUpdate(ctx kapi.Context, new, old runtime.Object) field.ErrorList {
	newBroker, ok := new.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to validate to")
	}
	oldBroker := old.(*sc.Broker)
	if !ok {
		glog.Warning("received a non-broker object to validate from")
	}

	return validateBrokerUpdate(newBroker, oldBroker)
}
