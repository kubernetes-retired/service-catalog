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
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// DescribeCmd contains the info needed to describe a broker in detail
type DescribeCmd struct {
	*command.Context
	*command.Namespaced
	*command.Scoped

	Name string
}

// NewDescribeCmd builds a "svcat describe broker" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &DescribeCmd{
		Context:    cxt,
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
	}
	cmd := &cobra.Command{
		Use:     "broker NAME",
		Aliases: []string{"brokers", "brk"},
		Short:   "Show details of a specific broker",
		Example: command.NormalizeExamples(`
  svcat describe broker asb
`),
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	describeCmd.AddNamespaceFlags(cmd.Flags(), false)
	describeCmd.AddScopedFlags(cmd.Flags(), true)
	return cmd
}

// Validate checks that the required arguments have been provided
func (c *DescribeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a broker name is required")
	}
	c.Name = args[0]

	return nil
}

// Run retrieves the broker(s) with the requested name, interprets
// possible errors if we need to ask the user for more info, and displays
// the found broker to the user
func (c *DescribeCmd) Run() error {
	if c.Namespace == "" {
		c.Namespace = c.App.CurrentNamespace
	}
	scopeOpts := servicecatalog.ScopeOptions{
		Scope:     c.Scope,
		Namespace: c.Namespace,
	}
	broker, err := c.App.RetrieveBrokerByID(c.Name, scopeOpts)
	if err != nil {
		if strings.Contains(err.Error(), servicecatalog.MultipleBrokersFoundError) {
			return fmt.Errorf(err.Error() + ", please specify a scope with --scope")
		}
		return err
	}
	output.WriteBrokerDetails(c.Output, broker)
	return nil
}
