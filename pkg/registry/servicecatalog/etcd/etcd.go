package etcd

import (
	"k8s.io/kubernetes/pkg/registry/generic"
	genericregistry "k8s.io/kubernetes/pkg/registry/generic/registry"
	"k8s.io/kubernetes/pkg/runtime"

	apis "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	registry "github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog"
)

// REST implements a RESTStorage for API services against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against API services.
//
// this seems like it needs to be specifically written for each resource?
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:     func() runtime.Object { return &apis.Broker{} },
		NewListFunc: func() runtime.Object { return &apis.BrokerList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*apis.Broker).Name, nil
		},
		PredicateFunc:     registry.MatchAPIService,
		QualifiedResource: apis.Resource("servicecatalog"),

		CreateStrategy: registry.Strategy,
		UpdateStrategy: registry.Strategy,
		DeleteStrategy: registry.Strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: registry.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	return &REST{store}
}
