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
	"math/rand"
	"net/http"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
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

	// CRDProbeIterationGap - the number of iterations after which the CRD probe action is performed
	// All probes are run after the time period defined in the `periodSeconds` parameter in the chart
	// Time needed for the CRD Probe to execute is `periodSeconds` * CRDProbeIterationGap
	CRDProbeIterationGap = 60
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

// CRDProbe provides functionality that ensures that all ServiceCatalog CRDs are ready
type CRDProbe struct {
	client  apiextensionsclientset.Interface
	delay   int
	counter int
}

// NewCRDProbe returns pointer to CRDProbe
func NewCRDProbe(apiextensionsClient apiextensionsclientset.Interface, delay int) *CRDProbe {
	return &CRDProbe{client: apiextensionsClient, counter: 0, delay: delay}
}

// Name returns name of CRD probe
func (r CRDProbe) Name() string {
	return fmt.Sprintf("ready-CRDs-%d", rand.Intn(1000))
}

// Check if all CRDs with specific label are ready
func (r *CRDProbe) Check(_ *http.Request) error {
	if r.counter < r.delay {
		r.counter++
		klog.V(4).Infof("%s CRDProbe skipped. Will be executed in %d iteration", r.Name(), r.delay-r.counter)
		return nil
	}
	r.counter = 0
	result, err := r.check()
	if result && err == nil {
		return nil
	}

	return fmt.Errorf("CRDs are not ready")
}

// IsReady returns true if all required CRDs are ready
func (r *CRDProbe) IsReady() (bool, error) {
	return r.check()
}

func (r *CRDProbe) check() (bool, error) {
	list, err := r.client.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{"svcat": "true"}).String()})
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
