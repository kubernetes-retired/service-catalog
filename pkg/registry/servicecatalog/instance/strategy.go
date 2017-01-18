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

package instance

// this was copied from where else and edited to fit our objects

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/golang/glog"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scv "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

// implements interfaces RESTCreateStrategy, RESTUpdateStrategy, RESTDeleteStrategy
type instanceRESTStrategy struct {
	runtime.ObjectTyper // inherit ObjectKinds method
	kapi.NameGenerator  // GenerateName method for CreateStrategy
}

var (
	instanceRESTStrategies = instanceRESTStrategy{
		// embeds to pull in existing code behavior from upstream

		ObjectTyper: kapi.Scheme,
		// use the generator from upstream k8s, or implement method
		// `GenerateName(base string) string`
		NameGenerator: kapi.SimpleNameGenerator,
	}
	_ rest.RESTCreateStrategy = instanceRESTStrategies
	_ rest.RESTUpdateStrategy = instanceRESTStrategies
	_ rest.RESTDeleteStrategy = instanceRESTStrategies
)

// Canonicalize does not transform a instance.
func (instanceRESTStrategy) Canonicalize(obj runtime.Object) {
	_, ok := obj.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to create")
	}
}

// NamespaceScoped returns false as instances are not scoped to a namespace.
func (instanceRESTStrategy) NamespaceScoped() bool {
	return false
}

// PrepareForCreate receives a the incoming Instance and clears it's
// Status. Status is not a user settable field.
func (instanceRESTStrategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
	instance, ok := obj.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to create")
	}

	// Creating a brand new object, thus it must have no
	// status. We can't fail here if they passed a status in, so
	// we just wipe it clean.
	instance.Status = sc.InstanceStatus{}
	// Fill in the first entry set to "creating"?
	instance.Status.Conditions = []sc.InstanceCondition{}
}

func (instanceRESTStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	return scv.ValidateInstance(obj.(*sc.Instance))
}

func (instanceRESTStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (instanceRESTStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (instanceRESTStrategy) PrepareForUpdate(ctx kapi.Context, new, old runtime.Object) {
	newInstance, ok := new.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to update to")
	}
	oldInstance := old.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to update from")
	}
	newInstance.Spec = oldInstance.Spec
	newInstance.Status = oldInstance.Status
}

func (instanceRESTStrategy) ValidateUpdate(ctx kapi.Context, new, old runtime.Object) field.ErrorList {
	newInstance, ok := new.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to validate to")
	}
	oldInstance := old.(*sc.Instance)
	if !ok {
		glog.Warning("received a non-instance object to validate from")
	}

	return scv.ValidateInstanceUpdate(newInstance, oldInstance)
}
