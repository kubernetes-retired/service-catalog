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
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/metrics"
	"github.com/kubernetes-sigs/service-catalog/pkg/pretty"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
)

// the Message strings have a terminating period and space so they can
// be easily combined with a follow on specific message.
const (
	errorListingClusterServiceClassesReason  string = "ErrorListingClusterServiceClasses"
	errorListingClusterServiceClassesMessage string = "Error listing cluster service classes."
	errorListingClusterServicePlansReason    string = "ErrorListingClusterServicePlans"
	errorListingClusterServicePlansMessage   string = "Error listing cluster service plans."
	errorDeletingClusterServiceClassReason   string = "ErrorDeletingClusterServiceClass"
	errorDeletingClusterServiceClassMessage  string = "Error deleting cluster service class."
	errorDeletingClusterServicePlanReason    string = "ErrorDeletingClusterServicePlan"
	errorDeletingClusterServicePlanMessage   string = "Error deleting cluster service plan."
	errorAuthCredentialsReason               string = "ErrorGettingAuthCredentials"

	successClusterServiceBrokerDeletedReason  string = "DeletedClusterServiceBrokerSuccessfully"
	successClusterServiceBrokerDeletedMessage string = "The broker %v was deleted successfully."

	// these reasons are re-used in other controller files.
	errorFetchingCatalogReason            string = "ErrorFetchingCatalog"
	errorFetchingCatalogMessage           string = "Error fetching catalog."
	errorSyncingCatalogReason             string = "ErrorSyncingCatalog"
	errorSyncingCatalogMessage            string = "Error syncing catalog from ClusterServiceBroker."
	successFetchedCatalogReason           string = "FetchedCatalog"
	successFetchedCatalogMessage          string = "Successfully fetched catalog entries from broker."
	errorReconciliationRetryTimeoutReason string = "ErrorReconciliationRetryTimeout"
)

func (c *controller) clusterServiceBrokerAdd(obj interface{}) {
	// DeletionHandlingMetaNamespaceKeyFunc returns a unique key for the resource and
	// handles the special case where the resource is of DeletedFinalStateUnknown type, which
	// acts a place holder for resources that have been deleted from storage but the watch event
	// confirming the deletion has not yet arrived.
	// Generally, the key is "namespace/name" for namespaced-scoped resources and
	// just "name" for cluster scoped resources.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	c.clusterServiceBrokerQueue.Add(key)
}

func (c *controller) clusterServiceBrokerUpdate(oldObj, newObj interface{}) {
	c.clusterServiceBrokerAdd(newObj)
}

func (c *controller) clusterServiceBrokerDelete(obj interface{}) {
	broker, ok := obj.(*v1beta1.ClusterServiceBroker)
	if broker == nil || !ok {
		return
	}

	klog.V(4).Infof("Received delete event for ClusterServiceBroker %v; no further processing will occur", broker.Name)
}

// shouldReconcileClusterServiceBroker determines whether a broker should be reconciled; it
// returns true unless the broker has a ready condition with status true and
// the controller's broker relist interval has not elapsed since the broker's
// ready condition became true, or if the broker's RelistBehavior is set to Manual.
func shouldReconcileClusterServiceBroker(broker *v1beta1.ClusterServiceBroker, now time.Time, defaultRelistInterval time.Duration) bool {
	return shouldReconcileServiceBrokerCommon(
		pretty.NewClusterServiceBrokerContextBuilder(broker),
		&broker.ObjectMeta,
		&broker.Spec.CommonServiceBrokerSpec,
		&broker.Status.CommonServiceBrokerStatus,
		now,
		defaultRelistInterval,
	)
}

func (c *controller) reconcileClusterServiceBrokerKey(key string) error {
	broker, err := c.clusterServiceBrokerLister.Get(key)
	pcb := pretty.NewContextBuilder(pretty.ClusterServiceBroker, "", key, "")

	klog.V(4).Info(pcb.Message("Processing service broker"))

	if errors.IsNotFound(err) {
		klog.Info(pcb.Message("Not doing work because it has been deleted"))
		c.brokerClientManager.RemoveBrokerClient(NewClusterServiceBrokerKey(key))
		return nil
	}
	if err != nil {
		klog.Info(pcb.Messagef("Unable to retrieve object from store: %v", err))
		return err
	}

	return c.reconcileClusterServiceBroker(broker)
}

