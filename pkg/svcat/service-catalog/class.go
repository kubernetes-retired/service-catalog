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
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MultipleClassesFoundError is the error returned when we find a clusterserviceclass
	// and a serviceclass with the same name
	MultipleClassesFoundError = "More than one class found"
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

	// IsClusterServiceCLass returns true if the class is a ClusterServiceClass
	IsClusterServiceClass() bool
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
		LabelSelector: labels.SelectorFromSet(labels.Set{
			v1beta1.GroupName + "/" + v1beta1.FilterSpecExternalName: name,
		}).String(),
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
		if opts.Scope.Matches(ClusterScope) {
			return nil, fmt.Errorf("class '%s' not found in cluster scope", name)
		} else if opts.Scope.Matches(NamespaceScope) {
			if opts.Namespace == "" {
				return nil, fmt.Errorf("class '%s' not found in any namespace", name)
			}
			return nil, fmt.Errorf("class '%s' not found in namespace %s", name, opts.Namespace)
		}
		return nil, fmt.Errorf("class '%s' not found", name)
	}

	return searchResults[0], nil
}

// RetrieveClassByID gets a class by its Kubernetes name.
func (sdk *SDK) RetrieveClassByID(kubeName string, opts ScopeOptions) (Class, error) {
	var csc *v1beta1.ClusterServiceClass
	var sc *v1beta1.ServiceClass
	var err error
	if opts.Scope.Matches(ClusterScope) {
		csc, err = sdk.ServiceCatalog().ClusterServiceClasses().Get(kubeName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			csc = nil
		}
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to get class (%s)", err)
		}
	}
	if opts.Scope.Matches(NamespaceScope) {
		sc, err = sdk.ServiceCatalog().ServiceClasses(opts.Namespace).Get(kubeName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			sc = nil
		}
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("unable to get class (%s)", err)
		}
	}

	switch {
	case csc != nil && sc != nil:
		return nil, fmt.Errorf(MultipleClassesFoundError+" for '%s'", kubeName)
	case csc == nil && sc == nil:
		return nil, fmt.Errorf("no matching class found for k8s name '%s'", kubeName)
	case csc != nil && sc == nil:
		return csc, nil
	case csc == nil && sc != nil:
		return sc, nil
	default:
		return nil, fmt.Errorf("this error shouldn't be happening")
	}
}

// RetrieveClassByPlan gets the class associated to a plan.
func (sdk *SDK) RetrieveClassByPlan(plan Plan) (Class, error) {
	var class Class
	var err error

	if plan.GetNamespace() == "" {
		class, err = sdk.ServiceCatalog().ClusterServiceClasses().Get(plan.GetClassID(), metav1.GetOptions{})
	} else {
		class, err = sdk.ServiceCatalog().ServiceClasses(plan.GetNamespace()).Get(plan.GetClassID(), metav1.GetOptions{})
	}
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
