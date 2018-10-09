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
	"errors"
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

const (
	// FieldExternalClassName is the jsonpath to a class's external name.
	FieldExternalClassName = "spec.externalName"
)

// CreateClassFromOptions allows to specify how a new class will be created
type CreateClassFromOptions struct {
	Name      string
	Scope     Scope
	Namespace string
	From      string
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

	// GetSpec returns the spec.
	GetSpec() v1beta1.CommonServiceClassSpec

	// GetServiceBrokerName returns the name of the service
	// broker for the class.
	GetServiceBrokerName() string

	// GetStatusText returns the status of the class.
	GetStatusText() string
}

// RetrieveClasses lists all classes defined in the cluster.
func (sdk *SDK) RetrieveClasses(opts ScopeOptions) ([]Class, error) {
	var classes []Class
	if opts.Scope.Matches(ClusterScope) {
		csc, err := sdk.ServiceCatalog().ClusterServiceClasses().List(metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list cluster-scoped classes (%s)", err)
		}
		for _, c := range csc.Items {
			class := c
			classes = append(classes, &class)
		}
	}

	if opts.Scope.Matches(NamespaceScope) {
		sc, err := sdk.ServiceCatalog().ServiceClasses(opts.Namespace).List(metav1.ListOptions{})
		if err != nil {
			// Gracefully handle when the feature-flag for namespaced broker resources isn't enabled on the server.
			if apierrors.IsNotFound(err) {
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
func (sdk *SDK) RetrieveClassByName(name string, opts ScopeOptions) (Class, error) {
	var searchResults []Class

	lopts := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(FieldExternalClassName, name).String(),
	}

	if opts.Scope.Matches(ClusterScope) {
		csc, err := sdk.ServiceCatalog().ClusterServiceClasses().List(lopts)
		if err != nil {
			return nil, fmt.Errorf("unable to search classes by name (%s)", err)
		}

		for _, c := range csc.Items {
			class := c
			searchResults = append(searchResults, &class)
		}
	}

	if opts.Scope.Matches(NamespaceScope) {
		sc, err := sdk.ServiceCatalog().ServiceClasses(opts.Namespace).List(lopts)
		if err != nil {
			// Gracefully handle when the feature-flag for namespaced broker resources isn't enabled on the server.
			if apierrors.IsNotFound(err) {
				sc = &v1beta1.ServiceClassList{}
			} else {
				return nil, fmt.Errorf("unable to search classes by name (%s)", err)
			}
		}

		for _, c := range sc.Items {
			class := c
			searchResults = append(searchResults, &class)
		}
	}

	if len(searchResults) > 1 {
		return nil, fmt.Errorf("more than one matching class found for '%s' %d", name, len(searchResults))
	}

	if len(searchResults) == 0 {
		return nil, fmt.Errorf("class '%s' not found", name)
	}

	return searchResults[0], nil
}

// RetrieveClassByID gets a class by its UUID.
func (sdk *SDK) RetrieveClassByID(uuid string) (*v1beta1.ClusterServiceClass, error) {
	class, err := sdk.ServiceCatalog().ClusterServiceClasses().Get(uuid, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get class (%s)", err)
	}
	return class, nil
}

// RetrieveClassByPlan gets the class associated to a plan.
func (sdk *SDK) RetrieveClassByPlan(plan Plan) (*v1beta1.ClusterServiceClass, error) {
	// Retrieve the class as well because plans don't have the external class name
	class, err := sdk.ServiceCatalog().ClusterServiceClasses().Get(plan.GetClassID(), metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get class (%s)", err)
	}

	return class, nil
}

// CreateClassFrom returns new created class
func (sdk *SDK) CreateClassFrom(opts CreateClassFromOptions) (Class, error) {
	if opts.Scope == AllScope {
		return nil, errors.New("invalid scope: all")
	}

	fromClass, err := sdk.RetrieveClassByName(opts.From, ScopeOptions{Scope: opts.Scope, Namespace: opts.Namespace})
	if err != nil {
		return nil, err
	}

	if opts.Scope.Matches(ClusterScope) {
		csc := fromClass.(*v1beta1.ClusterServiceClass)
		return sdk.createClusterServiceClass(csc, opts.Name)
	}

	sc := fromClass.(*v1beta1.ServiceClass)
	return sdk.createServiceClass(sc, opts.Name, opts.Namespace)
}

func (sdk *SDK) createClusterServiceClass(from *v1beta1.ClusterServiceClass, name string) (*v1beta1.ClusterServiceClass, error) {
	var class = &v1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       from.Spec,
	}
	class.Spec.ExternalName = name // this is the name displayed by svcat, not the k8s name

	created, err := sdk.ServiceCatalog().ClusterServiceClasses().Create(class)
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster service class (%s)", err)
	}

	return created, nil
}

func (sdk *SDK) createServiceClass(from *v1beta1.ServiceClass, name, namespace string) (*v1beta1.ServiceClass, error) {
	var class = &v1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       from.Spec,
	}
	class.Spec.ExternalName = name // this is the name displayed by svcat, not the k8s name

	created, err := sdk.ServiceCatalog().ServiceClasses(namespace).Create(class)
	if err != nil {
		return nil, fmt.Errorf("unable to create service class (%s)", err)
	}

	return created, nil
}
