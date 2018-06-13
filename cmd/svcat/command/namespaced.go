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

package command

import "github.com/spf13/pflag"

// HasNamespaceFlags represents a command that can be scoped to a namespace.
type HasNamespaceFlags interface {
	Command

	// ApplyNamespaceFlags persists the namespace-related flags:
	// * --namespace
	// * --all-namespaces
	ApplyNamespaceFlags(flags *pflag.FlagSet)
}

// Namespaced is the base command of all svcat commands that are namespace scoped.
type NamespacedCommand struct {
	*Context
	Namespace string
}

// NewNamespacedCommand from context.
func NewNamespacedCommand(cxt *Context) *NamespacedCommand {
	return &NamespacedCommand{Context: cxt}
}

}

// ApplyNamespaceFlags persists the namespace-related flags:
// * --namespace
// * --all-namespaces
func (c *NamespacedCommand) ApplyNamespaceFlags(flags *pflag.FlagSet) {
	c.Namespace = c.determineNamespace(flags)
}

// determineNamespace using the current context's namespace, and the user-requested namespace.
func (c *NamespacedCommand) determineNamespace(flags *pflag.FlagSet) string {
	currentNamespace := c.Context.App.CurrentNamespace

	namespace, _ := flags.GetString("namespace")
	allNamespaces, _ := flags.GetBool("all-namespaces")

	if allNamespaces {
		return ""
	}

	if namespace != "" {
		return namespace
	}

	return currentNamespace
}
