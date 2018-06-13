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
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type unbindCmd struct {
	*command.NamespacedCommand
	*command.WaitableCommand

	instanceName string
	bindingName  string
}

// NewUnbindCmd builds a "svcat unbind" command
func NewUnbindCmd(cxt *command.Context) *cobra.Command {
	unbindCmd := &unbindCmd{
		NamespacedCommand: command.NewNamespacedCommand(cxt),
		WaitableCommand:   command.NewWaitableCommand(),
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
	command.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().StringVar(
		&unbindCmd.bindingName,
		"name",
		"",
		"The name of the binding to remove",
	)
	unbindCmd.AddWaitFlags(cmd)

	return cmd
}

func (c *unbindCmd) Validate(args []string) error {
	if len(args) == 0 {
		if c.bindingName == "" {
			return fmt.Errorf("an instance or binding name is required")
		}
	} else {
		c.instanceName = args[0]
	}

	return nil
}

func (c *unbindCmd) Run() error {
	if c.instanceName != "" {
		return c.unbindInstance()
	}
	return c.deleteBinding()
}

func (c *unbindCmd) deleteBinding() error {
	err := c.App.DeleteBinding(c.Namespace, c.bindingName)
	if err != nil {
		return err
	}

	if c.Wait {
		binding := v1beta1.ServiceBinding{ObjectMeta: metav1.ObjectMeta{Namespace: c.Namespace, Name: c.bindingName}}
		hasErr := c.waitForBindingDeletes("waiting for the binding to be deleted...", binding)
		if hasErr {
			// Ensure a non-zero exit code is returned if the wait has trouble
			return errors.New("could not remove the binding")
		}
	} else {
		output.WriteDeletedResourceName(c.Output, c.bindingName)
	}

	// Don't return errors because we handle printing them as they occur above
	return nil
}

func (c *unbindCmd) unbindInstance() error {
	// Indicates an error occurred and that a non-zero exit code should be used
	var hasErrors bool

	bindings, err := c.App.Unbind(c.Namespace, c.instanceName)
	if err != nil {
		// Do not return immediately as we still need to potentially wait or print the deleted bindings
		hasErrors = true
		fmt.Fprintln(c.Output, err)
	}

	if c.Wait {
		hasErrors = c.waitForBindingDeletes("waiting for the bindings to be deleted...", bindings...) || hasErrors
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

// waitForBindingDeletes waits for the bindings to be deleted and prints either
// and error message or the name of the deleted binding.
func (c *unbindCmd) waitForBindingDeletes(waitMessage string, bindings ...v1beta1.ServiceBinding) bool {
	if len(bindings) == 0 {
		return false
	}

	// Indicates an error occurred and that a non-zero exit code should be used
	var hasErrors bool

	fmt.Fprintln(c.Output, waitMessage)

	var g sync.WaitGroup
	for _, binding := range bindings {
		g.Add(1)
		go func(ns, name string) {
			defer g.Done()

			binding, err := c.App.WaitForBinding(ns, name, c.Interval, c.Timeout)

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
