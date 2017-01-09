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
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

type brokerCreateStrategy struct {
	runtime.ObjectTyper // ObjectKinds method for CreateStrategy
	kapi.NameGenerator  // GenerateName method for CreateStrategy
}

// Strategy implements the call backs for the generic store
var createStrategy = brokerCreateStrategy{
	// this has an interesting NOTE on it. Not sure if it applies to us.
	kapi.Scheme,
	kapi.SimpleNameGenerator,
}

func (brokerCreateStrategy) Canonicalize(obj runtime.Object) {}

// Are brokers namespace scoped?
func (brokerCreateStrategy) NamespaceScoped() bool {
	return false
}

func (brokerCreateStrategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
	_ = obj.(*servicecatalog.Broker)
}

func (brokerCreateStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	return validateBroker(obj.(*servicecatalog.Broker))
}

type brokerDeleteStrategy struct {
	runtime.ObjectTyper // ObjectKinds method for DeleteStrategy
}

// Strategy implements the call backs for the generic store
var deleteStrategy = brokerDeleteStrategy{
	// this has an interesting NOTE on it. Not sure if it applies to us.
	kapi.Scheme,
}

type brokerUpdateStrategy struct {
	runtime.ObjectTyper // ObjectKinds method for DeleteStrategy
}

// Strategy implements the call backs for the generic store
var updateStrategy = brokerUpdateStrategy{
	// this has an interesting NOTE on it. Not sure if it applies to us.
	kapi.Scheme,
}

func (brokerUpdateStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (brokerUpdateStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (brokerUpdateStrategy) Canonicalize(obj runtime.Object) {}

// Are brokers namespace scoped?
func (brokerUpdateStrategy) NamespaceScoped() bool {
	return false
}

func (brokerUpdateStrategy) PrepareForUpdate(ctx kapi.Context, new, old runtime.Object) {
	newBroker := new.(*servicecatalog.Broker)
	oldBroker := old.(*servicecatalog.Broker)
	newBroker.Status = oldBroker.Status
}

func (brokerUpdateStrategy) ValidateUpdate(ctx kapi.Context, new, old runtime.Object) field.ErrorList {
	return validateBrokerUpdate(new.(*servicecatalog.Broker), old.(*servicecatalog.Broker))
}
