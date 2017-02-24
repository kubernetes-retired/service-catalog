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

package wip

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/util/runtime"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1alpha1"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers/servicecatalog/v1alpha1"
	listers "github.com/kubernetes-incubator/service-catalog/pkg/client/listers/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/injector"
)

// NewController returns a new Open Service Broker catalog
// controller.
func NewController(
	kubeClient kubernetes.Interface,
	serviceCatalogClient servicecatalogclientset.ServicecatalogV1alpha1Interface,
	brokerInformer informers.BrokerInformer,
	serviceClassInformer informers.ServiceClassInformer,
	instanceInformer informers.InstanceInformer,
	bindingInformer informers.BindingInformer,
	brokerClientCreateFunc brokerapi.CreateFunc,
) (Controller, error) {
	var (
		brokerLister       = brokerInformer.Lister()
		serviceClassLister = serviceClassInformer.Lister()
		instanceLister     = instanceInformer.Lister()

		controller = &controller{
			kubeClient:             kubeClient,
			serviceCatalogClient:   serviceCatalogClient,
			brokerClientCreateFunc: brokerClientCreateFunc,
			brokerLister:           brokerLister,
			serviceClassLister:     serviceClassLister,
			instanceLister:         instanceLister,
		}
	)

	brokerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.brokerAdd,
		UpdateFunc: controller.brokerUpdate,
		DeleteFunc: controller.brokerDelete,
	})

	serviceClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.serviceClassAdd,
		UpdateFunc: controller.serviceClassUpdate,
		DeleteFunc: controller.serviceClassDelete,
	})

	instanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.instanceAdd,
		UpdateFunc: controller.instanceUpdate,
		DeleteFunc: controller.instanceDelete,
	})

	bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.bindingAdd,
		UpdateFunc: controller.bindingUpdate,
		DeleteFunc: controller.bindingDelete,
	})

	return controller, nil
}

// Controller describes a controller that backs the service catalog API for
// Open Service Broker compliant Brokers.
type Controller interface {
	// Run runs the controller until the given stop channel can be read from.
	Run(stopCh <-chan struct{})
}

// controller is a concrete Controller.
type controller struct {
	kubeClient             kubernetes.Interface
	serviceCatalogClient   servicecatalogclientset.ServicecatalogV1alpha1Interface
	brokerClientCreateFunc brokerapi.CreateFunc
	brokerLister           listers.BrokerLister
	serviceClassLister     listers.ServiceClassLister
	instanceLister         listers.InstanceLister
}

// Run runs the controller until the given stop channel can be read from.
func (c *controller) Run(stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	glog.Info("Starting service-catalog controller")

	<-stopCh
	glog.Info("Shutting down service-catalog controller")
}

// Broker handlers and control-loop

func (c *controller) brokerAdd(obj interface{}) {
	broker, ok := obj.(*v1alpha1.Broker)
	if broker == nil || !ok {
		return
	}

	c.reconcileBroker(broker)
}

func (c *controller) brokerUpdate(oldObj, newObj interface{}) {
	c.brokerAdd(newObj)
}

func (c *controller) brokerDelete(obj interface{}) {
	broker, ok := obj.(*v1alpha1.Broker)
	if broker == nil || !ok {
		return
	}

	glog.V(4).Info("Received delete event for Broker %v", broker.Name)
}

const (
	ErrorFetchingCatalogReason  string = "ErrorFetchingCatalog"
	ErrorFetchingCatalogMessage string = "Error fetching catalog"
	ErrorSyncingCatalogReason   string = "ErrorSyncingCatalog"
	ErrorSyncingCatalogMessage  string = "Error syncing catalog from Broker"
)

