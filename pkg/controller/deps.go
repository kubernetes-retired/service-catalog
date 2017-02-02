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

package controller

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
)

func instanceHasBindings(cl apiclient.APIClient, inst *servicecatalog.Instance) (bool, error) {
	nsList, err := cl.Namespaces()
	if err != nil {
		return false, err
	}
	for _, ns := range nsList {
		bindList, err := cl.Bindings(ns).List()
		if err != nil {
			return false, err
		}
		for _, bind := range bindList {
			ref := bind.Spec.InstanceRef
			if ref.Name == inst.Name && ref.Namespace == inst.Namespace {
				return true, nil
			}
		}
	}
	return false, err
}

func serviceClassHasInstances(cl apiclient.APIClient, cls *servicecatalog.ServiceClass) (bool, error) {
	nsList, err := cl.Namespaces()
	if err != nil {
		return false, err
	}
	for _, ns := range nsList {
		instList, err := cl.Instances(ns).List()
		if err != nil {
			return false, err
		}
		for _, inst := range instList {
			if cls.Name == inst.Spec.ServiceClassName {
				return true, nil
			}
		}
	}
	return false, nil
}

func brokerHasInstances(cl apiclient.APIClient, broker *servicecatalog.Broker) (bool, error) {
	svcClassList, err := cl.ServiceClasses().List()
	if err != nil {
		return false, err
	}
	for _, cls := range svcClassList {
		// if this service class isn't for the given broker, then skip
		if cls.BrokerName != broker.Name {
			continue
		}
		hasInst, err := serviceClassHasInstances(cl, cls)
		if err != nil {
			return false, err
		}
		if hasInst {
			return true, nil
		}
	}
	return false, nil
}
