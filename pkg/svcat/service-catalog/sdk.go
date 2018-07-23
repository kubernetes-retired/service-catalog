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
	"time"

	apiv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// SvcatClient is an interface containing the variuos actions in the svcat pkg lib
// This interface is then faked with Counterfeiter for the cmd/svcat unit tests
type SvcatClient interface {
	Bind(string, string, string, string, string, interface{}, map[string]string) (*apiv1beta1.ServiceBinding, error)
	BindingParentHierarchy(*apiv1beta1.ServiceBinding) (*apiv1beta1.ServiceInstance, *apiv1beta1.ClusterServiceClass, *apiv1beta1.ClusterServicePlan, *apiv1beta1.ClusterServiceBroker, error)
	DeleteBinding(string, string) error
	DeleteBindings([]types.NamespacedName) ([]types.NamespacedName, error)
	IsBindingFailed(*apiv1beta1.ServiceBinding) bool
	IsBindingReady(*apiv1beta1.ServiceBinding) bool
	RetrieveBinding(string, string) (*apiv1beta1.ServiceBinding, error)
	RetrieveBindings(string) (*apiv1beta1.ServiceBindingList, error)
	RetrieveBindingsByInstance(*apiv1beta1.ServiceInstance) ([]apiv1beta1.ServiceBinding, error)
	Unbind(string, string) ([]types.NamespacedName, error)
	WaitForBinding(string, string, time.Duration, *time.Duration) (*apiv1beta1.ServiceBinding, error)

	Deregister(string) error
	RetrieveBrokers(opts ScopeOptions) ([]apiv1beta1.ClusterServiceBroker, error)
	RetrieveBroker(string) (*apiv1beta1.ClusterServiceBroker, error)
	RetrieveBrokerByClass(*apiv1beta1.ClusterServiceClass) (*apiv1beta1.ClusterServiceBroker, error)
	Register(string, string) (*apiv1beta1.ClusterServiceBroker, error)
	Sync(string, int) error

	RetrieveClasses(ScopeOptions) ([]Class, error)
	RetrieveClassByName(string) (*apiv1beta1.ClusterServiceClass, error)
	RetrieveClassByID(string) (*apiv1beta1.ClusterServiceClass, error)
	RetrieveClassByPlan(*apiv1beta1.ClusterServicePlan) (*apiv1beta1.ClusterServiceClass, error)

	Deprovision(string, string) error
	InstanceParentHierarchy(*apiv1beta1.ServiceInstance) (*apiv1beta1.ClusterServiceClass, *apiv1beta1.ClusterServicePlan, *apiv1beta1.ClusterServiceBroker, error)
	InstanceToServiceClassAndPlan(*apiv1beta1.ServiceInstance) (*apiv1beta1.ClusterServiceClass, *apiv1beta1.ClusterServicePlan, error)
	IsInstanceFailed(*apiv1beta1.ServiceInstance) bool
	IsInstanceReady(*apiv1beta1.ServiceInstance) bool
	Provision(string, string, string, string, string, interface{}, map[string]string) (*apiv1beta1.ServiceInstance, error)
	RetrieveInstance(string, string) (*apiv1beta1.ServiceInstance, error)
	RetrieveInstanceByBinding(*apiv1beta1.ServiceBinding) (*apiv1beta1.ServiceInstance, error)
	RetrieveInstances(string, string, string) (*apiv1beta1.ServiceInstanceList, error)
	RetrieveInstancesByPlan(*apiv1beta1.ClusterServicePlan) ([]apiv1beta1.ServiceInstance, error)
	TouchInstance(string, string, int) error
	WaitForInstance(string, string, time.Duration, *time.Duration) (*apiv1beta1.ServiceInstance, error)

	RetrievePlans(*FilterOptions) ([]apiv1beta1.ClusterServicePlan, error)
	RetrievePlanByName(string) (*apiv1beta1.ClusterServicePlan, error)
	RetrievePlanByID(string) (*apiv1beta1.ClusterServicePlan, error)
	RetrievePlansByClass(*apiv1beta1.ClusterServiceClass) ([]apiv1beta1.ClusterServicePlan, error)
	RetrievePlanByClassAndPlanNames(string, string) (*apiv1beta1.ClusterServicePlan, error)

	RetrieveSecretByBinding(*apiv1beta1.ServiceBinding) (*apicorev1.Secret, error)

	ServerVersion() (*version.Info, error)
}

// SDK wrapper around the generated Go client for the Kubernetes Service Catalog
type SDK struct {
	K8sClient            kubernetes.Interface
	ServiceCatalogClient clientset.Interface
}

// ServiceCatalog is the underlying generated Service Catalog versioned interface
// It should be used instead of accessing the client directly.
func (sdk *SDK) ServiceCatalog() v1beta1.ServicecatalogV1beta1Interface {
	return sdk.ServiceCatalogClient.ServicecatalogV1beta1()
}

// Core is the underlying generated Core API versioned interface
// It should be used instead of accessing the client directly.
func (sdk *SDK) Core() corev1.CoreV1Interface {
	return sdk.K8sClient.CoreV1()
}