// reconcileBroker is the control-loop that reconciles a Broker.
func (c *controller) reconcileBroker(broker *v1alpha1.Broker) {
	glog.V(4).Infof("Processing Broker %v", broker.Name)

	username, password, err := GetAuthCredentialsFromBroker(c.kubeClient, broker)
	if err != nil {
		glog.Errorf("Error getting broker auth credentials for broker %v: %v", broker.Name, err)
		c.updateBrokerReadyCondition(broker, v1alpha1.ConditionFalse, ErrorFetchingCatalogReason, ErrorFetchingCatalogMessage)
		return
	}

	glog.V(4).Infof("Creating client for Broker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient := c.brokerClientCreateFunc(broker.Name, broker.Spec.URL, username, password)
	brokerCatalog, err := brokerClient.GetCatalog()
	if err != nil {
		glog.Errorf("Error getting broker catalog for broker %v: %v", broker.Name, err)
		err := c.updateBrokerReadyCondition(broker, v1alpha1.ConditionFalse, ErrorFetchingCatalogReason, ErrorFetchingCatalogMessage)
		if err != nil {
			glog.Errorf("Error updating ready condition for Broker %v: %v", broker.Name, err)
		}

		return
	} else {
		glog.V(5).Infof("Successfully fetched %v catalog entries for Broker %v", len(brokerCatalog.Services), broker.Name)
	}

	glog.V(4).Infof("Converting catalog response for Broker %v into service-catalog API", broker.Name)
	catalog, err := convertCatalog(brokerCatalog)
	if err != nil {
		glog.Errorf("Error converting catalog payload for broker %v to service-catalog API: %v", broker.Name, err)
		c.updateBrokerReadyCondition(broker, v1alpha1.ConditionFalse, ErrorSyncingCatalogReason, ErrorSyncingCatalogMessage)
		return
	} else {
		glog.V(5).Infof("Successfully converted catalog payload from Broker %v to service-catalog API", broker.Name)
	}

	for _, serviceClass := range catalog {
		glog.V(4).Infof("Reconciling serviceClass %v (broker %v)", serviceClass.Name, broker.Name)
		if err := c.reconcileServiceClassFromBrokerCatalog(broker, serviceClass); err != nil {
			glog.Errorf("Error reconciling serviceClass %v (broker %v): %v", serviceClass.Name, broker.Name, err)
			c.updateBrokerReadyCondition(broker, v1alpha1.ConditionFalse, ErrorSyncingCatalogReason, ErrorSyncingCatalogMessage)
			return
		} else {
			glog.V(5).Infof("Reconciled serviceClass %v (broker %v)", serviceClass.Name, broker.Name)
		}
	}

	c.updateBrokerReadyCondition(broker, v1alpha1.ConditionTrue, "FetchedCatalog", "Successfully fetched catalog from broker")
}

// reconcileServiceClassFromBrokerCatalog reconciles a ServiceClass after the
// Broker's catalog has been re-listed.
func (c *controller) reconcileServiceClassFromBrokerCatalog(broker *v1alpha1.Broker, serviceClass *v1alpha1.ServiceClass) error {
	serviceClass.BrokerName = broker.Name

	existingServiceClass, err := c.serviceClassLister.Get(serviceClass.Name)
	if err != nil {
		// An error returned from a lister Get call means that the object does
		// not exist.  Create a new ServiceClass.
		if _, err := c.serviceCatalogClient.ServiceClasses().Create(serviceClass); err != nil {
			glog.Errorf("Error creating serviceClass %v from Broker %v: %v", serviceClass.Name, broker.Name, err)
			return err
		}

		return nil
	}

	// There was an existing service class -- project the update onto it and
	// update it.
	clone, err := api.Scheme.DeepCopy(existingServiceClass)
	if err != nil {
		return err
	}

	toUpdate := clone.(*v1alpha1.ServiceClass)
	toUpdate.Bindable = serviceClass.Bindable
	toUpdate.Plans = serviceClass.Plans
	toUpdate.PlanUpdatable = serviceClass.PlanUpdatable
	toUpdate.OSBTags = serviceClass.OSBTags
	toUpdate.OSBRequires = serviceClass.OSBRequires
	toUpdate.OSBMaxDBPerNode = serviceClass.OSBMaxDBPerNode
	toUpdate.OSBDashboardOAuth2ClientID = serviceClass.OSBDashboardOAuth2ClientID
	toUpdate.OSBDashboardSecret = serviceClass.OSBDashboardSecret
	toUpdate.OSBDashboardRedirectURI = serviceClass.OSBDashboardRedirectURI

	toUpdate.Description = serviceClass.Description
	toUpdate.DisplayName = serviceClass.DisplayName
	toUpdate.ImageURL = serviceClass.ImageURL
	toUpdate.LongDescription = serviceClass.LongDescription
	toUpdate.ProviderDisplayName = serviceClass.ProviderDisplayName
	toUpdate.DocumentationURL = serviceClass.DocumentationURL
	toUpdate.SupportURL = serviceClass.SupportURL

	if _, err := c.serviceCatalogClient.ServiceClasses().Update(toUpdate); err != nil {
		glog.Errorf("Error updating serviceClass %v from Broker %v: %v", serviceClass.Name, broker.Name, err)
		return err
	}

	return nil
}

// updateBrokerReadyCondition updates the ready condition for the given Broker
// with the given status, reason, and message.
func (c *controller) updateBrokerReadyCondition(broker *v1alpha1.Broker, status v1alpha1.ConditionStatus, reason, message string) error {

	clone, err := api.Scheme.DeepCopy(broker)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.Broker)
	toUpdate.Status.Conditions = []v1alpha1.BrokerCondition{{
		Type:    v1alpha1.BrokerConditionReady,
		Status:  status,
		Reason:  reason,
		Message: message,
	}}

	glog.V(4).Infof("Updating ready condition for Broker %v to %v", broker.Name, status)
	_, err = c.serviceCatalogClient.Brokers().UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating ready condition for Broker %v: %v", broker.Name, err)
	} else {
		glog.V(5).Infof("Updated ready condition for Broker %v to %v", broker.Name, status)
	}

	return err
}

