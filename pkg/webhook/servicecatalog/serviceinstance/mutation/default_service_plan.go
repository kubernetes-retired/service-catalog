/*
Copyright 2019 The Kubernetes Authors.

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

/*
This feature was copied from Service Catalog admission plugin https://github.com/kubernetes-sigs/service-catalog/blob/v0.1.41/plugin/pkg/admission/serviceplan/defaultserviceplan/admission.go
If you want to track previous changes please check there.
*/

package mutation

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultServicePlan holds logic which sets the default service plan
type DefaultServicePlan struct {
	client client.Client
}

// SetDefaultPlan sets the default service plan if it's not specified and if only one plan exists
func (d *DefaultServicePlan) SetDefaultPlan(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) *webhookutil.WebhookError {
	if instance.Spec.ClusterServicePlanSpecified() || instance.Spec.ServicePlanSpecified() {
		return nil
	}

	if instance.Spec.ClusterServiceClassSpecified() {
		return d.handleDefaultClusterServicePlan(ctx, instance, log)
	} else if instance.Spec.ServiceClassSpecified() {
		return d.handleDefaultServicePlan(ctx, instance, log)
	}

	return webhookutil.NewWebhookError("class not specified on ServiceInstance, cannot choose default plan", http.StatusInternalServerError)
}

