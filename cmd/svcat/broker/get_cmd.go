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
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// GetCmd contains the information needed to get a broker or list of brokers
type GetCmd struct {
	*command.Namespaced
	*command.Formatted
	*command.Scoped

	Name string
}

// NewGetCmd builds a "svcat get brokers" command
func NewGetCmd(cxt *command.Context) *cobra.Command {
	getCmd := &GetCmd{
		Namespaced: command.NewNamespaced(cxt),
		Formatted:  command.NewFormatted(),
		Scoped:     command.NewScoped(),
	}
	cmd := &cobra.Command{
		Use:     "brokers [NAME]",
		Aliases: []string{"broker", "brk"},
		Short:   "List brokers, optionally filtered by name, scope or namespace",
		Example: command.NormalizeExamples(`
  svcat get brokers
  svcat get brokers --scope=cluster
  svcat get brokers --scope=all
  svcat get broker minibroker
`),
		PreRunE: command.PreRunE(getCmd),
		RunE:    command.RunE(getCmd),
	}
	getCmd.AddOutputFlags(cmd.Flags())
	getCmd.AddScopedFlags(cmd.Flags(), true)
	getCmd.AddNamespaceFlags(cmd.Flags(), true)
	return cmd
}

// Validate checks that the required arguments have been provided
func (c *GetCmd) Validate(args []string) error {
	if len(args) > 0 {
		c.Name = args[0]
	}

	return nil
}

// Run determines if we're getting all brokers or a single broker,
// then queries the backend to get that information
func (c *GetCmd) Run() error {
	if c.Name == "" {
		return c.getAll()
	}

	return c.get()
}

func (c *GetCmd) getAll() error {
	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}
	brokers, err := c.App.RetrieveBrokers(opts)
	if err != nil {
		return err
	}

	output.WriteBrokerList(c.Output, c.OutputFormat, brokers...)
	return nil
}

func (c *GetCmd) get() error {
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
	output.WriteBroker(c.Output, c.OutputFormat, broker)
	return nil
}