// Service class handlers and control-loop

func (c *controller) serviceClassAdd(obj interface{}) {
	serviceClass, ok := obj.(*v1alpha1.ServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	c.reconcileServiceClass(serviceClass)
}

func (c *controller) reconcileServiceClass(serviceClass *v1alpha1.ServiceClass) {
	glog.V(4).Infof("Processing Instance %v", serviceClass.Name)
}

func (c *controller) serviceClassUpdate(oldObj, newObj interface{}) {
	c.serviceClassAdd(newObj)
}

func (c *controller) serviceClassDelete(obj interface{}) {
	serviceClass, ok := obj.(*v1alpha1.ServiceClass)
	if serviceClass == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for ServiceClass %v", serviceClass.Name)
}

// Instance handlers and control-loop

func (c *controller) instanceAdd(obj interface{}) {
	instance, ok := obj.(*v1alpha1.Instance)
	if instance == nil || !ok {
		return
	}

	c.reconcileInstance(instance)
}

func (c *controller) instanceUpdate(oldObj, newObj interface{}) {
	c.instanceAdd(newObj)
}

// const (
// 	ErrorFetchingCatalogReason  string = "ErrorFetchingCatalog"
// 	ErrorFetchingCatalogMessage string = "Error fetching catalog"
// 	ErrorSyncingCatalogReason   string = "ErrorSyncingCatalog"
// 	ErrorSyncingCatalogMessage  string = "Error syncing catalog from Broker"
// )

// reconcileInstance is the control-loop for reconciling Instances.
func (c *controller) reconcileInstance(instance *v1alpha1.Instance) {
	glog.V(4).Infof("Processing Instance %v/%v", instance.Namespace, instance.Name)

	serviceClass, err := c.serviceClassLister.Get(instance.Spec.ServiceClassName)
	if err != nil {
		glog.Errorf("Instance %v/%v references a non-existent ServiceClass %v", instance.Namespace, instance.Name, instance.Spec.ServiceClassName)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServiceClass",
			"The instance references a ServiceClass that does not exist",
		)
		return
	}

	servicePlan := findServicePlan(instance.Spec.PlanName, serviceClass.Plans)
	if servicePlan == nil {
		glog.Errorf("Instance %v/%v references a non-existent ServicePlan %v on ServiceClass %v", instance.Namespace, instance.Name, servicePlan.Name, serviceClass.Name)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServicePlan",
			"The instance references a ServicePlan that does not exist",
		)
		return
	}

	broker, err := c.brokerLister.Get(serviceClass.BrokerName)
	if err != nil {
		glog.Errorf("Instance %v/%v references a non-existent broker %v", instance.Namespace, instance.Name, serviceClass.BrokerName)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentBroker",
			"The instance references a Broker that does not exist",
		)
		return
	}

	username, password, err := GetAuthCredentialsFromBroker(c.kubeClient, broker)
	if err != nil {
		glog.Errorf("Error getting broker auth credentials for broker %v: %v", broker.Name, err)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ErrorGettingAuthCredentials",
			"Error getting auth credentials",
		)
		return
	}

	glog.V(4).Infof("Creating client for Broker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient := c.brokerClientCreateFunc(broker.Name, broker.Spec.URL, username, password)

	parameters := make(map[string]interface{})
	if len(instance.Spec.Parameters.Raw) > 0 {
		err = yaml.Unmarshal([]byte(instance.Spec.Parameters.Raw), &parameters)
		if err != nil {
			glog.Errorf("Failed to unmarshal Instance parameters\n%s\n %v", instance.Spec.Parameters, err)
			c.updateInstanceCondition(
				instance,
				v1alpha1.InstanceConditionReady,
				v1alpha1.ConditionFalse,
				"ErrorWithParameters",
				"Error unmarshaling instance parameters",
			)
			return
		}
	}

	request := &brokerapi.CreateServiceInstanceRequest{
		ServiceID:  serviceClass.OSBGUID,
		PlanID:     servicePlan.OSBGUID,
		Parameters: parameters,
	}

	// TODO: handle async provisioning

	glog.V(4).Infof("Provisioning a new Instance %v/%v of ServiceClass %v at Broker %v", instance.Namespace, instance.Name, serviceClass.Name, broker.Name)
	response, err := brokerClient.CreateServiceInstance(instance.Spec.OSBGUID, request)
	if err != nil {
		glog.Errorf("Error provisioning Instance %v/%v of ServiceClass %v at Broker %v: %v", instance.Namespace, instance.Name, serviceClass.Name, broker.Name, err)
		c.updateInstanceCondition(
			instance,
			v1alpha1.InstanceConditionReady,
			v1alpha1.ConditionFalse,
			"ProvisionCallFailed",
			"Provision call failed")
		return
	} else {
		glog.V(5).Infof("Successfully provisioned Instance %v/%v of ServiceClass %v at Broker %v: response: %v", instance.Namespace, instance.Name, serviceClass.Name, broker.Name, response)
	}

	// TODO: process response

	c.updateInstanceCondition(
		instance,
		v1alpha1.InstanceConditionReady,
		v1alpha1.ConditionTrue,
		"ProvisionedSuccessfully",
		"The instance was provisioned successfully",
	)
}

