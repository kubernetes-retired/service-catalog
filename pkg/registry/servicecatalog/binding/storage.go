package binding

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBindingStorage struct {
}

// NewStorage creates a new rest.Storage responsible for accessing Binding
// resources
func NewStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
