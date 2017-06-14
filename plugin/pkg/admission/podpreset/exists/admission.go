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
	"time"

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

	// how long to wait for a missing namespace before re-checking the cache (and then doing a live lookup)
	// this accomplishes two things:
	// 1. It allows a watch-fed cache time to observe a namespace creation event
	// 2. It allows time for a namespace creation to distribute to members of a storage cluster,
	//    so the live lookup has a better chance of succeeding even if it isn't performed against the leader.
	missingNamespaceWait = 50 * time.Millisecond
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
	sync.Mutex        // protects next two fields
	podPresetsChecked bool
	podPresetsExists  bool
}

var _ = scadmission.WantsKubeClientSet(&exists{})

func (l *exists) hasSupportForPodPresets() (bool, error) {
	// If we've previously checked the cluster, use that information.
	l.Lock()
	if l.podPresetsChecked {
		l.Unlock()
		return l.podPresetsExists, nil
	}
	l.Unlock()

	// This is the first time or we couldn't tell previously for sure
	//(error?), so check the cluster. Since this does IO, note we release
	// the lock above.
	exists, err := l.checkClusterForPodPresets()

	// Grab the lock again, check to make sure that somebody else hasn't
	// checked the cluster in the meantime. In theory this shouldn't matter,
	// but if they succeeded and we got an error, we should just use their
	// result.
	l.Lock()
	if l.podPresetsChecked {
		l.Unlock()
		return l.podPresetsExists, nil
	}
	// No errors, so we know  if the cluster supports PodPreset or not
	if err == nil {
		l.podPresetsExists = exists
		l.podPresetsChecked = true
	}
	l.Unlock()
	return exists, err
}

// checkClusterForPodPresets uses discovery to find if the cluster has the
// right version of settings that supports PodPresets
// This is a separate function from above since it does IO and we don't
// want to hang on to a lock while doing this.
func (l *exists) checkClusterForPodPresets() (bool, error) {
	if resourceList, err := l.discovery.ServerResourcesForGroupVersion(SupportedPodPresetVersion); err != nil {
		if errors.IsNotFound(err) {
			glog.V(4).Infof("No PodPreset support in the cluster, not allowing PodPresetTemplates in Bindings")
			// No such resource, which means the cluster does not support PodPresets
			return false, nil
		} else {
			// Some other kind of error, so just return it
			glog.V(4).Infof("ServerResourcesForGroupVersion failed: %s", err)
			return false, err
		}
	} else {
		for _, resource := range resourceList.APIResources {
			if resource.Name == "podpresets" {
				glog.V(4).Infof("PodPreset support found, allowing PodPresetTemplates in Bindings")
				return true, nil
			}
		}
	}
	return false, nil
}

func (l *exists) Admit(a admission.Attributes) error {
	// We only care about updates / creates
	if a.GetOperation() != admission.Create && a.GetOperation() != admission.Update {
		return nil
	}

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

	// Request is for creating a binding with pod preset, check to see if the cluster supports it
	ppExists, err := l.hasSupportForPodPresets()
	if err != nil {
		return errors.NewInternalError(err)
	}

	if ppExists {
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

func (l *exists) SetKubeClientSet(client kubeclientset.Interface) {
	l.discovery = client.Discovery()
}

func (l *exists) Validate() error {
	if l.discovery == nil {
		return fmt.Errorf("missing discovery")
	}
	return nil
}