func findServicePlan(name string, plans []v1alpha1.ServicePlan) *v1alpha1.ServicePlan {
	for _, plan := range plans {
		if name == plan.Name {
			return &plan
		}
	}

	return nil
}

// updateInstanceCondition updates the given condition for the given Instance
// with the given status, reason, and message.
func (c *controller) updateInstanceCondition(
	instance *v1alpha1.Instance,
	conditionType v1alpha1.InstanceConditionType,
	status v1alpha1.ConditionStatus,
	reason, message string) error {

	clone, err := api.Scheme.DeepCopy(instance)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.Instance)

	toUpdate.Status.Conditions = []v1alpha1.InstanceCondition{{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}}

	glog.V(4).Infof("Updating %v condition for Instance %v/%v to %v", conditionType, instance.Namespace, instance.Name, status)
	_, err = c.serviceCatalogClient.Instances(instance.Namespace).UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Failed to update condition %v for Instance %v/%v to true: %v", conditionType, instance.Namespace, instance.Name, err)
	}

	return err
}

func (c *controller) instanceDelete(obj interface{}) {
	instance, ok := obj.(*v1alpha1.Instance)
	if instance == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for Instance %v/%v", instance.Namespace, instance.Name)
}

// Binding handlers and control-loop

func (c *controller) bindingAdd(obj interface{}) {
	binding, ok := obj.(*v1alpha1.Binding)
	if binding == nil || !ok {
		return
	}

	c.reconcileBinding(binding)
}

