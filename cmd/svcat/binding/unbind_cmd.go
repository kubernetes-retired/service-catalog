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
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

type unbindCmd struct {
	*command.Namespaced
	instanceName string
	bindingName  string
	wait         bool
	rawTimeout   string
	timeout      *time.Duration
}

// NewUnbindCmd builds a "svcat unbind" command
func NewUnbindCmd(cxt *command.Context) *cobra.Command {
	unbindCmd := &unbindCmd{Namespaced: command.NewNamespacedCommand(cxt)}
	cmd := &cobra.Command{
		Use:   "unbind INSTANCE_NAME",
		Short: "Unbinds an instance. When an instance name is specified, all of its bindings are removed, otherwise use --name to remove a specific binding",
		Example: `
  svcat unbind wordpress-mysql-instance
  svcat unbind --name wordpress-mysql-binding
`,
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
	cmd.Flags().BoolVar(&unbindCmd.wait, "wait", false,
		"Wait until the operation completes.")
	cmd.Flags().StringVar(&unbindCmd.rawTimeout, "timeout", "5m",
		"Timeout for --wait, specified in human readable format: 30s, 1m, 1h. Specify -1 to wait indefinitely.")
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

	if c.wait && c.rawTimeout != "-1" {
		timeout, err := time.ParseDuration(c.rawTimeout)
		if err != nil {
			return fmt.Errorf("invalid --timeout value (%s)", err)
		}
		c.timeout = &timeout
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

	if c.wait {
		glog.V(2).Infof("Waiting for the binding to be deleted...")
		pollInterval := 1 * time.Second

		var binding *v1beta1.ServiceBinding
		binding, err = c.App.WaitForBinding(c.Namespace, c.bindingName, pollInterval, c.timeout)

		// The binding failed to delete cleanly, dump out more information on why
		if c.App.IsBindingFailed(binding) {
			output.WriteBindingDetails(c.Output, binding)
		}
	}

	if err == nil {
		output.WriteDeletedResourceName(c.Output, c.bindingName)
	}
	return err
}

func (c *unbindCmd) unbindInstance() error {
	bindings, err := c.App.Unbind(c.Namespace, c.instanceName)
	if err != nil {
		return err
	}

	if c.wait {
		glog.V(2).Infof("Waiting for the bindings to be deleted...")
		pollInterval := 1 * time.Second
		var g sync.WaitGroup
		for _, binding := range bindings {
			g.Add(1)
			go func(ns, name string) {
				defer g.Done()

				binding, err := c.App.WaitForBinding(ns, name, pollInterval, c.timeout)

				if err != nil {
					fmt.Fprintf(c.Output, "Error: %s", err.Error())
				} else if c.App.IsBindingFailed(binding) {
					fmt.Fprintf(c.Output, "could not delete binding %s/%s", ns, name)
				} else {
					output.WriteDeletedResourceName(c.Output, name)
				}
			}(binding.Namespace, binding.Name)
		}
		g.Wait()
	}

	// Don't return errors because we handle printing them as they occur above
	return nil
}
