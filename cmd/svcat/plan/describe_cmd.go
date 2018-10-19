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

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	*command.Namespaced
	*command.Scoped
	lookupByUUID bool
	showSchemas  bool
	uuid         string
	name         string
}

// NewDescribeCmd builds a "svcat describe plan" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{
		Namespaced: command.NewNamespaced(cxt),
		Scoped:     command.NewScoped(),
	}
	cmd := &cobra.Command{
		Use:     "plan NAME",
		Aliases: []string{"plans", "pl"},
		Short:   "Show details of a specific plan",
		Example: command.NormalizeExamples(`
  svcat describe plan standard800
  svcat describe plan --uuid 08e4b43a-36bc-447e-a81f-8202b13e339c
  svcat describe plan PLAN_NAME --scope cluster
  svcat describe plan PLAN_NAME --scope namespace --namespace NAMESPACE_NAME
`),
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	cmd.Flags().BoolVarP(
		&describeCmd.lookupByUUID,
		"uuid",
		"u",
		false,
		"Whether or not to get the class by UUID (the default is by name)",
	)
	cmd.Flags().BoolVarP(
		&describeCmd.showSchemas,
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
func (c *describeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a plan name or uuid is required")
	}

	if c.lookupByUUID {
		c.uuid = args[0]
	} else {
		c.name = args[0]
	}

	return nil
}

// Run a validated svcat command.
func (c *describeCmd) Run() error {
	return c.describe()
}

func (c *describeCmd) describe() error {
	var plan servicecatalog.Plan
	var err error

	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     c.Scope,
	}

	if c.lookupByUUID {
		plan, err = c.App.RetrievePlanByID(c.uuid, opts)
	} else if strings.Contains(c.name, "/") {
		names := strings.Split(c.name, "/")
		if len(names) != 2 {
			return fmt.Errorf("failed to parse class/plan name combination '%s'", c.name)
		}
		plan, err = c.App.RetrievePlanByClassAndName(names[0], names[1], opts)
	} else {
		plan, err = c.App.RetrievePlanByName(c.name, opts)
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

	if c.showSchemas {
		output.WritePlanSchemas(c.Output, plan)
	}

	return nil
}