func (c *controller) bindingUpdate(oldObj, newObj interface{}) {
	c.bindingAdd(newObj)
}

func (c *controller) reconcileBinding(binding *v1alpha1.Binding) {
	glog.V(4).Infof("Processing Binding %v/%v", binding.Namespace, binding.Name)

	instance, err := c.instanceLister.Instances(binding.Namespace).Get(binding.Spec.InstanceRef.Name)
	if err != nil {
		glog.Errorf("Binding %v/%v references a non-existent Instance %v/%v", binding.Namespace, binding.Name, binding.Namespace, instance.Name)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServiceClass",
			"The binding references an Instance that does not exist",
		)
		return
	}

	serviceClass, err := c.serviceClassLister.Get(instance.Spec.ServiceClassName)
	if err != nil {
		glog.Errorf("Binding %v/%v references a non-existent ServiceClass %v", binding.Namespace, binding.Name, instance.Spec.ServiceClassName)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServiceClass",
			"The binding references a ServiceClass that does not exist",
		)
		return
	}

	broker, err := c.brokerLister.Get(serviceClass.BrokerName)
	if err != nil {
		glog.Errorf("Binding %v/%v references a non-existent Broker %v", binding.Namespace, binding.Name, serviceClass.BrokerName)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentBroker",
			"The binding references a Broker that does not exist",
		)
		return
	}

	servicePlan := findServicePlan(instance.Spec.PlanName, serviceClass.Plans)
	if servicePlan == nil {
		glog.Errorf("Instance %v/%v references a non-existent ServicePlan %v on ServiceClass %v", instance.Namespace, instance.Name, servicePlan.Name, serviceClass.Name)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ReferencesNonexistentServicePlan",
			"The binding references a ServicePlan that does not exist",
		)
		return
	}

	username, password, err := GetAuthCredentialsFromBroker(c.kubeClient, broker)
	if err != nil {
		glog.Errorf("Error getting broker auth credentials for broker %v: %v", broker.Name, err)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ErrorGettingAuthCredentials",
			"Error getting auth credentials",
		)
		return
	}

	glog.V(4).Infof("Creating client for Broker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient := c.brokerClientCreateFunc(broker.Name, broker.Spec.URL, username, password)

	parameters := make(map[string]interface{})
	if len(binding.Spec.Parameters.Raw) > 0 {
		err = yaml.Unmarshal([]byte(binding.Spec.Parameters.Raw), &parameters)
		if err != nil {
			glog.Errorf("Failed to unmarshal Binding parameters\n%s\n %v", binding.Spec.Parameters, err)
			c.updateBindingCondition(
				binding,
				v1alpha1.BindingConditionReady,
				v1alpha1.ConditionFalse,
				"ErrorWithParameters",
				"Error unmarshaling binding parameters",
			)
			return
		}
	}

	request := &brokerapi.BindingRequest{
		ServiceID:  serviceClass.OSBGUID,
		PlanID:     servicePlan.OSBGUID,
		Parameters: parameters,
	}

	response, err := brokerClient.CreateServiceBinding(instance.Spec.OSBGUID, binding.Spec.OSBGUID, request)

	err = c.injectBinding(binding, &response.Credentials)
	if err != nil {
		glog.Errorf("Error injecting binding results for Binding %v/%v: %v", binding.Namespace, binding.Name, err)
		c.updateBindingCondition(
			binding,
			v1alpha1.BindingConditionReady,
			v1alpha1.ConditionFalse,
			"ErrorInjectingBindResult",
			"Error injecting bind result",
		)
		return
	}

	c.updateBindingCondition(
		binding,
		v1alpha1.BindingConditionReady,
		v1alpha1.ConditionTrue,
		"InjectedBindResult",
		"Injected bind result",
	)
}

