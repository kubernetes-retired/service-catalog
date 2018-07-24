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
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/spf13/cobra"
)

// RegisterCmd contains the information needed to register a broker
type RegisterCmd struct {
	BrokerName string
	Context    *command.Context
	URL        string
}

// NewRegisterCmd builds a "svcat register" command
func NewRegisterCmd(cxt *command.Context) *cobra.Command {
	registerCmd := &RegisterCmd{
		Context: cxt,
	}
	cmd := &cobra.Command{
		Use:   "register NAME --url URL",
		Short: "Registers a new broker with service catalog",
		Example: command.NormalizeExamples(`
		svcat register mysqlbroker --url http://mysqlbroker.com
		`),
		PreRunE: command.PreRunE(registerCmd),
		RunE:    command.RunE(registerCmd),
	}
	cmd.Flags().StringVar(&registerCmd.URL, "url", "",
		"The broker URL (Required)")
	cmd.MarkFlagRequired("url")
	return cmd
}

// Validate checks that the required arguements have been provided
func (c *RegisterCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a broker name is required")
	}
	c.BrokerName = args[0]

	return nil
}

// Run runs the command
func (c *RegisterCmd) Run() error {
	return c.Register()
}

// Register calls out to the pkg lib to create the broker and displays the output
func (c *RegisterCmd) Register() error {
	broker, err := c.Context.App.Register(c.BrokerName, c.URL)
	if err != nil {
		return err
	}

	output.WriteBrokerDetails(c.Context.Output, broker)
	return nil
}