func (d *DefaultServicePlan) handleDefaultClusterServicePlan(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) *webhookutil.WebhookError {
	clusterServiceClass, err := d.getClusterServiceClassByPlanReference(ctx, instance, log)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
		}
		msg := fmt.Sprintf("ClusterServiceClass %c does not exist, can not figure out the default ClusterServicePlan.",
			instance.Spec.PlanReference)
		log.V(4).Info(msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ClusterServicePlan available while others are still in progress?
	// Not currently. Creation of all ClusterServicePlans before creating
	// the ClusterServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getClusterServicePlansByClusterServiceClassName(ctx, clusterServiceClass.Name, log)
	if err != nil {
		msg := fmt.Sprintf("Error listing ClusterServicePlans for ClusterServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ClusterServicePlan", clusterServiceClass.Name, clusterServiceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) == 0 {
		msg := fmt.Sprintf("no ClusterServicePlans found at all for ClusterServiceClass %q", clusterServiceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	// check if more than one service plan was found and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ClusterServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", clusterServiceClass.Name, clusterServiceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	log.V(4).Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		instance.Namespace, instance.Name, p.Spec.ExternalName, p.Name, clusterServiceClass.Spec.ExternalName)
	if instance.Spec.ClusterServiceClassExternalName != "" {
		instance.Spec.ClusterServicePlanExternalName = p.Spec.ExternalName
	} else if instance.Spec.ClusterServiceClassExternalID != "" {
		instance.Spec.ClusterServicePlanExternalID = p.Spec.ExternalID
	} else {
		instance.Spec.ClusterServicePlanName = p.Name
	}

	return nil
}

func (d *DefaultServicePlan) handleDefaultServicePlan(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) *webhookutil.WebhookError {
	serviceClass, err := d.getServiceClassByPlanReference(ctx, instance, log)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return webhookutil.NewWebhookError(err.Error(), http.StatusForbidden)
		}
		msg := fmt.Sprintf("ServiceClass %c does not exist, can not figure out the default ServicePlan.",
			instance.Spec.PlanReference)
		log.V(4).Info(msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}
	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ServicePlan available while others are still in progress?
	// Not currently. Creation of all ServicePlans before creating
	// the ServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getServicePlansByServiceClassName(ctx, serviceClass.Name, serviceClass.Namespace, log)
	if err != nil {
		msg := fmt.Sprintf("Error listing ServicePlans for ServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ServicePlan", serviceClass.Name, serviceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) == 0 {
		msg := fmt.Sprintf("no ServicePlans found at all for ServiceClass %q", serviceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}

	// check if more than one service plan was found and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", serviceClass.Name, serviceClass.Spec.ExternalName)
		log.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return webhookutil.NewWebhookError(msg, http.StatusForbidden)
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	log.V(4).Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		instance.Namespace, instance.Name, p.Spec.ExternalName, p.Name, serviceClass.Spec.ExternalName)
	if instance.Spec.ServiceClassExternalName != "" {
		instance.Spec.ServicePlanExternalName = p.Spec.ExternalName
	} else if instance.Spec.ServiceClassExternalID != "" {
		instance.Spec.ServicePlanExternalID = p.Spec.ExternalID
	} else {
		instance.Spec.ServicePlanName = p.Name
	}

	return nil
}

func (d *DefaultServicePlan) getClusterServiceClassByPlanReference(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ClusterServiceClass, error) {
	if instance.Spec.PlanReference.ClusterServiceClassName != "" {
		return d.getClusterServiceClassByK8SName(ctx, instance, log)
	}

	return d.getClusterServiceClassByField(ctx, instance, log)
}

func (d *DefaultServicePlan) getServiceClassByPlanReference(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ServiceClass, error) {
	if instance.Spec.PlanReference.ServiceClassName != "" {
		return d.getServiceClassByK8SName(ctx, instance, log)
	}

	return d.getServiceClassByField(ctx, instance, log)
}

func (d *DefaultServicePlan) getClusterServiceClassByK8SName(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ClusterServiceClass, error) {
	log.V(4).Infof("Fetching ClusterServiceClass by k8s name %q", instance.Spec.PlanReference.ClusterServiceClassName)
	csc := &sc.ClusterServiceClass{}
	err := d.client.Get(ctx, client.ObjectKey{Name: instance.Spec.PlanReference.ClusterServiceClassName}, csc)
	return csc, err
}

func (d *DefaultServicePlan) getServiceClassByK8SName(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ServiceClass, error) {
	log.V(4).Infof("Fetching ServiceClass by k8s name %q", instance.Spec.PlanReference.ServiceClassName)
	serviceClass := &sc.ServiceClass{}
	err := d.client.Get(ctx, client.ObjectKey{Name: instance.Spec.PlanReference.ServiceClassName, Namespace: instance.Namespace}, serviceClass)

	return serviceClass, err
}

func (d *DefaultServicePlan) getClusterServiceClassByField(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ClusterServiceClass, error) {
	ref := instance.Spec.PlanReference

	filterLabel := ref.GetClusterServiceClassFilterLabelName()
	filterValue := ref.GetSpecifiedClusterServiceClass()

	log.V(4).Infof("Fetching ClusterServiceClass filtered by %q = %q", filterLabel, filterValue)

	serviceClassesList := &sc.ClusterServiceClassList{}
	err := d.client.List(ctx, serviceClassesList, client.MatchingLabels(map[string]string{
		filterLabel: util.GenerateSHA(filterValue),
	}))
	if err != nil {
		log.V(4).Infof("Listing ClusterServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClassesList.Items) == 1 {
		log.V(4).Infof("Found single ClusterServiceClass as %+v", serviceClassesList.Items[0])
		return &serviceClassesList.Items[0], nil
	}
	msg := fmt.Sprintf("could not find a single ClusterServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items))
	log.V(4).Info(msg)
	return nil, errors.New(msg)
}

func (d *DefaultServicePlan) getServiceClassByField(ctx context.Context, instance *sc.ServiceInstance, log *webhookutil.TracedLogger) (*sc.ServiceClass, error) {
	ref := instance.Spec.PlanReference

	filterLabel := ref.GetServiceClassFilterLabelName()
	filterValue := ref.GetSpecifiedServiceClass()

	log.V(4).Infof("Fetching ServiceClass filtered by %q = %q", filterLabel, filterValue)

	serviceClassesList := &sc.ServiceClassList{}
	err := d.client.List(ctx, serviceClassesList, client.MatchingLabels(map[string]string{
		filterLabel: util.GenerateSHA(filterValue),
	}), client.InNamespace(instance.Namespace))
	if err != nil {
		log.V(4).Infof("Listing ServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClassesList.Items) == 1 {
		log.V(4).Infof("Found single ServiceClass as %+v", serviceClassesList.Items[0])
		return &serviceClassesList.Items[0], nil
	}
	msg := fmt.Sprintf("could not find a single ServiceClass with %q = %q, found %v", filterLabel, filterValue, len(serviceClassesList.Items))
	log.V(4).Info(msg)
	return nil, errors.New(msg)
}

// getClusterServicePlansByClusterServiceClassName() returns a list of
// ClusterServicePlan for the specified cluster service class name
func (d *DefaultServicePlan) getClusterServicePlansByClusterServiceClassName(ctx context.Context, scName string, log *webhookutil.TracedLogger) ([]sc.ClusterServicePlan, error) {
	log.V(4).Infof("Fetching ClusterServicePlans by class name %q", scName)

	servicePlansList := &sc.ClusterServicePlanList{}
	err := d.client.List(ctx, servicePlansList, client.MatchingLabels(map[string]string{
		sc.GroupName + "/" + sc.FilterSpecClusterServiceClassRefName: util.GenerateSHA(scName),
	}))
	if err != nil {
		log.Infof("Listing ClusterServicePlans failed: %q", err)
		return nil, err
	}

	log.V(4).Infof("ClusterServicePlans fetched by filtering classname: %+v", servicePlansList.Items)
	r := servicePlansList.Items
	return r, err
}

// getServicePlansByServiceClassName() returns a list of
// ServicePlans for the specified service class name
func (d *DefaultServicePlan) getServicePlansByServiceClassName(ctx context.Context, scName string, scNamespace string, log *webhookutil.TracedLogger) ([]sc.ServicePlan, error) {
	log.V(4).Infof("Fetching ServicePlans by class name %q", scName)

	servicePlansList := &sc.ServicePlanList{}
	err := d.client.List(ctx, servicePlansList, client.MatchingLabels(map[string]string{
		sc.GroupName + "/" + sc.FilterSpecServiceClassRefName: util.GenerateSHA(scName),
	}), client.InNamespace(scNamespace))
	if err != nil {
		log.Infof("Listing ServicePlans failed: %q", err)
		return nil, err
	}
	log.V(4).Infof("ServicePlans fetched by filtering classname: %+v", servicePlansList.Items)
	r := servicePlansList.Items
	return r, err
}

// InjectClient injects client
func (d *DefaultServicePlan) InjectClient(c client.Client) error {
	d.client = c
	return nil
}
