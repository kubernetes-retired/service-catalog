package broker

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBrokerStorage struct {
}

// NewStorage creates a new rest.Storage responsible for accessing Broker
// resources
func NewStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
