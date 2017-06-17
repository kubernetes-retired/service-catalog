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

package exists

import (
	"fmt"
	"io"
	"sync"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/client-go/discovery"
	kubeclientset "k8s.io/client-go/kubernetes"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	scadmission "github.com/kubernetes-incubator/service-catalog/pkg/apiserver/admission"
)

const (
	// PluginName is name of admission plug-in
	PluginName = "ServiceCatalogPodPresetExists"
	// SupportedPodPresetVersion is the GroupVersion for what we require for
	// creating PodPresetTemplates in bindings.
	SupportedPodPresetVersion = "settings.k8s.io/v1alpha1"
)

func init() {
	admission.RegisterPlugin(PluginName, func(io.Reader) (admission.Interface, error) {
		return NewExists()
	})
}

// exists is an implementation of admission.Interface.
// It enforces that if a binding is being defined with a PodPresetTemplate
// that the cluster actually has support for it.
type exists struct {
	*admission.Handler
	discovery         discovery.DiscoveryInterface
	sync.Mutex        // protects podPresetsChecked and podPresetsExists
	podPresetsChecked bool
	podPresetsExists  bool
}

var _ = scadmission.WantsKubeClientSet(&exists{})

func (e *exists) supportsPodPresets() bool {
	e.Lock()
	defer e.Unlock()
	return e.podPresetsExists
}

// checkClusterForPodPresets uses discovery to find if the cluster has the
// right version of settings that supports PodPresets. It will return true
// if we have successfully queried the Discovery API and hence we know
// for sure if the cluster supports PodPresets or not.
// In that case, it also sets podPresetsChecked to true and podPresetsExists
// to true if the cluster supports PodPresets.
func (e *exists) checkClusterForPodPresets() bool {
	synced := false
	e.Lock()
	synced = e.podPresetsChecked
	e.Unlock()
	if synced {
		return true
	}

	resourceList, err := e.discovery.ServerResourcesForGroupVersion(SupportedPodPresetVersion)

	if err != nil && !errors.IsNotFound(err) {
		// We don't know if the cluster supports PodPresets or not
		glog.V(4).Infof("ServerResourcesForGroupVersion failed: %s", err)
		return false

	}

	e.Lock()
	defer e.Unlock()
	e.podPresetsChecked = true

	if err != nil && errors.IsNotFound(err) {
		// No such resource, which means the cluster does not support PodPresets
		glog.V(4).Infof("No PodPreset support in the cluster, not allowing PodPresetTemplates in Bindings")
		return true
	}

	for _, resource := range resourceList.APIResources {
		if resource.Name == "podpresets" {
			glog.V(4).Infof("PodPreset support found, allowing PodPresetTemplates in Bindings")
			e.podPresetsExists = true
		}
	}
	return true
}

func (e *exists) Admit(a admission.Attributes) error {
	// We only care about bindings
	if a.GetResource().Group != servicecatalog.GroupName || a.GetResource().GroupResource() != servicecatalog.Resource("bindings") {
		return nil
	}
	binding, ok := a.GetObject().(*servicecatalog.Binding)
	if !ok {
		return errors.NewBadRequest("Resource was marked with kind Binding but was unable to be converted")
	}
	if binding.Spec.AlphaPodPresetTemplate == nil {
		return nil
	}

	// we need to wait until we have successfully talked to the cluster and
	// determined if the cluster supports PodPresets or not.
	if !e.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("%s is not yet ready to handle request", PluginName))
	}

	// Request is for creating a binding with pod preset, check to see if the cluster supports it
	if e.supportsPodPresets() {
		return nil
	} else {
		return admission.NewForbidden(a, fmt.Errorf("Unable to create a binding with PodPresetTemplate because the cluster does not support PodPreset. Need support for %s resource", SupportedPodPresetVersion))
	}
}

// NewExists creates a new admission control handler that checks for existence of PodPresets
// before allowing a Binding to be created with PodPresetTemplate.
func NewExists() (admission.Interface, error) {
	return &exists{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}

func (e *exists) SetKubeClientSet(client kubeclientset.Interface) {
	e.discovery = client.Discovery()
	e.SetReadyFunc(e.checkClusterForPodPresets)
}

func (e *exists) Validate() error {
	if e.discovery == nil {
		return fmt.Errorf("missing discovery")
	}
	return nil
}
