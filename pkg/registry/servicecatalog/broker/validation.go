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

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// assuming a nil non-error is ok. not okay, should be empty struct `field.ErrorList{}`

// validateBroker makes sure a broker object is okay?
func validateBroker(broker *sc.Broker) field.ErrorList {
	allErrs := apivalidation.ValidateObjectMeta(&broker.ObjectMeta, false, /*namespace*/
		apivalidation.ValidateReplicationControllerName, // our custom name validator?
		field.NewPath("metadata"))
	allErrs = append(allErrs, validateBrokerSpec(&broker.Spec, field.NewPath("spec"))...)
	return allErrs
}

func validateBrokerSpec(spec *sc.BrokerSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	/* This is what is in the broker spec.

	URL string
	AuthUsername string
	AuthPassword string
	OSBGUID string
	*/

	if "" == spec.URL {
		allErrs = append(allErrs,
			field.Required(fldPath.Child("url"),
				"brokers must have a remote url to contact"))
	}
	if "" == spec.AuthUsername {
	}
	if "" == spec.AuthPassword {
	}
	if "" == spec.OSBGUID {
	}
	return allErrs
}

// validateBrokerUpdate checks that when changing from an older broker to a newer broker is okay ?
func validateBrokerUpdate(new *sc.Broker, old *sc.Broker) field.ErrorList {
	return field.ErrorList{}
}
