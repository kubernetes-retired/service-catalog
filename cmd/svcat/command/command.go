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

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Command represents an svcat command.
type Command interface {
	// GetContext retrieves the command's context.
	GetContext() *Context

	// Validate and load the arguments passed to the svcat command.
	Validate(args []string) error

	// Run a validated svcat command.
	Run() error
}

// NamespacedCommand represents a command that can be scoped to a namespace.
type NamespacedCommand interface {
	Command

	// SetNamespace sets the effective namespace for the command.
	SetNamespace(namespace string)
}

// PreRunE validates os args, and then saves them on the svcat command.
func PreRunE(cmd Command) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if nsCmd, ok := cmd.(NamespacedCommand); ok {
			namespace := DetermineNamespace(c.Flags(), nsCmd.GetContext().App.CurrentNamespace)
			nsCmd.SetNamespace(namespace)
		}
		return cmd.Validate(args)
	}
}

// RunE executes a validated svcat command.
func RunE(cmd Command) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		return cmd.Run()
	}
}

// AddNamespaceFlags applies the --namespace and --all-namespaces flags to a command.
// This is intended to be used in conjunction with the NamespacedCommand interface.
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this request",
	)

	if allowAll {
		flags.Bool(
			"all-namespaces",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace",
		)
	}
}

// DetermineNamespace using the current context's namespace, and the user-requested namespace.
func DetermineNamespace(flags *pflag.FlagSet, currentNamespace string) string {
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
