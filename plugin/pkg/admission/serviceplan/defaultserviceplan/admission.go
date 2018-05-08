/*
Copyright 2017 The Kubernetes Authors.

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

package defaultserviceplan

import (
	"errors"
	"fmt"
	"io"

	"github.com/golang/glog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apiserver/pkg/admission"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset"
	servicecataloginternalversion "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset/typed/servicecatalog/internalversion"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scadmission "github.com/kubernetes-incubator/service-catalog/pkg/apiserver/admission"
)

const (
	// PluginName is name of admission plug-in
	PluginName = "DefaultServicePlan"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(io.Reader) (admission.Interface, error) {
		return NewDefaultClusterServicePlan()
	})
}

// exists is an implementation of admission.Interface.
// It checks to see if Service Instance is being created without
// a Service Plan if there is only one Service Plan for the
// specified Service and defaults to that value.
// that the cluster actually has support for it.
type defaultServicePlan struct {
	*admission.Handler
	internalClientSet internalclientset.Interface
	cscClient         servicecataloginternalversion.ClusterServiceClassInterface
	cspClient         servicecataloginternalversion.ClusterServicePlanInterface
	scClient          servicecataloginternalversion.ServiceClassInterface
	spClient          servicecataloginternalversion.ServicePlanInterface
}

var _ = scadmission.WantsInternalServiceCatalogClientSet(&defaultServicePlan{})

func (d *defaultServicePlan) Admit(a admission.Attributes) error {
	// We only care about service Instances
	if a.GetResource().Group != servicecatalog.GroupName || a.GetResource().GroupResource() != servicecatalog.Resource("serviceinstances") {
		return nil
	}
	instance, ok := a.GetObject().(*servicecatalog.ServiceInstance)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind ServiceInstance but was unable to be converted")
	}

	// If the plan is specified, let it through and have the controller
	// deal with finding the right plan, etc.
	if instance.Spec.ClusterServicePlanSpecified() || instance.Spec.ServicePlanSpecified() {
		return nil
	}

	if instance.Spec.ClusterServiceClassSpecified() {
		return d.handleDefaultClusterServicePlan(a, instance)
	} else if instance.Spec.ServiceClassSpecified() {
		return d.handleDefaultServicePlan(a, instance)
	}

	return apierrors.NewInternalError(errors.New("Class not specified on ServiceInstance, cannot choose default plan"))

}

func (d *defaultServicePlan) handleDefaultClusterServicePlan(a admission.Attributes, instance *servicecatalog.ServiceInstance) error {
	sc, err := d.getClusterServiceClassByPlanReference(a, &instance.Spec.PlanReference)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return admission.NewForbidden(a, err)
		}
		msg := fmt.Sprintf("ClusterServiceClass %c does not exist, can not figure out the default ClusterServicePlan.",
			instance.Spec.PlanReference)
		glog.V(4).Info(msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ClusterServicePlan available while others are still in progress?
	// Not currently. Creation of all ClusterServicePlans before creating
	// the ClusterServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getClusterServicePlansByClusterServiceClassName(sc.Name)
	if err != nil {
		msg := fmt.Sprintf("Error listing ClusterServicePlans for ClusterServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ClusterServicePlan", sc.Name, sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) <= 0 {
		msg := fmt.Sprintf("no ClusterServicePlans found at all for ClusterServiceClass %q", sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// check if more than one service plan was specified and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ClusterServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", sc.Name, sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	glog.V(4).Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		instance.Namespace, instance.Name, p.Spec.ExternalName, p.Name, sc.Spec.ExternalName)
	if instance.Spec.ClusterServiceClassExternalName != "" {
		instance.Spec.ClusterServicePlanExternalName = p.Spec.ExternalName
	} else if instance.Spec.ClusterServiceClassExternalID != "" {
		instance.Spec.ClusterServicePlanExternalID = p.Spec.ExternalID
	} else {
		instance.Spec.ClusterServicePlanName = p.Name
	}

	return nil
}

func (d *defaultServicePlan) handleDefaultServicePlan(a admission.Attributes, instance *servicecatalog.ServiceInstance) error {
	// Need to dynamically set the namespaced clients based on the namespace of the instance
	d.scClient = d.internalClientSet.Servicecatalog().ServiceClasses(instance.Namespace)
	d.spClient = d.internalClientSet.Servicecatalog().ServicePlans(instance.Namespace)

	sc, err := d.getServiceClassByPlanReference(a, &instance.Spec.PlanReference)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return admission.NewForbidden(a, err)
		}
		msg := fmt.Sprintf("ClusterServiceClass %c does not exist, can not figure out the default ClusterServicePlan.",
			instance.Spec.PlanReference)
		glog.V(4).Info(msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// find all the service plans that belong to the service class

	// Need to be careful here. Is it possible to have only one
	// ServicePlan available while others are still in progress?
	// Not currently. Creation of all ServicePlans before creating
	// the ServiceClass ensures that this will work correctly. If
	// the order changes, we will need to rethink the
	// implementation of this controller.
	plans, err := d.getServicePlansByServiceClassName(sc.Name)
	if err != nil {
		msg := fmt.Sprintf("Error listing ServicePlans for ServiceClass (K8S: %v ExternalName: %v) - retry and specify desired ClusterServicePlan", sc.Name, sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// check if there were any service plans
	// TODO: in combination with not allowing classes with no plans, this should be impossible
	if len(plans) <= 0 {
		msg := fmt.Sprintf("no ServicePlans found at all for ServiceClass %q", sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}

	// check if more than one service plan was specified and error
	if len(plans) > 1 {
		msg := fmt.Sprintf("ServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", sc.Name, sc.Spec.ExternalName)
		glog.V(4).Infof(`ServiceInstance "%s/%s": %s`, instance.Namespace, instance.Name, msg)
		return admission.NewForbidden(a, errors.New(msg))
	}
	// otherwise, by default, pick the only plan that exists for the service class

	p := plans[0]
	glog.V(4).Infof(`ServiceInstance "%s/%s": Using default plan %q (K8S: %q) for Service Class %q`,
		instance.Namespace, instance.Name, p.Spec.ExternalName, p.Name, sc.Spec.ExternalName)
	if instance.Spec.ServiceClassExternalName != "" {
		instance.Spec.ServicePlanExternalName = p.Spec.ExternalName
	} else if instance.Spec.ServiceClassExternalID != "" {
		instance.Spec.ServicePlanExternalID = p.Spec.ExternalID
	} else {
		instance.Spec.ServicePlanName = p.Name
	}

	return nil
}

// NewDefaultClusterServicePlan creates a new admission control handler that
// fills in a default Service Plan if omitted from Service Instance
// creation request and if there exists only one plan in the
// specified Service Class
func NewDefaultClusterServicePlan() (admission.Interface, error) {
	return &defaultServicePlan{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}

func (d *defaultServicePlan) SetInternalServiceCatalogClientSet(i internalclientset.Interface) {
	d.cscClient = i.Servicecatalog().ClusterServiceClasses()
	d.cspClient = i.Servicecatalog().ClusterServicePlans()
	d.internalClientSet = i
}

func (d *defaultServicePlan) ValidateInitialization() error {
	if d.cscClient == nil {
		return errors.New("missing clusterserviceclass interface")
	}
	if d.cspClient == nil {
		return errors.New("missing clusterserviceplan interface")
	}
	return nil
}

func (d *defaultServicePlan) getClusterServiceClassByPlanReference(a admission.Attributes, ref *servicecatalog.PlanReference) (*servicecatalog.ClusterServiceClass, error) {
	if ref.ClusterServiceClassName != "" {
		return d.getClusterServiceClassByK8SName(a, ref.ClusterServiceClassName)
	}

	return d.getClusterServiceClassByField(a, ref)
}

func (d *defaultServicePlan) getServiceClassByPlanReference(a admission.Attributes, ref *servicecatalog.PlanReference) (*servicecatalog.ServiceClass, error) {
	if ref.ServiceClassName != "" {
		return d.getServiceClassByK8SName(a, ref.ServiceClassName)
	}

	return d.getServiceClassByField(a, ref)
}

func (d *defaultServicePlan) getClusterServiceClassByK8SName(a admission.Attributes, scK8SName string) (*servicecatalog.ClusterServiceClass, error) {
	glog.V(4).Infof("Fetching ClusterServiceClass by k8s name %q", scK8SName)
	return d.cscClient.Get(scK8SName, apimachineryv1.GetOptions{})
}

func (d *defaultServicePlan) getServiceClassByK8SName(a admission.Attributes, scK8SName string) (*servicecatalog.ServiceClass, error) {
	glog.V(4).Infof("Fetching ServiceClass by k8s name %q", scK8SName)
	return d.scClient.Get(scK8SName, apimachineryv1.GetOptions{})
}

func (d *defaultServicePlan) getClusterServiceClassByField(a admission.Attributes, ref *servicecatalog.PlanReference) (*servicecatalog.ClusterServiceClass, error) {
	filterField := ref.GetClassFilterFieldName()
	filterValue := ref.GetSpecifiedClusterServiceClass()

	glog.V(4).Infof("Fetching ClusterServiceClass filtered by %q = %q", filterField, filterValue)
	fieldSet := fields.Set{
		filterField: filterValue,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := apimachineryv1.ListOptions{FieldSelector: fieldSelector}
	serviceClasses, err := d.cscClient.List(listOpts)
	if err != nil {
		glog.V(4).Infof("Listing ClusterServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClasses.Items) == 1 {
		glog.V(4).Infof("Found single ClusterServiceClass as %+v", serviceClasses.Items[0])
		return &serviceClasses.Items[0], nil
	}
	msg := fmt.Sprintf("Could not find a single ClusterServiceClass with %q = %q, found %v", filterField, filterValue, len(serviceClasses.Items))
	glog.V(4).Info(msg)
	return nil, admission.NewNotFound(a)
}

func (d *defaultServicePlan) getServiceClassByField(a admission.Attributes, ref *servicecatalog.PlanReference) (*servicecatalog.ServiceClass, error) {
	filterField := ref.GetClassFilterFieldName()
	filterValue := ref.GetSpecifiedServiceClass()

	glog.V(4).Infof("Fetching ServiceClass filtered by %q = %q", filterField, filterValue)
	fieldSet := fields.Set{
		filterField: filterValue,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := apimachineryv1.ListOptions{FieldSelector: fieldSelector}
	serviceClasses, err := d.scClient.List(listOpts)
	if err != nil {
		glog.V(4).Infof("Listing ServiceClasses failed: %q", err)
		return nil, err
	}
	if len(serviceClasses.Items) == 1 {
		glog.V(4).Infof("Found single ServiceClass as %+v", serviceClasses.Items[0])
		return &serviceClasses.Items[0], nil
	}
	msg := fmt.Sprintf("Could not find a single ServiceClass with %q = %q, found %v", filterField, filterValue, len(serviceClasses.Items))
	glog.V(4).Info(msg)
	return nil, admission.NewNotFound(a)
}

// getClusterServicePlansByClusterServiceClassName() returns a list of
// ServicePlans for the specified service class name
func (d *defaultServicePlan) getClusterServicePlansByClusterServiceClassName(scName string) ([]servicecatalog.ClusterServicePlan, error) {
	glog.V(4).Infof("Fetching ClusterServicePlans by class name %q", scName)
	fieldSet := fields.Set{
		"spec.clusterServiceClassRef.name": scName,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := apimachineryv1.ListOptions{FieldSelector: fieldSelector}
	servicePlans, err := d.cspClient.List(listOpts)
	if err != nil {
		glog.Infof("Listing ClusterServicePlans failed: %q", err)
		return nil, err
	}
	glog.V(4).Infof("ClusterServicePlans fetched by filtering classname: %+v", servicePlans.Items)
	r := servicePlans.Items
	return r, err
}

// getServicePlansByServiceClassName() returns a list of
// ServicePlans for the specified service class name
func (d *defaultServicePlan) getServicePlansByServiceClassName(scName string) ([]servicecatalog.ServicePlan, error) {
	glog.V(4).Infof("Fetching ServicePlans by class name %q", scName)
	fieldSet := fields.Set{
		"spec.serviceClassRef.name": scName,
	}
	fieldSelector := fields.SelectorFromSet(fieldSet).String()
	listOpts := apimachineryv1.ListOptions{FieldSelector: fieldSelector}
	servicePlans, err := d.spClient.List(listOpts)
	if err != nil {
		glog.Infof("Listing ServicePlans failed: %q", err)
		return nil, err
	}
	glog.V(4).Infof("ServicePlans fetched by filtering classname: %+v", servicePlans.Items)
	r := servicePlans.Items
	return r, err
}
