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

package broker

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// DeregisterCmd contains the info needed to delete a broker
type DeregisterCmd struct {
	*command.Namespaced
	*command.Scoped
	*command.Waitable

	BrokerName string
}

// NewDeregisterCmd builds a "svcat deregister" command
func NewDeregisterCmd(cxt *command.Context) *cobra.Command {
	deregisterCmd := &DeregisterCmd{
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
		Waitable:   command.NewWaitable(),
	}
	cmd := &cobra.Command{
		Use:   "deregister NAME",
		Short: "Deregisters an existing broker with service catalog",
		Example: command.NormalizeExamples(`
		svcat deregister mysqlbroker
		svcat deregister mysqlbroker --namespace=mysqlnamespace
		svcat deregister mysqlclusterbroker --cluster
		`),
		PreRunE: command.PreRunE(deregisterCmd),
		RunE:    command.RunE(deregisterCmd),
	}
	deregisterCmd.AddNamespaceFlags(cmd.Flags(), false)
	deregisterCmd.AddScopedFlags(cmd.Flags(), false)
	deregisterCmd.AddWaitFlags(cmd)
	return cmd
}

// Validate checks that the required arguments have been provided
func (c *DeregisterCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a broker name is required")
	}
	c.BrokerName = args[0]

	return nil
}

// Run runs the command
func (c *DeregisterCmd) Run() error {
	return c.Deregister()
}

// Deregister calls out to the pkg lib to delete the broker and display the output
func (c *DeregisterCmd) Deregister() error {
	scopeOptions := &servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	err := c.Context.App.Deregister(c.BrokerName, scopeOptions)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.Context.Output, "Successfully removed broker %q\n", c.BrokerName)
	return nil
}
