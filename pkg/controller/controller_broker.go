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

package controller

import (
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/kubernetes-incubator/service-catalog/pkg/api"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

func (c *controller) brokerAdd(obj interface{}) {
	// DeletionHandlingMetaNamespaceKeyFunc returns a unique key for the resource and
	// handles the special case where the resource is of DeletedFinalStateUnknown type, which
	// acts a place holder for resources that have been deleted from storage but the watch event
	// confirming the deletion has not yet arrived.
	// Generally, the key is "namespace/name" for namespaced-scoped resources and
	// just "name" for cluster scoped resources.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.brokerQueue.Add(key)
}

func (c *controller) brokerUpdate(oldObj, newObj interface{}) {
	c.brokerAdd(newObj)
}

func (c *controller) brokerDelete(obj interface{}) {
	broker, ok := obj.(*v1alpha1.ClusterServiceBroker)
	if broker == nil || !ok {
		return
	}

	glog.V(4).Infof("Received delete event for ClusterServiceBroker %v; no further processing will occur", broker.Name)
}

// the Message strings have a terminating period and space so they can
// be easily combined with a follow on specific message.
const (
	errorFetchingCatalogReason                     string = "ErrorFetchingCatalog"
	errorFetchingCatalogMessage                    string = "Error fetching catalog. "
	errorSyncingCatalogReason                      string = "ErrorSyncingCatalog"
	errorSyncingCatalogMessage                     string = "Error syncing catalog from ServiceBroker. "
	errorWithParameters                            string = "ErrorWithParameters"
	errorListingClusterServiceClassesReason        string = "ErrorListingServiceClasses"
	errorListingClusterServiceClassesMessage       string = "Error listing service classes."
	errorListingClusterServicePlansReason          string = "ErrorListingServicePlans"
	errorListingClusterServicePlansMessage         string = "Error listing service plans."
	errorDeletingClusterServiceClassReason         string = "ErrorDeletingServiceClass"
	errorDeletingClusterServiceClassMessage        string = "Error deleting service class."
	errorDeletingClusterServicePlanReason          string = "ErrorDeletingServicePlan"
	errorDeletingClusterServicePlanMessage         string = "Error deleting service plan."
	errorNonexistentClusterServiceClassReason      string = "ReferencesNonexistentServiceClass"
	errorNonexistentClusterServiceClassMessage     string = "ReferencesNonexistentServiceClass"
	errorNonexistentClusterServicePlanReason       string = "ReferencesNonexistentServicePlan"
	errorNonexistentClusterServiceBrokerReason     string = "ReferencesNonexistentBroker"
	errorNonexistentServiceInstanceReason          string = "ReferencesNonexistentInstance"
	errorAuthCredentialsReason                     string = "ErrorGettingAuthCredentials"
	errorFindingNamespaceServiceInstanceReason     string = "ErrorFindingNamespaceForInstance"
	errorProvisionCallFailedReason                 string = "ProvisionCallFailed"
	errorErrorCallingProvisionReason               string = "ErrorCallingProvision"
	errorUpdateInstanceCallFailedReason            string = "UpdateInstanceCallFailed"
	errorErrorCallingUpdateInstanceReason          string = "ErrorCallingUpdateInstance"
	errorDeprovisionCalledReason                   string = "DeprovisionCallFailed"
	errorDeprovisionBlockedByCredentialsReason     string = "DeprovisionBlockedByExistingCredentials"
	errorBindCallReason                            string = "BindCallFailed"
	errorInjectingBindResultReason                 string = "ErrorInjectingBindResult"
	errorEjectingBindReason                        string = "ErrorEjectingServiceBinding"
	errorEjectingBindMessage                       string = "Error ejecting binding."
	errorUnbindCallReason                          string = "UnbindCallFailed"
	errorWithOngoingAsyncOperation                 string = "ErrorAsyncOperationInProgress"
	errorWithOngoingAsyncOperationMessage          string = "Another operation for this service instance is in progress. "
	errorNonbindableClusterServiceClassReason      string = "ErrorNonbindableServiceClass"
	errorServiceInstanceNotReadyReason             string = "ErrorInstanceNotReady"
	errorPollingLastOperationReason                string = "ErrorPollingLastOperation"
	errorWithOriginatingIdentity                   string = "Error with Originating Identity"
	errorReconciliationRetryTimeoutReason          string = "ErrorReconciliationRetryTimeout"
	errorServiceBindingOrphanMitigation string = "ServiceBindingNeedsOrphanMitigation"
	errorOrphanMigitationReason                    string = "OrphanMitigationFailed"

	successInjectedBindResultReason           string = "InjectedBindResult"
	successInjectedBindResultMessage          string = "Injected bind result"
	successDeprovisionReason                  string = "DeprovisionedSuccessfully"
	successDeprovisionMessage                 string = "The instance was deprovisioned successfully"
	successUpdateInstanceReason               string = "InstanceUpdatedSuccessfully"
	successUpdateInstanceMessage              string = "The instance was updated successfully"
	successProvisionReason                    string = "ProvisionedSuccessfully"
	successProvisionMessage                   string = "The instance was provisioned successfully"
	successFetchedCatalogReason               string = "FetchedCatalog"
	successFetchedCatalogMessage              string = "Successfully fetched catalog entries from broker."
	successClusterServiceBrokerDeletedReason  string = "DeletedSuccessfully"
	successClusterServiceBrokerDeletedMessage string = "The broker %v was deleted successfully."
	successUnboundReason                      string = "UnboundSuccessfully"
	successOrphanMitigationReason             string = "OrphanMitigationSuccessful"
	asyncProvisioningReason                   string = "Provisioning"
	asyncProvisioningMessage                  string = "The instance is being provisioned asynchronously"
	asyncUpdatingInstanceReason               string = "UpdatingInstance"
	asyncUpdatingInstanceMessage              string = "The instance is being updated asynchronously"
	asyncDeprovisioningReason                 string = "Deprovisioning"
	asyncDeprovisioningMessage                string = "The instance is being deprovisioned asynchronously"
	bindingInFlightReason                     string = "BindingRequestInFlight"
	bindingInFlightMessage                    string = "Binding request for ServiceBinding in-flight to Broker"
	unbindingInFlightReason                   string = "UnbindingRequestInFlight"
	unbindingInFlightMessage                  string = "Unbind request for ServiceBinding in-flight to Broker"
	provisioningInFlightReason                string = "ProvisionRequestInFlight"
	provisioningInFlightMessage               string = "Provision request for ServiceInstance in-flight to Broker"
	instanceUpdatingInFlightReason            string = "UpdateInstanceRequestInFlight"
	instanceUpdatingInFlightMessage           string = "Update request for ServiceInstance in-flight to Broker"
	deprovisioningInFlightReason              string = "DeprovisionRequestInFlight"
	deprovisioningInFlightMessage             string = "Deprovision request for ServiceInstance in-flight to Broker"
)

// shouldReconcileClusterServiceBroker determines whether a broker should be reconciled; it
// returns true unless the broker has a ready condition with status true and
// the controller's broker relist interval has not elapsed since the broker's
// ready condition became true, or if the broker's RelistBehavior is set to Manual.
func shouldReconcileClusterServiceBroker(broker *v1alpha1.ClusterServiceBroker, now time.Time) bool {
	if broker.Status.ReconciledGeneration != broker.Generation {
		// If the spec has changed, we should reconcile the broker.
		return true
	}
	if broker.DeletionTimestamp != nil || len(broker.Status.Conditions) == 0 {
		// If the deletion timestamp is set or the broker has no status
		// conditions, we should reconcile it.
		return true
	}

	// find the ready condition in the broker's status
	for _, condition := range broker.Status.Conditions {
		if condition.Type == v1alpha1.ServiceBrokerConditionReady {
			// The broker has a ready condition

			if condition.Status == v1alpha1.ConditionTrue {

				// The broker's ready condition has status true, meaning that
				// at some point, we successfully listed the broker's catalog.
				if broker.Spec.RelistBehavior == v1alpha1.ServiceBrokerRelistBehaviorManual {
					// If a broker is configured with RelistBehaviorManual, it should
					// ignore the Duration and only relist based on spec changes

					glog.V(10).Infof(
						"Not processing ClusterServiceBroker %v: RelistBehavior is set to Manual",
						broker.Name,
					)
					return false
				}

				if broker.Spec.RelistDuration == nil {
					glog.Errorf(
						"Unable to process ClusterServiceBroker %v: RelistBehavior is set to Duration with a nil RelistDuration value",
						broker.Name,
					)
					return false
				}

				// By default, the broker should relist if it has been longer than the
				// RelistDuration since the broker's ready condition became true.
				duration := broker.Spec.RelistDuration.Duration
				intervalPassed := now.After(condition.LastTransitionTime.Add(duration))
				if intervalPassed == false {
					glog.V(10).Infof(
						"Not processing ClusterServiceBroker %v because RelistDuration has not elapsed since the broker became ready",
						broker.Name,
					)
				}
				return intervalPassed
			}

			// The broker's ready condition wasn't true; we should try to re-
			// list the broker.
			return true
		}
	}

	// The broker didn't have a ready condition; we should reconcile it.
	return true
}

func (c *controller) reconcileClusterServiceBrokerKey(key string) error {
	broker, err := c.brokerLister.Get(key)
	if errors.IsNotFound(err) {
		glog.Infof("Not doing work for ClusterServiceBroker %v because it has been deleted", key)
		return nil
	}
	if err != nil {
		glog.Infof("Unable to retrieve ClusterServiceBroker %v from store: %v", key, err)
		return err
	}

	return c.reconcileClusterServiceBroker(broker)
}

// reconcileClusterServiceBroker is the control-loop that reconciles a Broker. An
// error is returned to indicate that the binding has not been fully
// processed and should be resubmitted at a later time.
func (c *controller) reconcileClusterServiceBroker(broker *v1alpha1.ClusterServiceBroker) error {
	glog.V(4).Infof("Processing ClusterServiceBroker %v", broker.Name)

	// * If the broker's ready condition is true and the RelistBehavior has been
	// set to Manual, do not reconcile it.
	// * If the broker's ready condition is true and the relist interval has not
	// elapsed, do not reconcile it.
	if !shouldReconcileClusterServiceBroker(broker, time.Now()) {
		return nil
	}

	if broker.DeletionTimestamp == nil { // Add or update
		authConfig, err := getAuthCredentialsFromClusterServiceBroker(c.kubeClient, broker)
		if err != nil {
			s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
			glog.Info(s)
			c.recorder.Event(broker, apiv1.EventTypeWarning, errorAuthCredentialsReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}

		clientConfig := NewClientConfigurationForBroker(broker, authConfig)

		glog.V(4).Infof("Creating client for ClusterServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL)
		brokerClient, err := c.brokerClientCreateFunc(clientConfig)
		if err != nil {
			s := fmt.Sprintf("Error creating client for broker %q: %s", broker.Name, err)
			glog.Info(s)
			c.recorder.Event(broker, apiv1.EventTypeWarning, errorAuthCredentialsReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}

		glog.V(4).Infof("Adding/Updating ClusterServiceBroker %v", broker.Name)
		now := metav1.Now()
		brokerCatalog, err := brokerClient.GetCatalog()
		if err != nil {
			s := fmt.Sprintf("Error getting broker catalog for broker %q: %s", broker.Name, err)
			glog.Warning(s)
			c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorFetchingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
				return err
			}
			if broker.Status.OperationStartTime == nil {
				clone, err := api.Scheme.DeepCopy(broker)
				if err == nil {
					toUpdate := clone.(*v1alpha1.ClusterServiceBroker)
					toUpdate.Status.OperationStartTime = &now
					if _, err := c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(toUpdate); err != nil {
						glog.Errorf("Error updating operation start time of ClusterServiceBroker %q: %v", broker.Name, err)
						return err
					}
				}
			} else if !time.Now().Before(broker.Status.OperationStartTime.Time.Add(c.reconciliationRetryDuration)) {
				s := fmt.Sprintf("Stopping reconciliation retries on ClusterServiceBroker %q because too much time has elapsed", broker.Name)
				glog.Info(s)
				c.recorder.Event(broker, apiv1.EventTypeWarning, errorReconciliationRetryTimeoutReason, s)
				clone, err := api.Scheme.DeepCopy(broker)
				if err == nil {
					toUpdate := clone.(*v1alpha1.ClusterServiceBroker)
					toUpdate.Status.OperationStartTime = nil
					toUpdate.Status.ReconciledGeneration = toUpdate.Generation
					if err := c.updateClusterServiceBrokerCondition(toUpdate,
						v1alpha1.ServiceBrokerConditionFailed,
						v1alpha1.ConditionTrue,
						errorReconciliationRetryTimeoutReason,
						s); err != nil {
						return err
					}
				}
				return nil
			}
			return err
		}
		glog.V(5).Infof("Successfully fetched %v catalog entries for ClusterServiceBroker %v", len(brokerCatalog.Services), broker.Name)

		if broker.Status.OperationStartTime != nil {
			clone, err := api.Scheme.DeepCopy(broker)
			if err != nil {
				return err
			}
			toUpdate := clone.(*v1alpha1.ClusterServiceBroker)
			toUpdate.Status.OperationStartTime = nil
			if _, err := c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(toUpdate); err != nil {
				glog.Errorf("Error updating operation start time of ClusterServiceBroker %q: %v", broker.Name, err)
				return err
			}
		}

		glog.V(4).Infof("Converting catalog response for ClusterServiceBroker %v into service-catalog API", broker.Name)
		serviceClasses, servicePlans, err := convertCatalog(brokerCatalog)
		if err != nil {
			s := fmt.Sprintf("Error converting catalog payload for broker %q to service-catalog API: %s", broker.Name, err)
			glog.Warning(s)
			c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorSyncingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}
		glog.V(5).Infof("Successfully converted catalog payload from ClusterServiceBroker %v to service-catalog API", broker.Name)

		if len(serviceClasses) == 0 {
			s := fmt.Sprintf("Error getting catalog payload for broker %q; received zero services; at least one service is required", broker.Name)
			glog.Warning(s)
			c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorSyncingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s); err != nil {
				return err
			}
			return stderrors.New(s)
		}

		for _, servicePlan := range servicePlans {
			glog.V(4).Infof("Reconciling servicePlan %q (broker %q)", servicePlan.Spec.ExternalName, broker.Name)
			if err := c.reconcileServicePlan(broker, servicePlan); err != nil {
				s := fmt.Sprintf(
					"Error reconciling servicePlan %q (broker %q): %s",
					servicePlan.Spec.ExternalName,
					broker.Name,
					err,
				)
				glog.Warning(s)
				c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorSyncingCatalogReason, s)
				c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s)
				return err
			}
			glog.V(5).Infof("Reconciled servicePlan %v (broker %v)", servicePlan.Spec.ExternalName, broker.Name)
		}

		for _, serviceClass := range serviceClasses {
			glog.V(4).Infof("Reconciling ClusterServiceClass %v (broker %v)", serviceClass.Spec.ExternalName, broker.Name)
			if err := c.reconcileClusterServiceClassFromClusterServiceBrokerCatalog(broker, serviceClass); err != nil {
				s := fmt.Sprintf(
					"Error reconciling ClusterServiceClass %q (broker %q): %s",
					serviceClass.Spec.ExternalName,
					broker.Name,
					err,
				)
				glog.Warning(s)
				c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorSyncingCatalogReason, s)
				if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s); err != nil {
					return err
				}
				return err
			}

			glog.V(5).Infof("Reconciled ClusterServiceClass %v (broker %v)", serviceClass.Spec.ExternalName, broker.Name)
		}

		if err := c.updateClusterServiceBrokerCondition(broker, v1alpha1.ServiceBrokerConditionReady, v1alpha1.ConditionTrue, successFetchedCatalogReason, successFetchedCatalogMessage); err != nil {
			return err
		}

		c.recorder.Event(broker, apiv1.EventTypeNormal, successFetchedCatalogReason, successFetchedCatalogMessage)

		return nil
	}

	// All updates not having a DeletingTimestamp will have been handled above
	// and returned early. If we reach this point, we're dealing with an update
	// that's actually a soft delete-- i.e. we have some finalization to do.
	if finalizers := sets.NewString(broker.Finalizers...); finalizers.Has(v1alpha1.FinalizerServiceCatalog) {
		glog.V(4).Infof("Finalizing ClusterServiceBroker %v", broker.Name)

		// Get ALL ServiceClasses. Remove those that reference this ClusterServiceBroker.
		svcClasses, err := c.serviceClassLister.List(labels.Everything())
		if err != nil {
			c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorListingClusterServiceClassesReason, "%v %v", errorListingClusterServiceClassesMessage, err)
			if err := c.updateClusterServiceBrokerCondition(
				broker,
				v1alpha1.ServiceBrokerConditionReady,
				v1alpha1.ConditionUnknown,
				errorListingClusterServiceClassesReason,
				errorListingClusterServiceClassesMessage,
			); err != nil {
				return err
			}
			return err
		}

		// Delete ClusterServiceClasses that are for THIS ClusterServiceBroker.
		// TODO: Change this to use field selectors.
		for _, svcClass := range svcClasses {
			if svcClass.Spec.ClusterServiceBrokerName == broker.Name {
				// TODO : implement and switch to filter
				// selectors. This is horrifyingly inefficient
				// otherwise with lots of plans.
				//
				// find all the service plans owned by this service class
				plans, err := c.servicePlanLister.List(labels.Everything())
				if err != nil {
					c.updateClusterServiceBrokerCondition(
						broker,
						v1alpha1.ServiceBrokerConditionReady,
						v1alpha1.ConditionUnknown,
						errorListingClusterServicePlansReason,
						errorListingClusterServicePlansMessage,
					)
					c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorListingClusterServicePlansReason, "%v %v", errorListingClusterServicePlansMessage, err)
					return err
				}
				for _, plan := range plans {
					// only process the plans this class owns
					if svcClass.Name == plan.Spec.ClusterServiceClassRef.Name {
						err := c.serviceCatalogClient.ClusterServicePlans().Delete(plan.Name, &metav1.DeleteOptions{})
						if err != nil && !errors.IsNotFound(err) {
							s := fmt.Sprintf(
								"Error deleting ServicePlan %q (ClusterServiceClass %q): %s",
								plan.Name,
								svcClass.Spec.ExternalName,
								err,
							)
							glog.Warning(s)
							c.updateClusterServiceBrokerCondition(
								broker,
								v1alpha1.ServiceBrokerConditionReady,
								v1alpha1.ConditionUnknown,
								errorDeletingClusterServicePlanMessage,
								errorDeletingClusterServicePlanReason+s,
							)
							c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorDeletingClusterServicePlanReason, "%v %v", errorDeletingClusterServicePlanMessage, s)
							return err
						}
					}
				}

				err = c.serviceCatalogClient.ClusterServiceClasses().Delete(svcClass.Name, &metav1.DeleteOptions{})
				if err != nil && !errors.IsNotFound(err) {
					s := fmt.Sprintf(
						"Error deleting ClusterServiceClass %q (ClusterServiceBroker %q): %s",
						svcClass.Spec.ExternalName,
						broker.Name,
						err,
					)
					glog.Warning(s)
					c.recorder.Eventf(broker, apiv1.EventTypeWarning, errorDeletingClusterServiceClassReason, "%v %v", errorDeletingClusterServiceClassMessage, s)
					if err := c.updateClusterServiceBrokerCondition(
						broker,
						v1alpha1.ServiceBrokerConditionReady,
						v1alpha1.ConditionUnknown,
						errorDeletingClusterServiceClassMessage,
						errorDeletingClusterServiceClassReason+s,
					); err != nil {
						return err
					}
					return err
				}
			}
		}

		if err := c.updateClusterServiceBrokerCondition(
			broker,
			v1alpha1.ServiceBrokerConditionReady,
			v1alpha1.ConditionFalse,
			successClusterServiceBrokerDeletedReason,
			"The broker was deleted successfully",
		); err != nil {
			return err
		}
		// Clear the finalizer
		finalizers.Delete(v1alpha1.FinalizerServiceCatalog)
		c.updateClusterServiceBrokerFinalizers(broker, finalizers.List())

		c.recorder.Eventf(broker, apiv1.EventTypeNormal, successClusterServiceBrokerDeletedReason, successClusterServiceBrokerDeletedMessage, broker.Name)
		glog.V(5).Infof("Successfully deleted ServiceBroker %v", broker.Name)
		return nil
	}

	return nil
}

