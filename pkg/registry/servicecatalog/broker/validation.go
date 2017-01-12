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

// this contains stubs of nothing until it is understood how it works

import (
	// commented out until we use the base validation utilities
	apivalidation "k8s.io/kubernetes/pkg/api/validation"
	// "k8s.io/kubernetes/pkg/api/validation/path"
	// utilvalidation "k8s.io/kubernetes/pkg/util/validation"
	"k8s.io/kubernetes/pkg/util/validation/field"

	discoveryapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// assuming a nil non-error is ok. not okay, should be empty struct `field.ErrorList{}`

// validateBroker makes sure a broker object is okay?
func validateBroker(broker *discoveryapi.Broker) field.ErrorList {
	allErrs := apivalidation.ValidateObjectMeta(&broker.ObjectMeta, false, /*namespace*/
		apivalidation.ValidateReplicationControllerName, // our custom name validator?
		field.NewPath("metadata"))
	return allErrs
}

// validateBrokerUpdate checks that when changing from an older broker to a newer broker is okay ?
func validateBrokerUpdate(new *discoveryapi.Broker, old *discoveryapi.Broker) field.ErrorList {
	return field.ErrorList{}
}