func (c *controller) clusterServiceBrokerClient(broker *v1beta1.ClusterServiceBroker) (osb.Client, error) {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)
	klog.V(4).Info(pcb.Message("Updating broker client"))
	authConfig, err := c.getAuthCredentialsFromClusterServiceBroker(broker)
	if err != nil {
		s := fmt.Sprintf("Error getting broker auth credentials: %s", err)
		klog.Info(pcb.Message(s))
		c.recorder.Event(broker, corev1.EventTypeWarning, errorAuthCredentialsReason, s)
		if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
			return nil, err
		}
		return nil, err
	}
	clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig, c.OSBAPITimeOut)
	brokerClient, err := c.brokerClientManager.UpdateBrokerClient(NewClusterServiceBrokerKey(broker.Name), clientConfig)
	if err != nil {
		s := fmt.Sprintf("Error creating client for broker %q: %s", broker.Name, err)
		klog.Info(pcb.Message(s))
		c.recorder.Event(broker, corev1.EventTypeWarning, errorAuthCredentialsReason, s)
		if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
			return nil, err
		}
		return nil, err
	}
	return brokerClient, nil
}

// reconcileClusterServiceBroker is the control-loop that reconciles a Broker. An
// error is returned to indicate that the binding has not been fully
// processed and should be resubmitted at a later time.
func (c *controller) reconcileClusterServiceBroker(broker *v1beta1.ClusterServiceBroker) error {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)
	klog.V(4).Infof(pcb.Message("Processing"))

	// * If the broker's ready condition is true and the RelistBehavior has been
	// set to Manual, do not reconcile it.
	// * If the broker's ready condition is true and the relist interval has not
	// elapsed, do not reconcile it.
	if !shouldReconcileClusterServiceBroker(broker, time.Now(), c.brokerRelistInterval) {
		return nil
	}

	if broker.DeletionTimestamp == nil { // Add or update
		klog.V(4).Info(pcb.Message("Processing adding/update event"))

		brokerClient, err := c.clusterServiceBrokerClient(broker)
		if err != nil {
			return err
		}

		// get the broker's catalog
		now := metav1.Now()
		brokerCatalog, err := brokerClient.GetCatalog()
		if err != nil {
			s := fmt.Sprintf("Error getting broker catalog: %s", err)
			klog.Warning(pcb.Message(s))
			c.recorder.Eventf(broker, corev1.EventTypeWarning, errorFetchingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorFetchingCatalogReason, errorFetchingCatalogMessage+s); err != nil {
				return err
			}
			if broker.Status.OperationStartTime == nil {
				toUpdate := broker.DeepCopy()
				toUpdate.Status.OperationStartTime = &now
				if _, err := c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(context.Background(), toUpdate, metav1.UpdateOptions{}); err != nil {
					klog.Error(pcb.Messagef("Error updating operation start time: %v", err))
					return err
				}
			} else if !time.Now().Before(broker.Status.OperationStartTime.Time.Add(c.reconciliationRetryDuration)) {
				s := "Stopping reconciliation retries because too much time has elapsed"
				klog.Info(pcb.Message(s))
				c.recorder.Event(broker, corev1.EventTypeWarning, errorReconciliationRetryTimeoutReason, s)
				toUpdate := broker.DeepCopy()
				toUpdate.Status.OperationStartTime = nil
				toUpdate.Status.ReconciledGeneration = toUpdate.Generation
				return c.updateClusterServiceBrokerCondition(toUpdate,
					v1beta1.ServiceBrokerConditionFailed,
					v1beta1.ConditionTrue,
					errorReconciliationRetryTimeoutReason,
					s)
			}
			return err
		}

		klog.V(5).Info(pcb.Messagef("Successfully fetched %v catalog entries", len(brokerCatalog.Services)))

		// set the operation start time if not already set
		if broker.Status.OperationStartTime != nil {
			toUpdate := broker.DeepCopy()
			toUpdate.Status.OperationStartTime = nil
			updated, err := c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(context.Background(), toUpdate, metav1.UpdateOptions{})
			if err != nil {
				klog.Error(pcb.Messagef("Error updating operation start time: %v", err))
				return err
			}
			broker = updated
		}

		// get the existing services and plans for this broker so that we can
		// detect when services and plans are removed from the broker's
		// catalog
		existingServiceClasses, existingServicePlans, err := c.getCurrentServiceClassesAndPlansForBroker(broker)
		if err != nil {
			return err
		}

		existingServiceClassMap := convertClusterServiceClassListToMap(existingServiceClasses)
		existingServicePlanMap := convertClusterServicePlanListToMap(existingServicePlans)

		// convert the broker's catalog payload into our API objects
		klog.V(4).Info(pcb.Message("Converting catalog response into service-catalog API"))
		payloadServiceClasses, payloadServicePlans, err := convertAndFilterCatalog(brokerCatalog, broker.Spec.CatalogRestrictions, existingServiceClassMap, existingServicePlanMap)
		if err != nil {
			s := fmt.Sprintf("Error converting catalog payload for broker %q to service-catalog API: %s", broker.Name, err)
			klog.Warning(pcb.Message(s))
			c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}
		klog.V(5).Info(pcb.Message("Successfully converted catalog payload from to service-catalog API"))

		// reconcile the serviceClasses that were part of the broker's catalog
		// payload
		for _, payloadServiceClass := range payloadServiceClasses {
			existingServiceClass, _ := existingServiceClassMap[payloadServiceClass.Name]
			delete(existingServiceClassMap, payloadServiceClass.Name)
			if existingServiceClass == nil {
				existingServiceClass, _ = existingServiceClassMap[payloadServiceClass.Spec.ExternalID]
				delete(existingServiceClassMap, payloadServiceClass.Spec.ExternalID)
			}

			klog.V(4).Info(pcb.Messagef("Reconciling %s", pretty.ClusterServiceClassName(payloadServiceClass)))
			if err := c.reconcileClusterServiceClassFromClusterServiceBrokerCatalog(broker, payloadServiceClass, existingServiceClass); err != nil {
				s := fmt.Sprintf(
					"Error reconciling %s (broker %q): %s",
					pretty.ClusterServiceClassName(payloadServiceClass), broker.Name, err,
				)
				klog.Warning(pcb.Message(s))
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
				if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s); err != nil {
					return err
				}
				return err
			}

			klog.V(5).Info(pcb.Messagef("Reconciled %s", pretty.ClusterServiceClassName(payloadServiceClass)))
		}

		// handle the serviceClasses that were not in the broker's payload;
		// mark these as having been removed from the broker's catalog
		for _, existingServiceClass := range existingServiceClassMap {
			if existingServiceClass.Status.RemovedFromBrokerCatalog {
				continue
			}

			// Do not delete user-defined classes
			if !isServiceCatalogManagedResource(existingServiceClass) {
				continue
			}

			klog.V(4).Info(pcb.Messagef("%s has been removed from broker's catalog; marking", pretty.ClusterServiceClassName(existingServiceClass)))
			existingServiceClass.Status.RemovedFromBrokerCatalog = true
			_, err := c.serviceCatalogClient.ClusterServiceClasses().UpdateStatus(context.Background(), existingServiceClass, metav1.UpdateOptions{})
			if err != nil {
				s := fmt.Sprintf(
					"Error updating status of %s: %v",
					pretty.ClusterServiceClassName(existingServiceClass), err,
				)
				klog.Warning(pcb.Message(s))
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
				if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s); err != nil {
					return err
				}
				return err
			}
		}

		// reconcile the plans that were part of the broker's catalog payload
		for _, payloadServicePlan := range payloadServicePlans {
			existingServicePlan, _ := existingServicePlanMap[payloadServicePlan.Name]
			delete(existingServicePlanMap, payloadServicePlan.Name)
			if existingServicePlan == nil {
				existingServicePlan, _ = existingServicePlanMap[payloadServicePlan.Spec.ExternalID]
				delete(existingServicePlanMap, payloadServicePlan.Spec.ExternalID)
			}

			klog.V(4).Infof(
				"ClusterServiceBroker %q: reconciling %s",
				broker.Name, pretty.ClusterServicePlanName(payloadServicePlan),
			)
			if err := c.reconcileClusterServicePlanFromClusterServiceBrokerCatalog(broker, payloadServicePlan, existingServicePlan); err != nil {
				s := fmt.Sprintf(
					"Error reconciling %s: %s",
					pretty.ClusterServicePlanName(payloadServicePlan), err,
				)
				klog.Warning(pcb.Message(s))
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
				c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s)
				return err
			}
			klog.V(5).Info(pcb.Messagef("Reconciled %s", pretty.ClusterServicePlanName(payloadServicePlan)))

		}

		// handle the servicePlans that were not in the broker's payload;
		// mark these as deleted
		for _, existingServicePlan := range existingServicePlanMap {
			if existingServicePlan.Status.RemovedFromBrokerCatalog {
				continue
			}

			// Do not delete user-defined plans
			if !isServiceCatalogManagedResource(existingServicePlan) {
				continue
			}

			klog.V(4).Info(pcb.Messagef("%s has been removed from broker's catalog; marking", pretty.ClusterServicePlanName(existingServicePlan)))
			existingServicePlan.Status.RemovedFromBrokerCatalog = true
			_, err := c.serviceCatalogClient.ClusterServicePlans().UpdateStatus(context.Background(), existingServicePlan, metav1.UpdateOptions{})
			if err != nil {
				s := fmt.Sprintf(
					"Error updating status of %s: %v",
					pretty.ClusterServicePlanName(existingServicePlan),
					err,
				)
				klog.Warning(pcb.Message(s))
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
				if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason,
					errorSyncingCatalogMessage+s); err != nil {
					return err
				}
				return err
			}
		}

		// everything worked correctly; update the broker's ready condition to
		// status true
		if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionTrue, successFetchedCatalogReason, successFetchedCatalogMessage); err != nil {
			return err
		}

		c.recorder.Event(broker, corev1.EventTypeNormal, successFetchedCatalogReason, successFetchedCatalogMessage)

		// Update metrics with the number of serviceclasses and serviceplans from this broker
		metrics.BrokerServiceClassCount.WithLabelValues(broker.Name, "").Set(float64(len(payloadServiceClasses)))
		metrics.BrokerServicePlanCount.WithLabelValues(broker.Name, "").Set(float64(len(payloadServicePlans)))

		return nil
	}

	// All updates not having a DeletingTimestamp will have been handled above
	// and returned early. If we reach this point, we're dealing with an update
	// that's actually a soft delete-- i.e. we have some finalization to do.
	if finalizers := sets.NewString(broker.Finalizers...); finalizers.Has(v1beta1.FinalizerServiceCatalog) {
		klog.V(4).Info(pcb.Message("Finalizing"))

		existingServiceClasses, existingServicePlans, err := c.getCurrentServiceClassesAndPlansForBroker(broker)
		if err != nil {
			return err
		}

		klog.V(4).Info(pcb.Messagef("Found %d ClusterServiceClasses and %d ClusterServicePlans to delete", len(existingServiceClasses), len(existingServicePlans)))

		for _, plan := range existingServicePlans {
			klog.V(4).Info(pcb.Messagef("Deleting %s", pretty.ClusterServicePlanName(&plan)))
			err := c.serviceCatalogClient.ClusterServicePlans().Delete(context.Background(), plan.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				s := fmt.Sprintf("Error deleting %s: %s", pretty.ClusterServicePlanName(&plan), err)
				klog.Warning(pcb.Message(s))
				c.updateClusterServiceBrokerCondition(
					broker,
					v1beta1.ServiceBrokerConditionReady,
					v1beta1.ConditionUnknown,
					errorDeletingClusterServicePlanMessage,
					errorDeletingClusterServicePlanReason+s,
				)
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorDeletingClusterServicePlanReason, "%v %v", errorDeletingClusterServicePlanMessage, s)
				return err
			}
		}

		for _, svcClass := range existingServiceClasses {
			klog.V(4).Info(pcb.Messagef("Deleting %s", pretty.ClusterServiceClassName(&svcClass)))
			err = c.serviceCatalogClient.ClusterServiceClasses().Delete(context.Background(), svcClass.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				s := fmt.Sprintf("Error deleting %s: %s", pretty.ClusterServiceClassName(&svcClass), err)
				klog.Warning(pcb.Message(s))
				c.recorder.Eventf(broker, corev1.EventTypeWarning, errorDeletingClusterServiceClassReason, "%v %v", errorDeletingClusterServiceClassMessage, s)
				if err := c.updateClusterServiceBrokerCondition(
					broker,
					v1beta1.ServiceBrokerConditionReady,
					v1beta1.ConditionUnknown,
					errorDeletingClusterServiceClassMessage,
					errorDeletingClusterServiceClassReason+s,
				); err != nil {
					return err
				}
				return err
			}
		}

		if err := c.updateClusterServiceBrokerCondition(
			broker,
			v1beta1.ServiceBrokerConditionReady,
			v1beta1.ConditionFalse,
			successClusterServiceBrokerDeletedReason,
			"The broker was deleted successfully",
		); err != nil {
			return err
		}
		// Clear the finalizer
		finalizers.Delete(v1beta1.FinalizerServiceCatalog)
		c.updateClusterServiceBrokerFinalizers(broker, finalizers.List())

		c.recorder.Eventf(broker, corev1.EventTypeNormal, successClusterServiceBrokerDeletedReason, successClusterServiceBrokerDeletedMessage, broker.Name)
		klog.V(5).Info(pcb.Message("Successfully deleted"))

		// delete the metrics associated with this broker
		metrics.BrokerServiceClassCount.DeleteLabelValues(broker.Name, "")
		metrics.BrokerServicePlanCount.DeleteLabelValues(broker.Name, "")
		return nil
	}

	return nil
}