func (c *controller) injectBinding(binding *v1alpha1.Binding, credentials *brokerapi.Credential) error {
	secret := &v1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      binding.Spec.SecretName,
			Namespace: binding.Namespace,
		},
		Data: make(map[string][]byte),
	}

	for k, v := range *credentials {
		var err error
		secret.Data[k], err = injector.Serialize(v)
		if err != nil {
			return fmt.Errorf("Unable to serialize credential value %q: %v; %s",
				k, v, err)
		}
	}

	found := false

	_, err := c.kubeClient.Core().Secrets(binding.Namespace).Get(binding.Spec.SecretName)
	if err == nil {
		found = true
	}

	if found {
		_, err = c.kubeClient.Core().Secrets(binding.Namespace).Update(secret)
	} else {
		_, err = c.kubeClient.Core().Secrets(binding.Namespace).Create(secret)
	}

	return err
}

// updateBindingCondition updates the given condition for the given Binding
// with the given status, reason, and message.
func (c *controller) updateBindingCondition(
	binding *v1alpha1.Binding,
	conditionType v1alpha1.BindingConditionType,
	status v1alpha1.ConditionStatus,
	reason, message string) error {

	clone, err := api.Scheme.DeepCopy(binding)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.Binding)

	toUpdate.Status.Conditions = []v1alpha1.BindingCondition{{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}}

	logContext := fmt.Sprintf("%v condition for Binding %v/%v to %v (Reason: %q, Message: %q)",
		conditionType, binding.Namespace, binding.Name, status, reason, message)

	glog.V(4).Infof("Updating %v", logContext)
	_, err = c.serviceCatalogClient.Bindings(binding.Namespace).UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating %v: %v", logContext, err)
	}
	return err
}

func (c *controller) bindingDelete(obj interface{}) {
	binding, ok := obj.(*v1alpha1.Binding)
	if binding == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for Binding %v/%v", binding.Namespace, binding.Name)
}

// Broker utility methods - move?

// GetAuthCredentialsFromBroker returns the auth credentials, if any,
// contained in the secret referenced in the Broker's AuthSecret field, or
// returns an error. If the AuthSecret field is nil, empty values are
// returned.
func GetAuthCredentialsFromBroker(client kubernetes.Interface, broker *v1alpha1.Broker) (username, password string, err error) {
	if broker.Spec.AuthSecret == nil {
		return "", "", nil
	}

	authSecret, err := client.Core().Secrets(broker.Spec.AuthSecret.Namespace).Get(broker.Spec.AuthSecret.Name)
	if err != nil {
		return "", "", err
	}

	usernameBytes, ok := authSecret.Data["username"]
	if !ok {
		return "", "", fmt.Errorf("auth secret didn't contain username")
	}

	passwordBytes, ok := authSecret.Data["password"]
	if !ok {
		return "", "", fmt.Errorf("auth secret didn't contain password")
	}

	return string(usernameBytes), string(passwordBytes), nil
}

// convertCatalog converts a service broker catalog into an array of ServiceClasses
func convertCatalog(in *brokerapi.Catalog) ([]*v1alpha1.ServiceClass, error) {
	ret := make([]*v1alpha1.ServiceClass, len(in.Services))
	for i, svc := range in.Services {
		plans := convertServicePlans(svc.Plans)
		ret[i] = &v1alpha1.ServiceClass{
			Bindable:      svc.Bindable,
			Plans:         plans,
			PlanUpdatable: svc.PlanUpdateable,
			OSBGUID:       svc.ID,
			OSBTags:       svc.Tags,
			OSBRequires:   svc.Requires,
			// OSBMetadata:   svc.Metadata,
		}
		ret[i].SetName(svc.Name)
	}
	return ret, nil
}

func convertServicePlans(plans []brokerapi.ServicePlan) []v1alpha1.ServicePlan {
	ret := make([]v1alpha1.ServicePlan, len(plans))
	for i, plan := range plans {
		ret[i] = v1alpha1.ServicePlan{
			Name:    plan.Name,
			OSBGUID: plan.ID,
			// OSBMetadata: plan.Metadata,
			OSBFree: plan.Free,
		}
	}
	return ret
}
