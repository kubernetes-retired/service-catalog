package servicecatalog

// this was copied from where else and edited to fit our objects

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/registry/generic"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/storage"
	"k8s.io/kubernetes/pkg/util/validation/field"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/validation"
)

type apiServerStrategy struct {
	runtime.ObjectTyper
	kapi.NameGenerator
}

// Strategy implements the call backs for the generic store
var Strategy = apiServerStrategy{kapi.Scheme, kapi.SimpleNameGenerator}

func (apiServerStrategy) NamespaceScoped() bool {
	return false
}

func (apiServerStrategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
	_ = obj.(*servicecatalog.Broker)
}

func (apiServerStrategy) PrepareForUpdate(ctx kapi.Context, new, old runtime.Object) {
	newAPIService := new.(*servicecatalog.Broker)
	oldAPIService := old.(*servicecatalog.Broker)
	newAPIService.Status = oldAPIService.Status
}

func (apiServerStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	return validation.ValidateBroker(obj.(*servicecatalog.Broker))
}

func (apiServerStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (apiServerStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (apiServerStrategy) Canonicalize(obj runtime.Object) {
}

func (apiServerStrategy) ValidateUpdate(ctx kapi.Context, new, old runtime.Object) field.ErrorList {
	return validation.ValidateBrokerUpdate(new.(*servicecatalog.Broker), old.(*servicecatalog.Broker))
}

// GetAttrs returns attrs.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	apiserver, ok := obj.(*servicecatalog.Broker)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a broker, COME ON")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), aPIServiceToSelectableFields(apiserver), nil
}

// MatchAPIService is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchAPIService(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// APIServiceToSelectableFields returns a field set that represents the object.
// no reason for this to be exported that I can see.
func aPIServiceToSelectableFields(obj *servicecatalog.Broker) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}