// reconcileClusterServiceClassFromClusterServiceBrokerCatalog reconciles a
// ClusterServiceClass after the ClusterServiceBroker's catalog has been re-
// listed. The serviceClass parameter is the serviceClass from the broker's
// catalog payload. The existingServiceClass parameter is the serviceClass
// that already exists for the given broker with this serviceClass' k8s name.
func (c *controller) reconcileClusterServiceClassFromClusterServiceBrokerCatalog(broker *v1beta1.ClusterServiceBroker, serviceClass, existingServiceClass *v1beta1.ClusterServiceClass) error {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)
	serviceClass.Spec.ClusterServiceBrokerName = broker.Name

	if existingServiceClass == nil {
		otherServiceClass, err := c.clusterServiceClassLister.Get(serviceClass.Name)
		if err != nil {
			// we expect _not_ to find a service class this way, so a not-
			// found error is expected and legitimate.
			if !errors.IsNotFound(err) {
				return err
			}
		} else {
			// we do not expect to find an existing service class if we were
			// not already passed one; the following if statement will almost
			// certainly evaluate to true.
			if otherServiceClass.Spec.ClusterServiceBrokerName != broker.Name {
				errMsg := fmt.Sprintf("%s already exists for Broker %q",
					pretty.ClusterServiceClassName(serviceClass), otherServiceClass.Spec.ClusterServiceBrokerName,
				)
				klog.Error(pcb.Message(errMsg))
				return fmt.Errorf(errMsg)
			}
		}

		markAsServiceCatalogManagedResource(serviceClass, broker)

		klog.V(5).Info(pcb.Messagef("Fresh %s; creating", pretty.ClusterServiceClassName(serviceClass)))
		if _, err := c.serviceCatalogClient.ClusterServiceClasses().Create(context.Background(), serviceClass, metav1.CreateOptions{}); err != nil {
			klog.Error(pcb.Messagef("Error creating %s: %v", pretty.ClusterServiceClassName(serviceClass), err))
			return err
		}

		return nil
	}

	if existingServiceClass.Spec.ExternalID != serviceClass.Spec.ExternalID {
		errMsg := fmt.Sprintf(
			"%s already exists with OSB guid %q, received different guid %q",
			pretty.ClusterServiceClassName(serviceClass), existingServiceClass.Name, serviceClass.Name,
		)
		klog.Error(pcb.Message(errMsg))
		return fmt.Errorf(errMsg)
	}

	klog.V(5).Info(pcb.Messagef("Found existing %s; updating", pretty.ClusterServiceClassName(serviceClass)))

	// There was an existing service class -- project the update onto it and
	// update it.
	toUpdate := existingServiceClass.DeepCopy()
	toUpdate.Spec.BindingRetrievable = serviceClass.Spec.BindingRetrievable
	toUpdate.Spec.Bindable = serviceClass.Spec.Bindable
	toUpdate.Spec.PlanUpdatable = serviceClass.Spec.PlanUpdatable
	toUpdate.Spec.Tags = serviceClass.Spec.Tags
	toUpdate.Spec.Description = serviceClass.Spec.Description
	toUpdate.Spec.Requires = serviceClass.Spec.Requires
	toUpdate.Spec.ExternalName = serviceClass.Spec.ExternalName
	toUpdate.Spec.ExternalMetadata = serviceClass.Spec.ExternalMetadata

	markAsServiceCatalogManagedResource(toUpdate, broker)

	updatedServiceClass, err := c.serviceCatalogClient.ClusterServiceClasses().Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(pcb.Messagef("Error updating %s: %v", pretty.ClusterServiceClassName(serviceClass), err))
		return err
	}

	if updatedServiceClass.Status.RemovedFromBrokerCatalog {
		klog.V(4).Info(pcb.Messagef("Resetting RemovedFromBrokerCatalog status on %s", pretty.ClusterServiceClassName(serviceClass)))
		updatedServiceClass.Status.RemovedFromBrokerCatalog = false
		_, err := c.serviceCatalogClient.ClusterServiceClasses().UpdateStatus(context.Background(), updatedServiceClass, metav1.UpdateOptions{})
		if err != nil {
			s := fmt.Sprintf("Error updating status of %s: %v", pretty.ClusterServiceClassName(updatedServiceClass), err)
			klog.Warning(pcb.Message(s))
			c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}
	}

	return nil
}

