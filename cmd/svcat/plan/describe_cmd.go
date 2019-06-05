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

package plan

import (
	"fmt"
	"strings"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/output"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// DescribeCmd contains the needed info to fetch detailed info about a specific
// plan
type DescribeCmd struct {
	*command.Namespaced
	*command.Scoped
	LookupByKubeName bool
	ShowSchemas      bool
	KubeName         string
	Name             string
}

// NewDescribeCmd builds a "svcat describe plan" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &DescribeCmd{
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
	}
	cmd := &cobra.Command{
		Use:     "plan NAME",
		Aliases: []string{"plans", "pl"},
		Short:   "Show details of a specific plan",
		Example: command.NormalizeExamples(`
  svcat describe plan standard800
  svcat describe plan --kube-name 08e4b43a-36bc-447e-a81f-8202b13e339c
  svcat describe plan PLAN_NAME --scope cluster
  svcat describe plan PLAN_NAME --scope namespace --namespace NAMESPACE_NAME
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
	cmd.Flags().BoolVarP(
		&describeCmd.ShowSchemas,
		"show-schemas",
		"",
		true,
		"Whether or not to show instance and binding parameter schemas",
	)
	describeCmd.AddNamespaceFlags(cmd.Flags(), false)
	describeCmd.AddScopedFlags(cmd.Flags(), false)
	return cmd
}

// Validate and load the arguments passed to the svcat command.
func (c *DescribeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a plan name or Kubernetes name is required")
	}

	if c.LookupByKubeName {
		c.KubeName = args[0]
	} else {
		c.Name = args[0]
	}

	return nil
}

// Run determines how we are fetching a plan based
// on the provided arugments, and fetches the specified
// plan
func (c *DescribeCmd) Run() error {
	var plan servicecatalog.Plan
	var err error

	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}

	if c.LookupByKubeName {
		plan, err = c.App.RetrievePlanByID(c.KubeName, opts)
	} else if strings.Contains(c.Name, "/") {
		names := strings.Split(c.Name, "/")
		if len(names) != 2 {
			return fmt.Errorf("failed to parse class/plan name combination '%s'", c.Name)
		}
		plan, err = c.App.RetrievePlanByClassAndName(names[0], names[1], opts)
	} else {
		plan, err = c.App.RetrievePlanByName(c.Name, opts)
	}
	if err != nil {
		return err
	}

	// Retrieve the class as well because plans don't have the external class name
	class, err := c.App.RetrieveClassByPlan(plan)
	if err != nil {
		return err
	}

	output.WritePlanDetails(c.Output, plan, class)

	output.WriteDefaultProvisionParameters(c.Output, plan)

	instances, err := c.App.RetrieveInstancesByPlan(plan)
	if err != nil {
		return err
	}
	output.WriteAssociatedInstances(c.Output, instances)

	if c.ShowSchemas {
		output.WritePlanSchemas(c.Output, plan)
	}

	return nil
}
