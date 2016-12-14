package v1alpha1

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBrokerStorage struct {
}

func NewServiceBrokerStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
