package broker

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBrokerStorage struct {
}

// NewServiceBrokerStorage creates a new rest.Storage responsible for accessing Broker
// resources
func NewServiceBrokerStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
