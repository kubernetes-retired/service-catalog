/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	// FieldExternalClassName is the jsonpath to a class's external name.
	FieldExternalClassName = "spec.externalName"
)

// CreateClassFromOptions allows to specify how a new class will be created
type CreateClassFromOptions struct {
	Name      string
	Namespace string
	From      string
	Scope     Scope
}

// Class provides a unifying layer of cluster and namespace scoped class resources.
type Class interface {

	// GetName returns the class's name.
	GetName() string

	// GetNamespace returns the class's namespace, or "" if it's cluster-scoped.
	GetNamespace() string

	// GetExternalName returns the class's external name.
	GetExternalName() string

	// GetDescription returns the class description.
	GetDescription() string
}

// RetrieveClasses lists all classes defined in the cluster.
func (sdk *SDK) RetrieveClasses(opts ScopeOptions) ([]Class, error) {
	var classes []Class
	if opts.Scope.Matches(ClusterScope) {
		csc, err := sdk.ServiceCatalog().ClusterServiceClasses().List(v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list cluster-scoped classes (%s)", err)
		}
		for _, c := range csc.Items {
			class := c
			classes = append(classes, &class)
		}
	}

	if opts.Scope.Matches(NamespaceScope) {
		sc, err := sdk.ServiceCatalog().ServiceClasses(opts.Namespace).List(v1.ListOptions{})
		if err != nil {
			// Gracefully handle when the feature-flag for namespaced broker resources isn't enabled on the server.
			if errors.IsNotFound(err) {
				return classes, nil
			}
			return nil, fmt.Errorf("unable to list classes in %q (%s)", opts.Namespace, err)
		}
		for _, c := range sc.Items {
			class := c
			classes = append(classes, &class)
		}
	}

	return classes, nil
}

// RetrieveClassByName gets a class by its external name.
func (sdk *SDK) RetrieveClassByName(name string) (*v1beta1.ClusterServiceClass, error) {
	opts := v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(FieldExternalClassName, name).String(),
	}
	searchResults, err := sdk.ServiceCatalog().ClusterServiceClasses().List(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to search classes by name (%s)", err)
	}
	if len(searchResults.Items) == 0 {
		return nil, fmt.Errorf("class '%s' not found", name)
	}
	if len(searchResults.Items) > 1 {
		return nil, fmt.Errorf("more than one matching class found for '%s'", name)
	}
	return &searchResults.Items[0], nil
}

// RetrieveClassByID gets a class by its UUID.
func (sdk *SDK) RetrieveClassByID(uuid string) (*v1beta1.ClusterServiceClass, error) {
	class, err := sdk.ServiceCatalog().ClusterServiceClasses().Get(uuid, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get class (%s)", err)
	}
	return class, nil
}

// RetrieveClassByPlan gets the class associated to a plan.
func (sdk *SDK) RetrieveClassByPlan(plan *v1beta1.ClusterServicePlan,
) (*v1beta1.ClusterServiceClass, error) {
	// Retrieve the class as well because plans don't have the external class name
	class, err := sdk.ServiceCatalog().ClusterServiceClasses().Get(plan.Spec.ClusterServiceClassRef.Name, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get class (%s)", err)
	}

	return class, nil
}

// CreateClassFrom returns new created class
func (sdk *SDK) CreateClassFrom(opts CreateClassFromOptions) (Class, error) {
	fromClass, err := sdk.RetrieveClassByName(opts.From)
	if err != nil {
		return nil, err
	}

	if opts.Scope.Matches(ClusterScope) {
		fromClass.Name = opts.Name
		return sdk.createClusterServiceClass(fromClass)
	}

	fromServiceClass := &v1beta1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{Name: opts.Name, Namespace: opts.Namespace},
		Spec: v1beta1.ServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalName:       fromClass.Spec.ExternalName,
				ExternalID:         fromClass.Spec.ExternalID,
				Description:        fromClass.Spec.Description,
				Bindable:           fromClass.Spec.Bindable,
				BindingRetrievable: fromClass.Spec.BindingRetrievable,
				PlanUpdatable:      fromClass.Spec.PlanUpdatable,
				ExternalMetadata:   fromClass.Spec.ExternalMetadata,
				Tags:               fromClass.Spec.Tags,
				Requires:           fromClass.Spec.Requires,
			},
		},
		Status: v1beta1.ServiceClassStatus{
			CommonServiceClassStatus: v1beta1.CommonServiceClassStatus{
				RemovedFromBrokerCatalog: fromClass.Status.RemovedFromBrokerCatalog,
			},
		},
	}
	return sdk.createServiceClass(fromServiceClass)
}

func (sdk *SDK) createClusterServiceClass(from *v1beta1.ClusterServiceClass) (*v1beta1.ClusterServiceClass, error) {
	created, err := sdk.ServiceCatalog().ClusterServiceClasses().Create(from)
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster service class (%s)", err)
	}

	return created, nil
}

func (sdk *SDK) createServiceClass(class *v1beta1.ServiceClass) (*v1beta1.ServiceClass, error) {
	created, err := sdk.ServiceCatalog().ServiceClasses(class.GetNamespace()).Create(class)
	if err != nil {
		return nil, fmt.Errorf("unable to create service class (%s)", err)
	}

	return created, nil
}
