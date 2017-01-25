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
		hasInst, err := serviceClassHasInstances(cl, cls)
		if err != nil {
			return false, err
		}
		if cls.BrokerName == broker.Name && hasInst {
			return true, nil
		}
	}
	return false, nil
}
