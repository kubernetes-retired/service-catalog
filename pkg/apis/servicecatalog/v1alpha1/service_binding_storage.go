package v1alpha1

import (
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/registry/generic/registry"
)

type serviceBindingStorage struct {
}

func NewServiceBindingStorage() rest.Storage {
	return &registry.Store{
	// TODO: implement
	}
}