// reconcileClusterServicePlanFromClusterServiceBrokerCatalog reconciles a
// ServicePlan after the ServiceClass's catalog has been re-listed.
func (c *controller) reconcileClusterServicePlanFromClusterServiceBrokerCatalog(broker *v1beta1.ClusterServiceBroker, servicePlan, existingServicePlan *v1beta1.ClusterServicePlan) error {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)
	servicePlan.Spec.ClusterServiceBrokerName = broker.Name

	if existingServicePlan == nil {
		otherServicePlan, err := c.clusterServicePlanLister.Get(servicePlan.Name)
		if err != nil {
			// we expect _not_ to find a service class this way, so a not-
			// found error is expected and legitimate.
			if !errors.IsNotFound(err) {
				return err
			}
		} else {
			// we do not expect to find an existing service class if we were
			// not already passed one; the following if statement will almost
			// certainly evaluate to true.
			if otherServicePlan.Spec.ClusterServiceBrokerName != broker.Name {
				errMsg := fmt.Sprintf(
					"%s already exists for Broker %q",
					pretty.ClusterServicePlanName(servicePlan), otherServicePlan.Spec.ClusterServiceBrokerName,
				)
				klog.Error(pcb.Message(errMsg))
				return fmt.Errorf(errMsg)
			}
		}

		markAsServiceCatalogManagedResource(servicePlan, broker)

		// An error returned from a lister Get call means that the object does
		// not exist.  Create a new ClusterServicePlan.
		if _, err := c.serviceCatalogClient.ClusterServicePlans().Create(context.Background(), servicePlan, metav1.CreateOptions{}); err != nil {
			klog.Error(pcb.Messagef("Error creating %s: %v", pretty.ClusterServicePlanName(servicePlan), err))
			return err
		}

		return nil
	}

	if existingServicePlan.Spec.ExternalID != servicePlan.Spec.ExternalID {
		errMsg := fmt.Sprintf(
			"%s already exists with OSB guid %q, received different guid %q",
			pretty.ClusterServicePlanName(servicePlan), existingServicePlan.Spec.ExternalID, servicePlan.Spec.ExternalID,
		)
		klog.Error(pcb.Message(errMsg))
		return fmt.Errorf(errMsg)
	}

	klog.V(5).Info(pcb.Messagef("Found existing %s; updating", pretty.ClusterServicePlanName(servicePlan)))

	// There was an existing service plan -- project the update onto it and
	// update it.
	toUpdate := existingServicePlan.DeepCopy()
	toUpdate.Spec.Description = servicePlan.Spec.Description
	toUpdate.Spec.Bindable = servicePlan.Spec.Bindable
	toUpdate.Spec.Free = servicePlan.Spec.Free
	toUpdate.Spec.ExternalName = servicePlan.Spec.ExternalName
	toUpdate.Spec.ExternalMetadata = servicePlan.Spec.ExternalMetadata
	toUpdate.Spec.InstanceCreateParameterSchema = servicePlan.Spec.InstanceCreateParameterSchema
	toUpdate.Spec.InstanceUpdateParameterSchema = servicePlan.Spec.InstanceUpdateParameterSchema
	toUpdate.Spec.ServiceBindingCreateParameterSchema = servicePlan.Spec.ServiceBindingCreateParameterSchema

	markAsServiceCatalogManagedResource(toUpdate, broker)

	updatedPlan, err := c.serviceCatalogClient.ClusterServicePlans().Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(pcb.Messagef("Error updating %s: %v", pretty.ClusterServicePlanName(servicePlan), err))
		return err
	}

	if updatedPlan.Status.RemovedFromBrokerCatalog {
		updatedPlan.Status.RemovedFromBrokerCatalog = false
		klog.V(4).Info(pcb.Messagef("Resetting RemovedFromBrokerCatalog status on %s", pretty.ClusterServicePlanName(updatedPlan)))

		_, err := c.serviceCatalogClient.ClusterServicePlans().UpdateStatus(context.Background(), updatedPlan, metav1.UpdateOptions{})
		if err != nil {
			s := fmt.Sprintf("Error updating status of %s: %v", pretty.ClusterServicePlanName(updatedPlan), err)
			klog.Error(pcb.Message(s))
			c.recorder.Eventf(broker, corev1.EventTypeWarning, errorSyncingCatalogReason, s)
			if err := c.updateClusterServiceBrokerCondition(broker, v1beta1.ServiceBrokerConditionReady, v1beta1.ConditionFalse, errorSyncingCatalogReason, errorSyncingCatalogMessage+s); err != nil {
				return err
			}
			return err
		}
	}

	return nil
}

