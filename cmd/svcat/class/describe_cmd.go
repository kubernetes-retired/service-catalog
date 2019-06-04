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

package class

import (
	"fmt"
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// DescribeCmd contains the information needed to describe a specific class
type DescribeCmd struct {
	*command.Context
	*command.Namespaced
	*command.Scoped

	LookupByKubeName bool
	KubeName         string
	Name             string
}

// NewDescribeCmd builds a "svcat describe class" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &DescribeCmd{
		Context:    cxt,
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
	}
	cmd := &cobra.Command{
		Use:     "class NAME",
		Aliases: []string{"classes", "cl"},
		Short:   "Show details of a specific class",
		Example: command.NormalizeExamples(`
  svcat describe class mysqldb
  svcat describe class --kube-name 997b8372-8dac-40ac-ae65-758b4a5075a5
`),
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	cmd.Flags().BoolVarP(
		&describeCmd.LookupByKubeName,
		"kube-name",
		"k",
		false,
		"Whether or not to get the class by its Kubernetes name (the default is by external name)",
	)
	describeCmd.AddNamespaceFlags(cmd.Flags(), true)
	describeCmd.AddScopedFlags(cmd.Flags(), true)

	return cmd
}

// Validate checks that the required arguments have been provided
func (c *DescribeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a class external name or Kubernetes name is required")
	}

	if c.LookupByKubeName {
		c.KubeName = args[0]
	} else {
		c.Name = args[0]
	}

	return nil
}

// Run determines if we're getting a class by k8s name or
// external name, gets the details of the class, and prints
// the output to the user
func (c *DescribeCmd) Run() error {
	var class servicecatalog.Class
	var err error
	if c.Namespace == "" {
		c.Namespace = c.App.CurrentNamespace
	}
	scopeOpts := servicecatalog.ScopeOptions{
		Scope:     c.Scope,
		Namespace: c.Namespace,
	}

	if c.LookupByKubeName {
		class, err = c.App.RetrieveClassByID(c.KubeName, scopeOpts)
	} else {
		class, err = c.App.RetrieveClassByName(c.Name, scopeOpts)
	}
	if err != nil {
		if strings.Contains(err.Error(), servicecatalog.MultipleClassesFoundError) {
			return fmt.Errorf(err.Error() + ", please specify a scope with --scope or an exact Kubernetes name with --kube-name")
		}

		return err
	}

	output.WriteClassDetails(c.Output, class)

	opts := servicecatalog.ScopeOptions{Scope: servicecatalog.AllScope}
	plans, err := c.App.RetrievePlans(class.GetName(), opts)
	if err != nil {
		return err
	}
	output.WriteAssociatedPlans(c.Output, plans)

	return nil
}
