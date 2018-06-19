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

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	*command.Namespaced
	name string
}

// NewDescribeCmd builds a "svcat describe instance" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{Namespaced: command.NewNamespaced(cxt)}
	cmd := &cobra.Command{
		Use:     "instance NAME",
		Aliases: []string{"instances", "inst"},
		Short:   "Show details of a specific instance",
		Example: command.NormalizeExamples(`
  svcat describe instance wordpress-mysql-instance
`),
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	describeCmd.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}

func (c *describeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("an instance name is required")
	}
	c.name = args[0]

	return nil
}

func (c *describeCmd) Run() error {
	return c.describe()
}

func (c *describeCmd) describe() error {
	instance, err := c.App.RetrieveInstance(c.Namespace, c.name, "", "")
	if err != nil {
		return err
	}

	output.WriteInstanceDetails(c.Output, instance)

	bindings, err := c.App.RetrieveBindingsByInstance(instance)
	if err != nil {
		return err
	}
	output.WriteAssociatedBindings(c.Output, bindings)

	return nil
}
