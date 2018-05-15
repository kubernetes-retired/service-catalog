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

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	*command.Namespaced
	name        string
	showSecrets bool
}

// NewDescribeCmd builds a "svcat describe binding" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{Namespaced: command.NewNamespacedCommand(cxt)}
	cmd := &cobra.Command{
		Use:     "binding NAME",
		Aliases: []string{"bindings", "bnd"},
		Short:   "Show details of a specific binding",
		Example: `
  svcat describe binding wordpress-mysql-binding
`,
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	command.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().BoolVar(
		&describeCmd.showSecrets,
		"show-secrets",
		false,
		"Output the decoded secret values. By default only the length of the secret is displayed",
	)
	return cmd
}

func (c *describeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("name is required")
	}
	c.name = args[0]

	return nil
}

func (c *describeCmd) Run() error {
	return c.describe()
}

func (c *describeCmd) describe() error {
	binding, err := c.App.RetrieveBinding(c.Namespace, c.name)
	if err != nil {
		return err
	}

	output.WriteBindingDetails(c.Output, binding)

	secret, err := c.App.RetrieveSecretByBinding(binding)
	output.WriteAssociatedSecret(c.Output, secret, err, c.showSecrets)

	return nil
}
