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

package binding

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type unbindCmd struct {
	*command.Namespaced
	*command.Waitable

	instanceName string
	bindingNames []string
}

// NewUnbindCmd builds a "svcat unbind" command
func NewUnbindCmd(cxt *command.Context) *cobra.Command {
	unbindCmd := &unbindCmd{
		Namespaced: command.NewNamespaced(cxt),
		Waitable:   command.NewWaitable(),
	}
	cmd := &cobra.Command{
		Use:   "unbind INSTANCE_NAME",
		Short: "Unbinds an instance. When an instance name is specified, all of its bindings are removed, otherwise use --name to remove a specific binding",
		Example: command.NormalizeExamples(`
  svcat unbind wordpress-mysql-instance
  svcat unbind --name wordpress-mysql-binding
`),
		PreRunE: command.PreRunE(unbindCmd),
		RunE:    command.RunE(unbindCmd),
	}
	unbindCmd.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().StringSliceVar(
		&unbindCmd.bindingNames,
		"name",
		[]string{},
		"The name of the binding to remove",
	)
	unbindCmd.AddWaitFlags(cmd)

	return cmd
}

func (c *unbindCmd) Validate(args []string) error {
	if len(args) == 0 {
		if len(c.bindingNames) == 0 {
			return fmt.Errorf("an instance or binding name is required")
		}
	} else {
		c.instanceName = args[0]
	}

	return nil
}

func (c *unbindCmd) Run() error {
	// Indicates an error occurred and that a non-zero exit code should be used
	var hasErrors bool
	var bindings []types.NamespacedName
	var err error

	if c.instanceName != "" {
		bindings, err = c.App.Unbind(c.Namespace, c.instanceName)
	} else {
		bindings, err = c.App.DeleteBindings(c.getBindingsToDelete())
	}

	if err != nil {
		// Do not return immediately as we still need to potentially wait or print the deleted bindings
		hasErrors = true
		fmt.Fprintln(c.Output, err)
	}

	if c.Wait {
		hasErrors = c.waitForBindingDeletes("waiting for the binding(s) to be deleted...", bindings...) || hasErrors
	} else {
		for _, binding := range bindings {
			output.WriteDeletedResourceName(c.Output, binding.Name)
		}
	}

	if hasErrors {
		return errors.New("could not remove all bindings")
	}
	return nil
}

func (c *unbindCmd) getBindingsToDelete() []types.NamespacedName {
	bindings := []types.NamespacedName{}
	for _, name := range c.bindingNames {
		bindings = append(bindings, types.NamespacedName{Namespace: c.Namespace, Name: name})
	}
	return bindings
}

// waitForBindingDeletes waits for the bindings to be deleted and prints either
// and error message or the name of the deleted binding.
func (c *unbindCmd) waitForBindingDeletes(waitMessage string, bindings ...types.NamespacedName) bool {
	if len(bindings) == 0 {
		return false
	}

	// Indicates an error occurred and that a non-zero exit code should be used
	var hasErrors bool

	// Used to prevent concurrent writes to c.Output
	var mutex sync.Mutex

	fmt.Fprintln(c.Output, waitMessage)

	var g sync.WaitGroup
	for _, binding := range bindings {
		g.Add(1)
		go func(ns, name string) {
			defer g.Done()

			binding, err := c.App.WaitForBinding(ns, name, c.Interval, c.Timeout)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil && !apierrors.IsNotFound(errors.Cause(err)) {
				hasErrors = true
				fmt.Fprintln(c.Output, err)
			} else if c.App.IsBindingFailed(binding) {
				hasErrors = true
				fmt.Fprintf(c.Output, "could not delete binding %s/%s\n", ns, name)
			} else {
				output.WriteDeletedResourceName(c.Output, name)
			}
		}(binding.Namespace, binding.Name)
	}
	g.Wait()

	return hasErrors
}
