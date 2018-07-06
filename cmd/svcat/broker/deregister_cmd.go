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
	"github.com/spf13/cobra"
)

// DeregisterCmd contains the info needed to delete a broker
type DeregisterCmd struct {
	BrokerName string
	Context    *command.Context
}

// NewDeregisterCmd builds a "svcat deregister" command
func NewDeregisterCmd(cxt *command.Context) *cobra.Command {
	deregisterCmd := &DeregisterCmd{
		Context: cxt,
	}
	cmd := &cobra.Command{
		Use:   "deregister NAME",
		Short: "Deregisters an existing broker with service catalog",
		Example: command.NormalizeExamples(`
		svcat deregister mysqlbroker
		`),
		PreRunE: command.PreRunE(deregisterCmd),
		RunE:    command.RunE(deregisterCmd),
	}
	return cmd
}

// Validate checks that the required arguements have been provided
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
	err := c.Context.App.Deregister(c.BrokerName)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.Context.Output, "Successfully removed broker %q", c.BrokerName)
	return nil
}
