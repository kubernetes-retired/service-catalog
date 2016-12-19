package binding

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBindingStorage struct {
}

// NewServiceBindingStorage creates a new rest.Storage responsible for accessing Binding
// resources
func NewServiceBindingStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
