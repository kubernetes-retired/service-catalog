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

package probe

import (
	"fmt"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"net/http"
)

const (
	// CRDsAmount define the whole number of CRDs registered by the Service Catalog
	CRDsAmount = 8

	// ClusterServiceBroker define the name of the ClusterServiceBroker CRD
	ClusterServiceBroker = "clusterservicebrokers.servicecatalog.k8s.io"
	// ServiceBroker define the name of the ServiceBroker CRD
	ServiceBroker = "servicebrokers.servicecatalog.k8s.io"
	// ServiceClass define the name of the ServiceClass CRD
	ServiceClass = "serviceclasses.servicecatalog.k8s.io"
	// ClusterServiceClass define the name of the ClusterServiceClass CRD
	ClusterServiceClass = "clusterserviceclasses.servicecatalog.k8s.io"
	// ServicePlan define the name of the ServicePlan CRD
	ServicePlan = "serviceplans.servicecatalog.k8s.io"
	// ClusterServicePlan define the name of the ClusterServicePlan CRD
	ClusterServicePlan = "clusterserviceplans.servicecatalog.k8s.io"
	// ServiceInstance define the name of the ServiceInstance CRD
	ServiceInstance = "serviceinstances.servicecatalog.k8s.io"
	// ServiceBinding define the name of the ServiceBinding CRD
	ServiceBinding = "servicebindings.servicecatalog.k8s.io"
)

var customResourceDefinitionNames = []string{
	ClusterServiceBroker,
	ServiceBroker,
	ServiceClass,
	ClusterServiceClass,
	ServicePlan,
	ClusterServicePlan,
	ServiceInstance,
	ServiceBinding,
}

// ReadinessCRD provides functionality that ensures that all ServiceCatalog CRDs are ready
type ReadinessCRD struct {
	client apiextensionsclientset.Interface
}

// NewReadinessCRDProbe returns pointer to ReadinessCRD
func NewReadinessCRDProbe(apiextensionsClient apiextensionsclientset.Interface) (*ReadinessCRD, error) {
	return &ReadinessCRD{client: apiextensionsClient}, nil
}

// Name returns name of readiness probe
func (r ReadinessCRD) Name() string {
	return "ready-CRDs"
}

// Check if all CRDs with specific label are ready
func (r *ReadinessCRD) Check(_ *http.Request) error {
	result, err := r.check()
	if result && err == nil {
		return nil
	}

	return fmt.Errorf("CRDs are not ready")
}

// IsReady returns true if all required CRDs are ready
func (r *ReadinessCRD) IsReady() (bool, error) {
	return r.check()
}

func (r *ReadinessCRD) check() (bool, error) {
	list, err := r.client.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list CustomResourceDefinition: %s", err)
	}
	var amount int
	for _, crd := range list.Items {
		if !IsServiceCatalogCustomResourceDefinition(crd) {
			continue
		}
		if crdStatusConditionIsTrue(crd.Status) {
			klog.V(4).Infof("CRD %q is ready", crd.Name)
			amount++
		}
	}
	if amount != CRDsAmount {
		klog.V(4).Infof("the correct number of elements should be %d, there are %d elements", CRDsAmount, amount)
		return false, nil
	}

	klog.V(4).Infof("Readiness probe %s checked. There are %d CRDs", r.Name(), amount)
	return true, nil
}

func crdStatusConditionIsTrue(status v1beta1.CustomResourceDefinitionStatus) bool {
	for _, condition := range status.Conditions {
		if condition.Type != v1beta1.Established {
			continue
		}
		if condition.Status == "True" {
			return true
		}
	}

	return false
}

// IsServiceCatalogCustomResourceDefinition checks if CRD belongs to ServiceCatalog crd
func IsServiceCatalogCustomResourceDefinition(crd v1beta1.CustomResourceDefinition) bool {
	for _, crdName := range customResourceDefinitionNames {
		if crdName == crd.Name {
			return true
		}
	}

	return false
}
