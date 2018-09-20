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
	// FieldExternalPlanName is the jsonpath to a plan's external name.
	FieldExternalPlanName = "spec.externalName"

	// FieldServiceClassRef is the jsonpath to a plan's associated class name.
	FieldServiceClassRef = "spec.clusterServiceClassRef.name"
)

// RetrievePlanOptions allows to specify which plans will be retrieved
type RetrievePlanOptions struct {
	ClassID   string
	Namespace string
	Scope     Scope
}

// Plan provides a unifying layer of cluster and namespace scoped plan resources.
type Plan interface {

	// GetName returns the plan's name.
	GetName() string

	// GetNamespace returns the plan's namespace, or "" if it's cluster-scoped.
	GetNamespace() string

	// GetExternalName returns the plan's external name.
	GetExternalName() string

	// GetDescription returns the plan description.
	GetDescription() string

	// GetClassID returns the plan's class name.
	GetClassID() string
}

// RetrievePlans lists all plans defined in the cluster.
func (sdk *SDK) RetrievePlans(opts RetrievePlanOptions) ([]Plan, error) {
	var plans []Plan

	if opts.Scope.Matches(ClusterScope) {
		csp, err := sdk.ServiceCatalog().ClusterServicePlans().List(v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list cluster-scoped plans (%s)", err)
		}

		for _, p := range csp.Items {
			if opts.ClassID != "" && p.GetClassID() != opts.ClassID {
				continue
			}

			plan := p
			plans = append(plans, &plan)
		}
	}

	if opts.Scope.Matches(NamespaceScope) {
		sp, err := sdk.ServiceCatalog().ServicePlans(opts.Namespace).List(v1.ListOptions{})
		if err != nil {
			// Gracefully handle when the feature-flag for namespaced broker resources isn't enabled on the server.
			if errors.IsNotFound(err) {
				return plans, nil
			}
			return nil, fmt.Errorf("unable to list plans in %q (%s)", opts.Namespace, err)
		}
		for _, p := range sp.Items {
			if opts.ClassID != "" && p.GetClassID() != opts.ClassID {
				continue
			}

			plan := p
			plans = append(plans, &plan)
		}
	}

	return plans, nil
}

// RetrievePlanByName gets a plan by its external name.
func (sdk *SDK) RetrievePlanByName(name string) (*v1beta1.ClusterServicePlan, error) {
	opts := v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(FieldExternalPlanName, name).String(),
	}
	searchResults, err := sdk.ServiceCatalog().ClusterServicePlans().List(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to search plans by name '%s', (%s)", name, err)
	}
	if len(searchResults.Items) == 0 {
		return nil, fmt.Errorf("plan not found '%s'", name)
	}
	if len(searchResults.Items) > 1 {
		return nil, fmt.Errorf("more than one matching plan found for '%s'", name)
	}
	return &searchResults.Items[0], nil
}

// RetrievePlanByID gets a plan by its UUID.
func (sdk *SDK) RetrievePlanByID(uuid string) (*v1beta1.ClusterServicePlan, error) {
	plan, err := sdk.ServiceCatalog().ClusterServicePlans().Get(uuid, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get plan by uuid '%s' (%s)", uuid, err)
	}
	return plan, nil
}

// RetrievePlanByClassAndPlanNames gets a plan by its class/plan name combination.
func (sdk *SDK) RetrievePlanByClassAndPlanNames(className, planName string,
) (*v1beta1.ClusterServicePlan, error) {
	class, err := sdk.RetrieveClassByName(className, ScopeOptions{Scope: ClusterScope})
	if err != nil {
		return nil, err
	}

	planOpts := v1.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector(FieldServiceClassRef, class.GetName()),
			fields.OneTermEqualSelector(FieldExternalPlanName, planName),
		).String(),
	}
	searchResults, err := sdk.ServiceCatalog().ClusterServicePlans().List(planOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to search plans by class/plan name '%s/%s' (%s)", className, planName, err)
	}
	if len(searchResults.Items) == 0 {
		return nil, fmt.Errorf("plan not found '%s/%s'", className, planName)
	}
	if len(searchResults.Items) > 1 {
		// Note: Should never occur, as class/plan name combo must be unique
		return nil, fmt.Errorf("more than one matching plan found for '%s/%s'", className, planName)
	}
	return &searchResults.Items[0], nil
}