// reconcileClusterServiceClassFromClusterServiceBrokerCatalog reconciles a ClusterServiceClass after the
// ClusterServiceBroker's catalog has been re-listed.
func (c *controller) reconcileClusterServiceClassFromClusterServiceBrokerCatalog(broker *v1alpha1.ClusterServiceBroker, serviceClass *v1alpha1.ClusterServiceClass) error {
	serviceClass.Spec.ClusterServiceBrokerName = broker.Name

	existingServiceClass, err := c.serviceClassLister.Get(serviceClass.Name)
	if errors.IsNotFound(err) {
		// An error returned from a lister Get call means that the object does
		// not exist.  Create a new ServiceClass.
		glog.V(5).Infof("Fresh ClusterServiceClass %q; creating", serviceClass.Spec.ExternalName)
		if _, err := c.serviceCatalogClient.ClusterServiceClasses().Create(serviceClass); err != nil {
			glog.Errorf("Error creating ClusterServiceClass %q from ClusterServiceBroker %q: %v", serviceClass.Spec.ExternalName, broker.Name, err)
			return err
		}

		return nil
	} else if err != nil {
		glog.Errorf("Error getting ClusterServiceClass %q: %v", serviceClass.Spec.ExternalName, err)
		return err
	}

	if existingServiceClass.Spec.ClusterServiceBrokerName != broker.Name {
		errMsg := fmt.Sprintf("ClusterServiceClass %q for ClusterServiceBroker %q already exists for Broker %q", serviceClass.Spec.ExternalName, broker.Name, existingServiceClass.Spec.ClusterServiceBrokerName)
		glog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if existingServiceClass.Spec.ExternalID != serviceClass.Spec.ExternalID {
		errMsg := fmt.Sprintf("ClusterServiceClass %q already exists with OSB guid %q, received different guid %q", serviceClass.Spec.ExternalName, existingServiceClass.Name, serviceClass.Name)
		glog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	glog.V(5).Infof("Found existing ClusterServiceClass %v; updating", serviceClass.Spec.ExternalName)

	// There was an existing service class -- project the update onto it and
	// update it.
	clone, err := api.Scheme.DeepCopy(existingServiceClass)
	if err != nil {
		return err
	}

	toUpdate := clone.(*v1alpha1.ClusterServiceClass)
	toUpdate.Spec.Bindable = serviceClass.Spec.Bindable
	toUpdate.Spec.PlanUpdatable = serviceClass.Spec.PlanUpdatable
	toUpdate.Spec.Tags = serviceClass.Spec.Tags
	toUpdate.Spec.Description = serviceClass.Spec.Description
	toUpdate.Spec.Requires = serviceClass.Spec.Requires

	if _, err := c.serviceCatalogClient.ClusterServiceClasses().Update(toUpdate); err != nil {
		glog.Errorf("Error updating ClusterServiceClass %q from ClusterServiceBroker %q: %v", serviceClass.Spec.ExternalName, broker.Name, err)
		return err
	}

	return nil
}

// reconcileServicePlan reconciles a ServicePlan after the
// ClusterServiceClass's catalog has been re-listed.
func (c *controller) reconcileServicePlan(broker *v1alpha1.ClusterServiceBroker, servicePlan *v1alpha1.ClusterServicePlan) error {
	servicePlan.Spec.ClusterServiceBrokerName = broker.Name

	existingServicePlan, err := c.servicePlanLister.Get(servicePlan.Name)
	if errors.IsNotFound(err) {
		// An error returned from a lister Get call means that the object does
		// not exist.  Create a new ServicePlan.
		if _, err := c.serviceCatalogClient.ClusterServicePlans().Create(servicePlan); err != nil {
			glog.Errorf("Error creating servicePlan %q: %v", servicePlan.Name, err)
			return err
		}

		return nil
	} else if err != nil {
		glog.Errorf("Error getting servicePlan %q: %v", servicePlan.Name, err)
		return err
	}

	if existingServicePlan.Spec.ExternalID != servicePlan.Spec.ExternalID {
		errMsg := fmt.Sprintf("ServicePlan %q already exists with OSB guid %q, received different guid %q", servicePlan.Name, existingServicePlan.Spec.ExternalID, servicePlan.Spec.ExternalID)
		glog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	glog.V(5).Infof("Found existing servicePlan %q; updating", servicePlan.Name)

	// There was an existing service plan -- project the update onto it and
	// update it.
	clone, err := api.Scheme.DeepCopy(existingServicePlan)
	if err != nil {
		return err
	}

	toUpdate := clone.(*v1alpha1.ClusterServicePlan)
	toUpdate.Spec.Description = servicePlan.Spec.Description
	toUpdate.Spec.Bindable = servicePlan.Spec.Bindable
	toUpdate.Spec.Free = servicePlan.Spec.Free
	toUpdate.Spec.ServiceInstanceCreateParameterSchema = servicePlan.Spec.ServiceInstanceCreateParameterSchema
	toUpdate.Spec.ServiceInstanceUpdateParameterSchema = servicePlan.Spec.ServiceInstanceUpdateParameterSchema
	toUpdate.Spec.ServiceBindingCreateParameterSchema = servicePlan.Spec.ServiceBindingCreateParameterSchema

	if _, err := c.serviceCatalogClient.ClusterServicePlans().Update(toUpdate); err != nil {
		glog.Errorf("Error updating ServicePlan %q: %v", servicePlan.Name, err)
		return err
	}

	return nil
}

// updateClusterServiceBrokerCondition updates the ready condition for the given Broker
// with the given status, reason, and message.
func (c *controller) updateClusterServiceBrokerCondition(broker *v1alpha1.ClusterServiceBroker, conditionType v1alpha1.ServiceBrokerConditionType, status v1alpha1.ConditionStatus, reason, message string) error {
	clone, err := api.Scheme.DeepCopy(broker)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.ClusterServiceBroker)
	newCondition := v1alpha1.ServiceBrokerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	t := time.Now()

	if len(broker.Status.Conditions) == 0 {
		glog.Infof("Setting lastTransitionTime for ClusterServiceBroker %q condition %q to %v", broker.Name, conditionType, t)
		newCondition.LastTransitionTime = metav1.NewTime(t)
		toUpdate.Status.Conditions = []v1alpha1.ServiceBrokerCondition{newCondition}
	} else {
		for i, cond := range broker.Status.Conditions {
			if cond.Type == conditionType {
				if cond.Status != newCondition.Status {
					glog.Infof("Found status change for ClusterServiceBroker %q condition %q: %q -> %q; setting lastTransitionTime to %v", broker.Name, conditionType, cond.Status, status, t)
					newCondition.LastTransitionTime = metav1.NewTime(t)
				} else {
					newCondition.LastTransitionTime = cond.LastTransitionTime
				}

				toUpdate.Status.Conditions[i] = newCondition
				break
			}
		}
	}

	// Set status.ReconciledGeneration if updating ready condition to true

	if conditionType == v1alpha1.ServiceBrokerConditionReady && status == v1alpha1.ConditionTrue {
		toUpdate.Status.ReconciledGeneration = toUpdate.Generation
	}

	glog.V(4).Infof("Updating ready condition for ClusterServiceBroker %v to %v", broker.Name, status)
	_, err = c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating ready condition for ClusterServiceBroker %q: %v", broker.Name, err)
	} else {
		glog.V(5).Infof("Updated ready condition for ClusterServiceBroker %q to %v", broker.Name, status)
	}

	return err
}

// updateClusterServiceBrokerFinalizers updates the given finalizers for the given Broker.
func (c *controller) updateClusterServiceBrokerFinalizers(
	broker *v1alpha1.ClusterServiceBroker,
	finalizers []string) error {

	// Get the latest version of the broker so that we can avoid conflicts
	// (since we have probably just updated the status of the broker and are
	// now removing the last finalizer).
	broker, err := c.serviceCatalogClient.ClusterServiceBrokers().Get(broker.Name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Error getting ClusterServiceBroker %q to finalize: %v", broker.Name, err)
	}

	clone, err := api.Scheme.DeepCopy(broker)
	if err != nil {
		return err
	}
	toUpdate := clone.(*v1alpha1.ClusterServiceBroker)

	toUpdate.Finalizers = finalizers

	logContext := fmt.Sprintf("finalizers for ClusterServiceBroker %v to %v",
		broker.Name, finalizers)

	glog.V(4).Infof("Updating %v", logContext)
	_, err = c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(toUpdate)
	if err != nil {
		glog.Errorf("Error updating %v: %v", logContext, err)
	}
	return err
}