// updateClusterServiceBrokerCondition updates the ready condition for the given Broker
// with the given status, reason, and message.
func (c *controller) updateClusterServiceBrokerCondition(broker *v1beta1.ClusterServiceBroker, conditionType v1beta1.ServiceBrokerConditionType, status v1beta1.ConditionStatus, reason, message string) error {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)
	toUpdate := broker.DeepCopy()
	newCondition := v1beta1.ServiceBrokerCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	t := time.Now()

	if len(broker.Status.Conditions) == 0 {
		klog.Info(pcb.Messagef("Setting lastTransitionTime for condition %q to %v", conditionType, t))
		newCondition.LastTransitionTime = metav1.NewTime(t)
		toUpdate.Status.Conditions = []v1beta1.ServiceBrokerCondition{newCondition}
	} else {
		for i, cond := range broker.Status.Conditions {
			if cond.Type == conditionType {
				if cond.Status != newCondition.Status {
					klog.Info(pcb.Messagef(
						"Found status change for condition %q: %q -> %q; setting lastTransitionTime to %v",
						conditionType, cond.Status, status, t,
					))
					newCondition.LastTransitionTime = metav1.NewTime(t)
				} else {
					newCondition.LastTransitionTime = cond.LastTransitionTime
				}

				toUpdate.Status.Conditions[i] = newCondition
				break
			}
		}
	}

	// Set status.ReconciledGeneration && status.LastCatalogRetrievalTime if updating ready condition to true

	if conditionType == v1beta1.ServiceBrokerConditionReady && status == v1beta1.ConditionTrue {
		toUpdate.Status.ReconciledGeneration = toUpdate.Generation
		now := metav1.NewTime(t)
		toUpdate.Status.LastCatalogRetrievalTime = &now
	}
	toUpdate.RecalculatePrinterColumnStatusFields()

	klog.V(4).Info(pcb.Messagef("Updating ready condition to %v", status))
	_, err := c.serviceCatalogClient.ClusterServiceBrokers().UpdateStatus(context.Background(), toUpdate, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(pcb.Messagef("Error updating ready condition: %v", err))
	} else {
		klog.V(5).Info(pcb.Messagef("Updated ready condition to %v", status))
	}

	return err
}

