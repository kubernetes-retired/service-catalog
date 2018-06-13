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

package instance

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/spf13/cobra"
)

type deprovisonCmd struct {
	*command.NamespacedCommand
	*command.WaitableCommand

	instanceName string
}

// NewDeprovisionCmd builds a "svcat deprovision" command
func NewDeprovisionCmd(cxt *command.Context) *cobra.Command {
	deprovisonCmd := &deprovisonCmd{
		NamespacedCommand: command.NewNamespacedCommand(cxt),
		WaitableCommand:   command.NewWaitableCommand(),
	}
	cmd := &cobra.Command{
		Use:   "deprovision NAME",
		Short: "Deletes an instance of a service",
		Example: command.NormalizeExamples(`
  svcat deprovision wordpress-mysql-instance
`),
		PreRunE: command.PreRunE(deprovisonCmd),
		RunE:    command.RunE(deprovisonCmd),
	}
	deprovisonCmd.AddNamespaceFlags(cmd.Flags(), false)
	deprovisonCmd.AddWaitFlags(cmd)

	return cmd
}

func (c *deprovisonCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.instanceName = args[0]

	return nil
}

func (c *deprovisonCmd) Run() error {
	return c.deprovision()
}

func (c *deprovisonCmd) deprovision() error {
	err := c.App.Deprovision(c.Namespace, c.instanceName)
	if err != nil {
		return err
	}

	if c.Wait {
		fmt.Fprintln(c.Output, "Waiting for the instance to be deleted...")

		var instance *v1beta1.ServiceInstance
		instance, err = c.App.WaitForInstance(c.Namespace, c.instanceName, c.Interval, c.Timeout)

		// The instance failed to deprovision cleanly, dump out more information on why
		if c.App.IsInstanceFailed(instance) {
			output.WriteInstanceDetails(c.Output, instance)
		}
	}

	if err == nil {
		output.WriteDeletedResourceName(c.Output, c.instanceName)
	}
	return err
}
