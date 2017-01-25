package controller

import (
	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
)

func brokerForServiceClass(
	cl apiclient.APIClient,
	cls *servicecatalog.ServiceClass,
) (*servicecatalog.Broker, error) {

	// Get the broker for the serviceclass.
	broker, err := cl.Brokers().Get(cls.BrokerName)
	if err != nil {
		glog.Errorf("Error fetching broker for service: %s : %v", cls.BrokerName, err)
		return nil, err
	}
	return broker, nil
}

func serviceClassForInstance(
	cl apiclient.APIClient,
	instance *servicecatalog.Instance,
) (*servicecatalog.ServiceClass, error) {

	// Get the serviceclass for the instance.
	sc, err := cl.ServiceClasses().Get(instance.Spec.ServiceClassName)
	if err != nil {
		glog.Errorf("Failed to fetch service type %s : %v", instance.Spec.ServiceClassName, err)
		return nil, err
	}
	return sc, nil
}

func instanceForBinding(
	cl apiclient.APIClient,
	binding *servicecatalog.Binding,
) (*servicecatalog.Instance, error) {

	// Get instance information for service being bound to.
	instance, err := cl.Instances(binding.Spec.InstanceRef.Namespace).Get(binding.Spec.InstanceRef.Name)
	if err != nil {
		glog.Errorf("Service instance does not exist %v: %v", binding.Spec.InstanceRef, err)
		return nil, err
	}
	return instance, nil
}