// updateClusterServiceBrokerFinalizers updates the given finalizers for the given Broker.
func (c *controller) updateClusterServiceBrokerFinalizers(
	broker *v1beta1.ClusterServiceBroker,
	finalizers []string) error {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)

	// Get the latest version of the broker so that we can avoid conflicts
	// (since we have probably just updated the status of the broker and are
	// now removing the last finalizer).
	broker, err := c.serviceCatalogClient.ClusterServiceBrokers().Get(context.Background(), broker.Name, metav1.GetOptions{})
	if err != nil {
		klog.Error(pcb.Messagef("Error finalizing: %v", err))
	}

	toUpdate := broker.DeepCopy()
	toUpdate.Finalizers = finalizers

	logContext := fmt.Sprint(pcb.Messagef("Updating finalizers to %v", finalizers))

	klog.V(4).Info(pcb.Messagef("Updating %v", logContext))
	_, err = c.serviceCatalogClient.ClusterServiceBrokers().Update(context.Background(), toUpdate, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(pcb.Messagef("Error updating %v: %v", logContext, err))
	}
	return err
}

func (c *controller) getCurrentServiceClassesAndPlansForBroker(broker *v1beta1.ClusterServiceBroker) ([]v1beta1.ClusterServiceClass, []v1beta1.ClusterServicePlan, error) {
	pcb := pretty.NewClusterServiceBrokerContextBuilder(broker)

	labelSelector := labels.SelectorFromSet(labels.Set{
		v1beta1.GroupName + "/" + v1beta1.FilterSpecClusterServiceBrokerName: util.GenerateSHA(broker.Name),
	}).String()

	listOpts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	existingServiceClasses, err := c.serviceCatalogClient.ClusterServiceClasses().List(context.Background(), listOpts)
	if err != nil {
		c.recorder.Eventf(broker, corev1.EventTypeWarning, errorListingClusterServiceClassesReason, "%v %v", errorListingClusterServiceClassesMessage, err)
		if err := c.updateClusterServiceBrokerCondition(
			broker,
			v1beta1.ServiceBrokerConditionReady,
			v1beta1.ConditionUnknown,
			errorListingClusterServiceClassesReason,
			errorListingClusterServiceClassesMessage,
		); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}
	klog.Info(pcb.Messagef("Found %d ClusterServiceClasses", len(existingServiceClasses.Items)))

	existingServicePlans, err := c.serviceCatalogClient.ClusterServicePlans().List(context.Background(), listOpts)
	if err != nil {
		c.recorder.Eventf(broker, corev1.EventTypeWarning, errorListingClusterServicePlansReason, "%v %v", errorListingClusterServicePlansMessage, err)
		if err := c.updateClusterServiceBrokerCondition(
			broker,
			v1beta1.ServiceBrokerConditionReady,
			v1beta1.ConditionUnknown,
			errorListingClusterServicePlansReason,
			errorListingClusterServicePlansMessage,
		); err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}
	klog.Info(pcb.Messagef("Found %d ClusterServicePlans", len(existingServicePlans.Items)))

	return existingServiceClasses.Items, existingServicePlans.Items, nil
}

