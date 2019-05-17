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

package cleaner

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

const (
	finalizerCheckPerdiodTime = 1 * time.Second
	finalizerCheckTimeout     = 30 * time.Second
)

// FinalizerCleaner is responsible for removing ServiceCatalog finalizers from ServiceCatalog CRs
// and makes sure all finalizers from CRs are removed
type FinalizerCleaner struct {
	client sc.Interface
}

// NewFinalizerCleaner returns new pointer to FinalizerCleaner
func NewFinalizerCleaner(scClient sc.Interface) *FinalizerCleaner {
	return &FinalizerCleaner{scClient}
}

// RemoveFinalizers removes specific finalizers from all ServiceCatalog CRs
func (fc *FinalizerCleaner) RemoveFinalizers() error {
	klog.V(4).Infof("Removing finalizers from %s", pretty.ClusterServiceBroker)
	err := removeFinalizerFromClusterServiceBroker(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ClusterServiceBroker, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ServiceBroker)
	err = removeFinalizerFromServiceBroker(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ServiceBroker, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ClusterServiceClass)
	err = removeFinalizerFromClusterServiceClass(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ClusterServiceClass, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ServiceClass)
	err = removeFinalizerFromServiceClass(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ServiceClass, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ClusterServicePlan)
	err = removeFinalizerFromClusterServicePlan(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ClusterServicePlan, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ServicePlan)
	err = removeFinalizerFromServicePlan(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ServicePlan, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ServiceInstance)
	err = removeFinalizerFromServiceInstance(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ServiceInstance, err)
	}

	klog.V(4).Infof("Removing finalizers from %s", pretty.ServiceBinding)
	err = removeFinalizerFromServiceBinding(fc.client)
	if err != nil {
		return fmt.Errorf("failed during removing %s finalizer: %s", pretty.ServiceBinding, err)
	}

	return nil
}

func removeFinalizerFromClusterServiceBroker(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ClusterServiceBroker, err)
	}

	for _, broker := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(broker.Finalizers)
		toUpdate := broker.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ClusterServiceBrokerName(toUpdate.Name), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServiceBrokers().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceBroker(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceBrokers(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ServiceBroker, err)
	}

	for _, broker := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(broker.Finalizers)
		toUpdate := broker.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceBrokers(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ServiceBrokerName(toUpdate.Name), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceBrokers(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromClusterServiceClass(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServiceClasses().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ClusterServiceClass, err)
	}

	for _, class := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(class.Finalizers)
		toUpdate := class.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServiceClasses().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ClusterServiceClassName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServiceClasses().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceClass(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceClasses(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ServiceClass, err)
	}

	for _, class := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(class.Finalizers)
		toUpdate := class.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceClasses(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ServiceClassName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceClasses(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromClusterServicePlan(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ClusterServicePlans().List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ClusterServicePlan, err)
	}

	for _, plan := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(plan.Finalizers)
		toUpdate := plan.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ClusterServicePlans().Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ClusterServicePlanName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ClusterServicePlans().Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServicePlan(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServicePlans(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ServicePlan, err)
	}

	for _, plan := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(plan.Finalizers)
		toUpdate := plan.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServicePlans(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ServicePlanName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServicePlans(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceInstance(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceInstances(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ServiceInstance, err)
	}

	for _, instance := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(instance.Finalizers)
		toUpdate := instance.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceInstances(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ServiceInstanceName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceInstances(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

func removeFinalizerFromServiceBinding(client sc.Interface) error {
	list, err := client.ServicecatalogV1beta1().ServiceBindings(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to list %s: %s", pretty.ServiceBinding, err)
	}

	for _, binding := range list.Items {
		finalizersList := removeServiceCatalogFinalizer(binding.Finalizers)
		toUpdate := binding.DeepCopy()
		toUpdate.Finalizers = finalizersList
		_, err := client.ServicecatalogV1beta1().ServiceBindings(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update %s: %s", pretty.ServiceBindingName(toUpdate), err)
		}
		err = wait.Poll(finalizerCheckPerdiodTime, finalizerCheckTimeout, func() (done bool, err error) {
			klog.V(4).Info("waiting for the finalizer to be removed")
			cr, err := client.ServicecatalogV1beta1().ServiceBindings(toUpdate.Namespace).Get(toUpdate.Name, v1.GetOptions{})
			return checkFinalizerIsRemoved(cr, err)
		})
		if err != nil {
			return fmt.Errorf("failed while waiting for finalizers will be removed: %s", err)
		}
	}

	return nil
}

// FinalizerGetter contract for structs which has finalizers
type FinalizerGetter interface {
	GetFinalizers() []string
}

func checkFinalizerIsRemoved(cr FinalizerGetter, err error) (bool, error) {
	if errors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if len(cr.GetFinalizers()) == 0 {
		return true, nil
	}
	klog.V(4).Info("finalizers not removed, retry...")
	return false, nil
}

func removeServiceCatalogFinalizer(finalizersList []string) []string {
	finalizers := sets.NewString(finalizersList...)
	if finalizers.Has(v1beta1.FinalizerServiceCatalog) {
		finalizers.Delete(v1beta1.FinalizerServiceCatalog)
	}

	return finalizers.List()
}
