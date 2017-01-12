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

/* Begin Create Definition */

type brokerCreateStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	kapi.NameGenerator  // GenerateName method for CreateStrategy
}

// implements RESTCreateStrategy interface
var createStrategy = brokerCreateStrategy{
	// embeds to pull in existing code behavior from upstream

	// this has an interesting NOTE on it. Not sure if it applies to us.
	ObjectTyper: kapi.Scheme,
	// use the generator from upstream k8s, or implement method
	// `GenerateName(base string) string`
	NameGenerator: kapi.SimpleNameGenerator,
}

// Canonicalize is called after validate, what happens if it creates an object to persist that does not pass validate? (I think the answer is "don't do that")
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

/* End Create Definition */

/* Begin Delete Definition */
// implements RESTDeleteStrategy interface
type brokerDeleteStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
}

// Strategy implements
var deleteStrategy = brokerDeleteStrategy{
	// this has an interesting NOTE on it. Not sure if it applies to us.
	ObjectTyper: kapi.Scheme,
}

type brokerUpdateStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
}

// There is no implementation code for Delete.
/* End Delete Definition */

/* Begin Update Definition */

// implements RESTUpdateStrategy interface
var updateStrategy = brokerUpdateStrategy{
	// this has an interesting NOTE on it. Not sure if it applies to us.
	ObjectTyper: kapi.Scheme,
}

func (brokerUpdateStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (brokerUpdateStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (brokerUpdateStrategy) Canonicalize(obj runtime.Object) {}

// Are brokers namespace scoped? Is the namespace concerned about an
// external namespace or some storage namespace?
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

/* End Update Definition */