func convertClusterServiceClassListToMap(list []v1beta1.ClusterServiceClass) map[string]*v1beta1.ClusterServiceClass {
	ret := make(map[string]*v1beta1.ClusterServiceClass, len(list))

	for i := range list {
		ret[list[i].Name] = &list[i]
	}

	return ret
}

func convertClusterServicePlanListToMap(list []v1beta1.ClusterServicePlan) map[string]*v1beta1.ClusterServicePlan {
	ret := make(map[string]*v1beta1.ClusterServicePlan, len(list))

	for i := range list {
		ret[list[i].Name] = &list[i]
	}

	return ret
}

func markAsServiceCatalogManagedResource(obj metav1.Object, broker *v1beta1.ClusterServiceBroker) {
	if isServiceCatalogManagedResource(obj) {
		return
	}

	var blockOwnerDeletion = false
	controllerRef := *metav1.NewControllerRef(broker, v1beta1.SchemeGroupVersion.WithKind("ClusterServiceBroker"))
	controllerRef.BlockOwnerDeletion = &blockOwnerDeletion

	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), controllerRef))
}

func isServiceCatalogManagedResource(resource metav1.Object) bool {
	c := metav1.GetControllerOf(resource)
	if c == nil {
		return false
	}

	return strings.HasPrefix(c.APIVersion, v1beta1.GroupName)
}
