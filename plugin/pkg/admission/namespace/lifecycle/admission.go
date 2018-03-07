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

package lifecycle

import (
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/namespace/lifecycle"
)

const (
	PluginName = "KubernetesNamespaceLifecycle"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	// This essentially registers the existing NamespaceLifecycle for kubernetes
	// another another name, for backwards-compatibility reasons
	plugins.Register(PluginName, func(io.Reader) (admission.Interface, error) {
		return NewLifecycle()
	})
}

// NewLifecycle creates a new namespace Lifecycle admission control handler
func NewLifecycle() (admission.Interface, error) {
	// NOTE: this list of namespaces comes from the original NamespaceLifecycle
	// admission plugin in k8s.io
	return lifecycle.NewLifecycle(sets.NewString(
		metav1.NamespaceDefault,
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	))
}
