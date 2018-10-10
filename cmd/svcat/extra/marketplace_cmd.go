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

package extra

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/spf13/cobra"
)

// MarketplaceCmd contains the information needed to query the marketplace of
// services available to the user
type MarketplaceCmd struct {
	*command.Namespaced
	*command.Formatted
}

// NewMarketplaceCmd builds a "svcat marketplace" command
func NewMarketplaceCmd(cxt *command.Context) *cobra.Command {
	mpCmd := &MarketplaceCmd{
		Namespaced: command.NewNamespaced(cxt),
		Formatted:  command.NewFormatted(),
	}
	cmd := &cobra.Command{
		Use:     "marketplace",
		Aliases: []string{"marketplace", "mp"},
		Short:   "List available service offerings",
		Example: command.NormalizeExamples(`
  svcat marketplace
	svcat marketplace --namespace dev
`),
		PreRunE: command.PreRunE(mpCmd),
		RunE:    command.RunE(mpCmd),
	}

	mpCmd.AddOutputFlags(cmd.Flags())
	mpCmd.AddNamespaceFlags(cmd.Flags(), true)
	return cmd
}

// Validate always returns true, there are no args to validate
func (c *MarketplaceCmd) Validate(args []string) error {
	return nil
}

// Run retrieves all service classes visible in the current namespace,
// retrieves the plans belonging to those classses, and then displays
// that to the user
func (c *MarketplaceCmd) Run() error {
	opts := servicecatalog.ScopeOptions{
		Namespace: c.Namespace,
		Scope:     servicecatalog.AllScope,
	}
	classes, err := c.App.RetrieveClasses(opts)
	if err != nil {
		return err
	}
	plans := make([][]servicecatalog.Plan, len(classes))
	classPlans, err := c.App.RetrievePlans("", opts)
	if err != nil {
		return err
	}
	for i, class := range classes {
		for _, plan := range classPlans {
			if plan.GetClassID() == class.GetName() {
				plans[i] = append(plans[i], plan)
			}
		}
	}
	output.WriteClassAndPlanDetails(c.Output, classes, plans)
	return nil
}
